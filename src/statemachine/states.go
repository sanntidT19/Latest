package statemachine

import (
	. "chansnstructs"
	"driver"
	"fmt"
	"time"
)

const MAX_SPEED_UP = 300
const MAX_SPEED_DOWN = -300
const SPEED_STOP = 0

const BUTT_PRESS = 1
const BUTT_NPRESS = -1

// Get it confirmed that this is the case
const DIR_UP = 0
const DIR_DOWN = 1

//if we get errors, this bool might be the bad guy

//Functions used when running the elevator, find out better name and add prefix

//The elevator manager
// - Takes in an array, and updates buttons/lights to be cleared/set, comparing with its own private array, only sends to a channel if something is different
// - Has a logic function sending the correct order to the elevator_worker
// - Sends these updates to the mothership: state(struct) updated, order (button struct) served, has/has not command (bool?)

var InStateMChans InternalStateMachineChannels

type InternalStateMachineChannels struct {
	internalButtPressChan chan Order
	orderArrayChan        chan [N_FLOORS][2]bool
	commandOrderChan      chan [N_FLOORS]bool
	ordersCalculatedChan  chan []Order
	orderServedChan       chan bool
	speedChan             chan float64 //This channel is only between the statemachine and its functions
	privateSensorChan     chan int
	currentStateChan      chan State
	goToFloorChan         chan int
	buttonUpdatedChan     chan Order
	toChooseNextOrderChan chan bool
	localCurrentStateChan chan State
	commandLightsChan     chan [N_FLOORS]bool
}

func Internal_state_machine_channels_init() {
	InStateMChans.commandLightsChan = make(chan [N_FLOORS]bool)
	InStateMChans.internalButtPressChan = make(chan Order)
	InStateMChans.toChooseNextOrderChan = make(chan bool)
	InStateMChans.orderArrayChan = make(chan [driver.N_FLOORS][2]bool)
	InStateMChans.commandOrderChan = make(chan [driver.N_FLOORS]bool)
	InStateMChans.ordersCalculatedChan = make(chan []Order)
	InStateMChans.orderServedChan = make(chan bool)
	InStateMChans.speedChan = make(chan float64)
	InStateMChans.privateSensorChan = make(chan int)
	InStateMChans.currentStateChan = make(chan State)
	InStateMChans.localCurrentStateChan = make(chan State)
	InStateMChans.goToFloorChan = make(chan int)
	InStateMChans.buttonUpdatedChan = make(chan Order) //is this correct? the old type Button
}

// /*orderArrayChan chan [][] int <- this will come from above, but not in this test program /* this may need to have a different name -> ,currentStateChan chan State
func Elevator_manager() {
	var currentState State
	var externalArray [N_FLOORS][2]bool
	var internalList [N_FLOORS]bool
	Internal_state_machine_channels_init()
	driver.Elev_init()

	//currentStateChan <- State{driver.Elev_get_direction(), driver.Elev_get_floor_sensor_signal()}

	//buttonSliceChan := Create_button_chan_slice()
	//lightSliceChan := Create_button_chan_slice()

	go Elevator_worker()
	go Light_updater()
	go Choose_next_order()
	go Send_orders_to_worker(&internalList)
	go Button_updater()
	for {
		select {
		case currentState = <-InStateMChans.currentStateChan:
			//fmt.Println("before currentState outwards")
			ExStateMChans.CurrentStateChan <- currentState
			//fmt.Println("before localcurrent")
			InStateMChans.localCurrentStateChan <- currentState
			//fmt.Println("after localcurrent")
		/*
		   case orderServed := <-InStateMChans.orderServedChan:
		   ExStateMChans.OrderServedChan <- orderServed
		   InStateMChans.toChooseNextOrderChan <- orderServed
		*/
		case internalOrder := <-InStateMChans.internalButtPressChan:
			internalList[internalOrder.Floor] = true
			InStateMChans.commandLightsChan <- internalList
			InStateMChans.orderArrayChan <- externalArray
			InStateMChans.commandOrderChan <- internalList
		case externalArray = <-ExStateMChans.SingleExternalListChan:
			InStateMChans.orderArrayChan <- externalArray
			InStateMChans.commandOrderChan <- internalList
		}
	}
}

//Going where it is told to go, based on information on where it is. Gotofloor needs to be buttonized.
func Elevator_worker() {
currentDir := driver.Elev_get_direction()
currentFloor := driver.Elev_get_floor_sensor_signal() //We know that we are in a floor at this point.
// previousFloor := -1 IF WE GET SHIT DONE INCLUDE THIS AND CASES BELOW
orderedFloor := -1
var goToF int
go Motor_control()
go Is_floor_reached()
var hasSentOrderServed bool
// This is where the statemachine is implemented, should it be a select case?
for {
	select {
//Slave could send a new command while the statemachine is serving another command, but it should fix the logic by itself
// New order
		case goToF = <-InStateMChans.goToFloorChan:
			hasSentOrderServed = false
			//fmt.Println("Button pressed sent to elevator worker")
			orderedFloor = goToF
			//fmt.Println("this is the ordered floor: ", orderedFloor)
			//You are in the floor, order served immediatly, maybe this if can be implemented in another case, but its here for now.
			if goToF == driver.Elev_get_floor_sensor_signal() {
				fmt.Println("t|1")
				hasSentOrderServed = true
				InStateMChans.speedChan <- SPEED_STOP
				fmt.Println("t|2")
				InStateMChans.orderServedChan <- true
				fmt.Println("t|3")
				Open_door() //Dont think we want this select loop to do anything else while the door is open. Solve with go open_door() if its not the case
			//You know you are under/above the current floor
			} else if goToF < currentFloor {
				InStateMChans.speedChan <- MAX_SPEED_DOWN
				currentDir = DIR_DOWN
				InStateMChans.currentStateChan <- State{currentDir, currentFloor} //Do this for all? Should we send if its already going down (no state changed then)
				//fmt.Println("Elevator worker has sent state Down")
			} else if goToF > currentFloor {
				InStateMChans.speedChan <- MAX_SPEED_UP
				currentDir = DIR_UP
				InStateMChans.currentStateChan <- State{currentDir, currentFloor}
				//fmt.Println("Elevator worker has sent state Up")
			}
/* //Your last floor was the current floor, but something may have been pulled, so you dont know where you lie relative to it. Cant use direction.
} else if gtf == currentFloor { //this may now go all the time
//Using previousfloor can give you an idea in some cases.
if previousFloor > currentFloor {
InStateMChans.speedChan <- MAX_SPEED_UP
currentDir = DIR_UP
InStateMChans.currentStateChan <- State{currentDir, currentFloor}
} else if previousFloor < currentFloor { //This will also be the case if prevFloor is undefined (-1) They can never be the same.
InStateMChans.speedChan <- MAX_SPEED_DOWN
currentDir = DIR_DOWN
InStateMChans.currentStateChan <- State{currentDir, currentFloor}
}*/

//New floor is reached and therefore shit is updated
		case cf := <-InStateMChans.privateSensorChan:
//previousFloor = currentFloor
			currentFloor = cf
			InStateMChans.currentStateChan <- State{currentDir, currentFloor}
		//	fmt.Println("ordered floor:", orderedFloor, "should be equal to : ", currentFloor)
			if orderedFloor == currentFloor && !hasSentOrderServed {
				//fmt.Println("reached current floor")
				hasSentOrderServed = true
				InStateMChans.speedChan <- SPEED_STOP
				InStateMChans.orderServedChan <- true
				Open_door()

			}
		}	
	}

}

func Send_orders_to_worker(internalList *[N_FLOORS]bool) {
	var currentOrderList []Order
	var currentOrderIter int
	for {
		select {
		//New orders sorted, picked. scrap the old one
			case currentOrderList = <-InStateMChans.ordersCalculatedChan:
				fmt.Println(currentOrderList)
				currentOrderIter = 0
				if newOrder := currentOrderList[currentOrderIter]; newOrder.TurnOn { //All unsetted orders have turnOn = false by default {
					InStateMChans.goToFloorChan <- newOrder.Floor
				}
				//fmt.Println("Currentorderlist: ",currentOrderList)
			case <-InStateMChans.orderServedChan:
				orderServed := currentOrderList[currentOrderIter]
				if orderServed.ButtonType != driver.COMMAND {
					ExStateMChans.ButtonPressedChan <- Order{orderServed.Floor, orderServed.ButtonType, false}
				}
				if orderServed.ButtonType == driver.COMMAND {
					internalList[orderServed.Floor] = false
					InStateMChans.commandLightsChan <- *internalList
				}
				currentOrderIter++
				if newOrder := currentOrderList[currentOrderIter]; newOrder.TurnOn { //All unsetted orders have turnOn = false by default {
					InStateMChans.goToFloorChan <- newOrder.Floor
				}
		}
	}
}

//This function has control over the orderlist and speaks directly to the worker.
/*func Send_orders_to_worker(internalList *[N_FLOORS]bool) {
	var currentOrderList []Order
	var currentOrderIter int
	var orderisServedandready bool
	go func(){
		for{
			<-InStateMChans.orderServedChan
			fmt.Println("order served trigger 1")
			orderServed := currentOrderList[currentOrderIter]
			if orderServed.ButtonType != driver.COMMAND {
				//fmt.Println("IAMSENDINGBUTTONPRESSEDCHANUPWARDS")
				ExStateMChans.ButtonPressedChan <- Order{orderServed.Floor, orderServed.ButtonType, false}
			}
			if orderServed.ButtonType == driver.COMMAND {
				internalList[orderServed.Floor] = false
				InStateMChans.commandLightsChan <- *internalList
			}
			currentOrderIter++
			orderisServedandready = true
		}
	}()
	go func(){
		for{
			if orderisServedandready{
				if newOrder := currentOrderList[currentOrderIter]; newOrder.TurnOn { //All unsetted orders have turnOn = false by default {
						if goToF == driver.Elev_get_floor_sensor_signal() {
						fmt.Println("t|1")
						hasSentOrderServed = true
					InStateMChans.speedChan <- SPEED_STOP
					fmt.Println("t|2")
					InStateMChans.orderServedChan <- true
					fmt.Println("t|3")
					Open_door()
					InStateMChans.goToFloorChan <- newOrder.Floor
					orderisServedandready = false
				}

			}
			time.Sleep(25*time.Millisecond)
		}
	}()

	for {
		select {
		//New orders sorted, picked. scrap the old one
		case currentOrderList = <-InStateMChans.ordersCalculatedChan:
			fmt.Println("orders calcultated)"
			//fmt.Println(currentOrderList)
			currentOrderIter = 0
			//fmt.Println("Currentorderlist: ",currentOrderList)
			if newOrder := currentOrderList[currentOrderIter]; newOrder.TurnOn{
				//fmt.Println("YOLOMCSWAGGERSON1")
				InStateMChans.goToFloorChan <- currentOrderList[currentOrderIter].Floor
				fmt.Println("go to floor chan 3")
				//fmt.Println("YOLOMCSWAGGERSON2")
			}/*
		case <-InStateMChans.orderServedChan:
			orderServed := currentOrderList[currentOrderIter]
			if orderServed.ButtonType != driver.COMMAND {
				fmt.Println("IAMSENDINGBUTTONPRESSEDCHANUPWARDS")
				ExStateMChans.ButtonPressedChan <- Order{orderServed.Floor, orderServed.ButtonType, false}
			}
			if orderServed.ButtonType == driver.COMMAND {
				internalList[orderServed.Floor] = false
				InStateMChans.commandLightsChan <- *internalList
			}
			currentOrderIter++
			if newOrder := currentOrderList[currentOrderIter]; newOrder.TurnOn { //All unsetted orders have turnOn = false by default {
				InStateMChans.goToFloorChan <- newOrder.Floor

			}
		}
	}
}*/

//Swapping : x,y = y,x
// logic here, no problemo, motherfucker
func Choose_next_order() {
	//fmt.Println("choose next order")
	var currentFloor, currentDir int
	var dirIter int //Deciding where to iterate first
	var orderArray [driver.N_FLOORS][2]bool
	var commandList [driver.N_FLOORS]bool
	var firstPriority, secondPriority int
	/*
	   go func(){
	   for{
	   select{
	   case orderArray = <-orderArrayChan:
	   fmt.Println("I am also in this choosenextorder-case")
	   incomingUpdateChan <- true
	   case commandList = <-commandOrderChan:
	   fmt.Println("hope im not here")
	   incomingUpdateChan <- true
	   }
	   }

	   }()
	*/
	for {
		select {
		case currentState := <-InStateMChans.localCurrentStateChan:
			currentFloor = currentState.CurrentFloor
			currentDir = currentState.Direction
		case orderArray = <-InStateMChans.orderArrayChan:
			commandList = <-InStateMChans.commandOrderChan
			//fmt.Println("I CALCULATE FOR YOU BOSS YOLO")
			//Makin a slice of sorted orders and sending it to the elevator manager.
			resultOrderSlice := make([]Order, driver.N_FLOORS*driver.N_BUTTONS) //This needs to be printed
			resultIter := 0
			dirIter = 1
			firstPriority = driver.UP
			secondPriority = driver.DOWN
			if currentDir == DIR_DOWN {
				dirIter = -1
				firstPriority = driver.DOWN
				secondPriority = driver.UP
			}
			//If we are above/below the floor we need to prioritize that as one of the less attractive ones
			if driver.Elev_get_floor_sensor_signal() != currentFloor { //Maybe assign a value here
				currentFloor += dirIter //Setting the startingfloor to iterate from
			}
			i := currentFloor
			// Iterating in the most desirable direction
			for i < driver.N_FLOORS && i >= 0 {
				if isCommandOrder := commandList[i]; isCommandOrder {
					resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
					resultIter++
				}
				if isOrder := orderArray[i][firstPriority]; isOrder {
					resultOrderSlice[resultIter] = Order{i, firstPriority, true}
					resultIter++
				}
				i += dirIter
			}
			i -= dirIter
			//Now going from top/bottom and checking the orders in the other direction, as well as commands not yet checked
			dirIter = dirIter * -1
			firstPriority, secondPriority = secondPriority, firstPriority
			canCheckCommand := false
			for i < driver.N_FLOORS && i >= 0 {
				if canCheckCommand {
					if isCommandOrder := commandList[i]; isCommandOrder {
						resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
						resultIter++
					}
				}
				if i == currentFloor {
					canCheckCommand = true
				}
				if isOrder := orderArray[i][firstPriority]; isOrder {
					resultOrderSlice[resultIter] = Order{i, firstPriority, true}
					resultIter++
				}
				i += dirIter
			}
			i -= dirIter

			//Lowest priority: checking from bottom/top to current floor if there are any bastards wanting an elevated experience all commands have been checked
			dirIter = dirIter * -1
			firstPriority, secondPriority = secondPriority, firstPriority
			for i != currentFloor {
				if isOrder := orderArray[i][firstPriority]; isOrder {
					resultOrderSlice[resultIter] = Order{i, firstPriority, true}
					resultIter++
				}
				i += dirIter
			}
			InStateMChans.ordersCalculatedChan <- resultOrderSlice
		}
	}
}

//This one creates the basic button slice for our friends

func Create_button_chan_slice() []chan Order { //A little unsure of this

	chanSlice := make([]chan Order, N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		chanSlice[i] = make(chan Order)
	}
	return chanSlice
}

//Slave should only send to buttonUpdated if something comes from above, ie not from button_updater
func Button_updater() { //Sending the struct a level up, to the state machine setting and turning off lights.

	buttonMatrix := make([][]int, N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		buttonMatrix[i] = make([]int, N_BUTTONS) //Golang creates a slice of zeros by default
	}

	buttonMatrix[N_FLOORS-1][UP] = -1
	buttonMatrix[0][DOWN] = -1
	go func() {
		for {
			order := <-InStateMChans.buttonUpdatedChan //Word from above that some button is updated
			
			if order.TurnOn {
				buttonMatrix[order.Floor][order.ButtonType] = 1
			} else {
				buttonMatrix[order.Floor][order.ButtonType] = 0

			}
		}
	}()

	//fmt.Print(buttonMatrix)
	//Continious checking of buttons, buttonChan is a buffered channel who can fit N_FLOORS*N_BUTTONS elements
	for {
		time.Sleep(time.Millisecond * 40) //Need a proper time to wait.
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if buttonVar := driver.Elev_get_button_signal(i, j); buttonVar != buttonMatrix[i][j] { //Sending the struct if its pushed and hasnt been sent already
					//fmt.Println("Here is drivers version of button pressed: ", buttonVar)
					//fmt.Println("floor and button:", i, j)
					if buttonVar == 1 && j != driver.COMMAND {
						ExStateMChans.ButtonPressedChan <- Order{i, j, true} // YO! Need to make this one sexier. Maybe one channel for each button
						buttonMatrix[i][j] = 1
						//Confirm press to avoid spamming
					} else if buttonVar == 1 && j == driver.COMMAND {
						buttonMatrix[i][j] = 1 //Confirm press to avoid spamming

						InStateMChans.internalButtPressChan <- Order{i, j, true}
						fmt.Println("Internal Button Pressed")

					}
				}
			}
		}
	}

}

//Try having a dedicated channel for each floor. Light updater will receive a floor command, and set the light on or off
//This will receive commands from two different holds, and only one will be served. I dont think this will be a problem

//If we get time: see how we can make this more dynamic. Yhis also turns off if the bool is false.
func Light_updater() {
	for {
		select {
		case lightArray := <-ExStateMChans.LightChan:
			for i := 0; i < N_FLOORS; i++ {
				for j := 0; j < N_BUTTONS-1; j++ {
					
					driver.Elev_set_button_lamp(i, j, lightArray[i][j])
					InStateMChans.buttonUpdatedChan <- Order{i, j, lightArray[i][j]}

				}
			}
		case commandLights := <-InStateMChans.commandLightsChan:
			for i := 0; i < N_FLOORS; i++ {
				fmt.Println("commandLights ")
				driver.Elev_set_button_lamp(i, driver.COMMAND, commandLights[i])
				InStateMChans.buttonUpdatedChan <- Order{i, driver.COMMAND, commandLights[i]}
			}
		}
	}
}

func Motor_control() { //I think speedchan should not be buffered
	for {
		speedVal := <-InStateMChans.speedChan
		/*if speedVal == 0 {
			driver.Elev_stop_elevator()
		} else {*/
		driver.Elev_set_speed(speedVal)
		//}
	}
}

// Gets sensor signal and tells which floor is the current

func Is_floor_reached() {
	var previousFloor int = -1
	for {
		if currentFloor := driver.Elev_get_floor_sensor_signal(); currentFloor != -1 && currentFloor != previousFloor {
			fmt.Println(currentFloor)
			driver.Elev_set_floor_indicator(currentFloor)
			previousFloor = currentFloor
			InStateMChans.privateSensorChan <- currentFloor

		}
		time.Sleep(time.Millisecond * 25)
	}
}

func Open_door() {
	driver.Elev_set_door_open_lamp(true)
	time.Sleep(time.Second * 3)
	driver.Elev_set_door_open_lamp(false)
}
