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
//	- Takes in an array, and updates buttons/lights to be cleared/set, comparing with its own private array, only sends to a channel if something is different
//	- Has a logic function sending the correct order to the elevator_worker
//	- Sends these updates to the mothership: state(struct) updated, order (button struct) served, has/has not command (bool?)

var InStateMChans InternalStateMachineChannels

type InternalStateMachineChannels struct {
	internalButtPressChan chan Order
	orderArrayChan        chan [N_FLOORS][2]int
	commandOrderChan      chan [N_FLOORS]int
	ordersCalculatedChan  chan []Order
	orderServedChan       chan bool
	speedChan             chan float64 //This channel is only between the statemachine and its functions
	privateSensorChan     chan int
	currentStateChan      chan State
	goToFloorChan         chan int
	buttonUpdatedChan     chan Order
	toChooseNextOrderChan chan bool

}

func Internal_state_machine_channels_init() {
	InStateMChans.internalButtPressChan = make(chan Order)
	InStateMChans.toChooseNextOrderChan = make(chan bool)
	InStateMChans.orderArrayChan = make(chan [driver.N_FLOORS][2]int)
	InStateMChans.commandOrderChan = make(chan [driver.N_FLOORS]int)
	InStateMChans.ordersCalculatedChan = make(chan []Order)
	InStateMChans.orderServedChan = make(chan bool)
	InStateMChans.speedChan = make(chan float64)
	InStateMChans.privateSensorChan = make(chan int)
	InStateMChans.currentStateChan = make(chan State)
	InStateMChans.goToFloorChan = make(chan int)
	InStateMChans.buttonUpdatedChan = make(chan Order) //is this correct? the old type Button
}

// /*orderArrayChan chan [][] int   <- this will come from above, but not in this test program  /* this may need to have a different name ->  ,currentStateChan chan State
func Elevator_manager() {
	var crrentState State
	var externalArray [N_FLOORS][2] bool
	var internalList [N_FLOORS] bool
	driver.Elev_init()

	//currentStateChan <- State{driver.Elev_get_direction(), driver.Elev_get_floor_sensor_signal()}

	//buttonSliceChan := Create_button_chan_slice()
	//lightSliceChan := Create_button_chan_slice()
	go Executer()
	fmt.Println("1")
	go Elevator_worker()
	fmt.Println("ew")
	go Light_updater()
	fmt.Println("lu")
	go Choose_next_order()
	fmt.Println("cno")
	go Send_orders_to_worker()
	fmt.Println("sotw")
	go Button_updater()
	fmt.Println("2")
	for{
		select{
		case  currentState = <- InStateMChans.currentStateChan:
			ExStateMChans.CurrentStateChan <- currentState
			InStateMChans.LocalCurrentStateChan <- currentState NEW CHANNEL
		case orderServed := <- InStateMChans.orderServedChan:
			ExStateMChans.OrderServedChan <-orderServed
			InStateMChans.toChooseNextOrderChan <-orderServed   NEW CHANNEL
		case internalOrder := <- InStateMChans.internalButtPressChan:
			internalList[internalOrder.floor]  = true
			InStateMchans.commandLightsChan <- interalList
			InStateMChans.orderArrayChan<-externalArray
			InStateMchans.commandOrderChan <- internalList
		case externalArray = <- ExStateMChans.SingleExternalListChan:
			InStateMChans.orderArrayChan <- externalArray
			InStateMChans.commandOrderChan <- internalList
		}
	}

}

//Going where it is told to go, based on information on where it is.   Gotofloor needs to be buttonized.
func Elevator_worker() {
	fmt.Println("elevator worker started")
	currentDir := driver.Elev_get_direction()
	currentFloor := driver.Elev_get_floor_sensor_signal() //We know that we are in a floor at this point.
	previousFloor := -1
	orderedFloor := -1
	go Motor_control()
	//go Is_floor_reached()
	// This is where the statemachine is implemented, should it be a select case?
	for {
		fmt.Println("t22")
		select {
		//Slave could send a new command while the statemachine is serving another command, but it should fix the logic by itself
		// New order
		case goToF := <-InStateMChans.goToFloorChan:
			orderedFloor = gotoF
			//You are in the floor, order served immediatly, maybe this if can be implemented in another case, but its here for now.
			if goToF == driver.Elev_get_floor_sensor_signal() {
				InStateMChans.speedChan <- SPEED_STOP
				InStateMChans.orderServedChan <- true
				Open_door() //Dont think we want this select loop to do anything else while the door is open. Solve with go open_door() if its not the case
				//You know you are under/above the current floor
			} else if goToF < currentFloor {
				InStateMChans.speedChan <- MAX_SPEED_DOWN
				currentDir = DIR_DOWN
				InStateMChans.currentStateChan <- State{currentDir, currentFloor} //Do this for all? Should we send if its already going down (no state changed then)

			} else if goToF > currentFloor {
				InStateMChans.speedChan <- MAX_SPEED_UP
				currentDir = DIR_UP
				InStateMChans.currentStateChan <- State{currentDir, currentFloor}
			}
			/*	//Your last floor was the current floor, but something may have been pulled, so you dont know where you lie relative to it. Cant use direction.
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
			previousFloor = currentFloor
			currentFloor = cf
			InStateMChans.currentStateChan <- State{currentDir, currentFloor}
			if orderedFloor == currentFloor {
				InStateMChans.speedChan <- SPEED_STOP
				InStateMChans.orderServedChan <- true
				Open_door()
			}
		}
	}

}

//This function has control over the orderlist and speaks directly to the worker.
func Send_orders_to_worker() {
	var currentOrderList []Order
	var currentOrderIter int
	for {
		select {
		//New orders sorted, picked. scrap the old one
		case currentOrderList = <-InStateMChans.ordersCalculatedChan:
			currentOrderIter = 0
			//fmt.Println("Currentorderlist: ",currentOrderList)
			InStateMChans.orderToManagerChan <- currentOrderList[currentOrderIter]
			currentOrderIter++
		case <-InStateMChans.toChooseNextOrderChan:
			if currentOrderList[currentOrderIter].TurnOn { //All unsetted orders have turnOn = false by default {
				InStateMChans.orderToManagerChan <- currentOrderList[currentOrderIter]
				currentOrderIter++
			}
		}
	}
}

//Swapping : x,y = y,x
// logic here, no problemo, motherfucker
func Choose_next_order() {
	fmt.Println("choose next order")
	var currentFloor, currentDir int
	var dirIter int //Deciding where to iterate first
	var orderArray [driver.N_FLOORS][2]int
	var commandList [driver.N_FLOORS]int
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
			//fmt.Println("I CALCULATE FOR YOU BOSS                                   YOLO")
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
				if isCommandOrder := commandList[i]; isCommandOrder == 1 {
					resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
					resultIter++
				}
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
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
					if isCommandOrder := commandList[i]; isCommandOrder == 1 {
						resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
						resultIter++
					}
				}
				if i == currentFloor {
					canCheckCommand = true
				}
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
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
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
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

	fmt.Println("making slice for you, sir")
	chanSlice := make([]chan Order, N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		chanSlice[i] = make(chan Order)
	}
	return chanSlice
}

//Slave should only send to buttonUpdated if something comes from above, ie not from button_updater
func Button_updater() { //Sending the struct a level up, to the state machine setting and turning off lights.
	fmt.Println("button updater")
	buttonMatrix := make([][]int, N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		buttonMatrix[i] = make([]int, N_BUTTONS) //Golang creates a slice of zeros by default
	}

	buttonMatrix[N_FLOORS-1][UP] = -1
	buttonMatrix[0][DOWN] = -1

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

						ExStateMChans.ButtonPressed <- Order{i, j, true} //   YO!   Need to make this one sexier. Maybe one channel for each button

						buttonMatrix[i][j] = 1
						//Confirm press to avoid spamming
					} else if buttonVar == 1 && j == driver.COMMAND {

						buttonMatrix[i][j] = 1 //Confirm press to avoid spamming

						InStateMChans.internalButtPressChan <- Order{i, j, true}
					}
				}
			}
		}
	}
	go func() {
		for {
			order := <-InStateMChans.buttonUpdatedChan //Word from above that some button is updated
			if butt.TurnOn {
				buttonMatrix[order.Floor][order.ButtonType] = 1
			} else {
				buttonMatrix[order.Floor][order.ButtonType] = 0

			}
		}
	}()
}

//Try having a dedicated channel for each floor. Light updater will receive a floor command, and set the light on or off
//This will receive commands from two different holds, and only one will be served. I dont think this will be a problem

//If we get time: see how we can make this more dynamic. Yhis also turns off if the bool is false.
func Light_updater() {
	fmt.Println("light updater")
	
	for {
		select {
			case lightArray := <- ExCommChans.LightChan:
			for i := 0; i < N_FLOORS{
				for j :=0 ; i < N_BUTTONS -1{
					driver.Elev_set_button_lamp(i, j, lightArray[i][j])
					InStateMChans.orderUpdatedChan <- Order{i,j,lightArray[i][j]} NEW CHANNEL
				}
			}
			case commandLights := <- InStateMChans.commandLightsChan:
			for i := 0;  i < N_FLOORS{
				driver.Elev_set_button_lamp(i,COMMAND,commandLights[i])
				InStateMChans.orderUpdatedChan <- Order{i,COMMAND,commandLights[i]} 
			}
		}
	}
}

func Motor_control() { //I think speedchan should not be buffered
	fmt.Println("motrocontol")
	for {
		speedVal := <-InStateMChans.speedChan
		if speedVal == 0 {
			driver.Elev_stop_elevator()
		} else {
			driver.Elev_set_speed(speedVal)
		}
	}
}

// Gets sensor signal and tells which floor is the current

func Is_floor_reached() {
	var previousFloor int = -1
	for {
		if currentFloor := driver.Elev_get_floor_sensor_signal(); currentFloor != -1 && currentFloor != previousFloor {
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

/*

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

var InStateMChans InternalStateMachineChannels

//if we get errors, this bool might be the bad guy

//Functions used when running the elevator, find out better name and add prefix

//The elevator manager
//	- Takes in an array, and updates buttons/lights to be cleared/set, comparing with its own private array, only sends to a channel if something is different
//	- Has a logic function sending the correct order to the elevator_worker
//	- Sends these updates to the mothership: state(struct) updated, order (button struct) served, has/has not command (bool?)

type InternalStateMachineChannels struct {
}

func Internal_state_machine_channels_init() {

}

// /*orderArrayChan chan [][] int   <- this will come from above, but not in this test program   this may need to have a different name ->  ,currentStateChan chan State

func Elevator_manager() {
	driver.Elev_init()

	//This must be done more smoothly later
	currentStateChan := make(chan State)
	//currentStateChan <- State{driver.Elev_get_direction(), driver.Elev_get_floor_sensor_signal()}

	var orderArray [driver.N_FLOORS][2]int // Up and down goes here
	var commandOrderList [driver.N_FLOORS]int
	var managersCurrentOrder Order
	externalButtPressChan := make(chan Order)
	internalButtPressChan := make(chan Order)
	buttUpdatedChan := make(chan Order)
	orderArrayChan := make(chan [driver.N_FLOORS][2]int)
	setLightChan := make(chan Order)
	commandOrderChan := make(chan [driver.N_FLOORS]int)
	ordersCalculatedChan := make(chan []Order)
	orderToWorkerChan := make(chan int)
	orderServedChan := make(chan bool)
	orderToManagerChan := make(chan Order)
	tempOrderServedChan := make(chan bool)

	//buttonSliceChan := Create_button_chan_slice()
	//lightSliceChan := Create_button_chan_slice()

	go Button_updater(externalButtPressChan, internalButtPressChan, buttUpdatedChan)
	go Light_updater(setLightChan, buttUpdatedChan)

	go Choose_next_order(orderArrayChan, commandOrderChan, currentStateChan, ordersCalculatedChan)

	go Elevator_worker(orderToWorkerChan, currentStateChan, orderServedChan)
	go Send_orders_to_worker(orderToManagerChan, ordersCalculatedChan, tempOrderServedChan)

	for {
		select {
		//If something is pressed, the channels are updated
		case orderButt := <-internalButtPressChan:

			commandOrderList[orderButt.Floor] = 1
			//fmt.Println("Command order pressed: ", commandOrderList)

			orderArrayChan <- orderArray
			commandOrderChan <- commandOrderList

			setLightChan <- orderButt
		case orderButt := <-externalButtPressChan:

			orderArray[orderButt.Floor][orderButt.ButtonType] = 1
			//fmt.Println("External order pressed: ", orderArray)

			orderArrayChan <- orderArray
			commandOrderChan <- commandOrderList

			setLightChan <- orderButt
		case managersCurrentOrder = <-orderToManagerChan: //Just so the manager can keep up with the current order
			//fmt.Println("managersCurrentOrder:", managersCurrentOrder)
			orderToWorkerChan <- managersCurrentOrder.Floor
		case <-orderServedChan:
			//fmt.Println("Order Served!")
			//fmt.Println("managersCurrentOrder:", managersCurrentOrder)
			if managersCurrentOrder.ButtonType == driver.COMMAND {
				commandOrderList[managersCurrentOrder.Floor] = 0
			} else {
				orderArray[managersCurrentOrder.Floor][managersCurrentOrder.ButtonType] = 0
			}
			//fmt.Println("CommandOrderlist: ", commandOrderList)
			//fmt.Println("External order array: ", orderArray)
			setLightChan <- Order{managersCurrentOrder.Floor, managersCurrentOrder.ButtonType, false}
			time.Sleep(time.Second * 1)
			tempOrderServedChan <- true
		}
	}

}

//Going where it is told to go, based on information on where it is.   Gotofloor needs to be buttonized.
func Elevator_worker(goToFloorChan chan int, currentStateChan chan State, orderServedChan chan bool) {
	speedChan := make(chan float64) //This channel is only between the statemachine and its functions
	privateSensorChan := make(chan int)
	currentDir := driver.Elev_get_direction()
	currentFloor := driver.Elev_get_floor_sensor_signal() //We know that we are in a floor at this point.
	previousFloor := -1
	orderedFloor := -1
	go Motor_control(speedChan)
	go Is_floor_reached(privateSensorChan)
	// This is where the statemachine is implemented, should it be a select case?
	for {
		select {
		//Slave could send a new command while the statemachine is serving another command, but it should fix the logic by itself
		// New order
		case gtf := <-goToFloorChan:
			orderedFloor = gtf
			//You are in the floor, order served immediatly, maybe this if can be implemented in another case, but its here for now.
			if gtf == driver.Elev_get_floor_sensor_signal() {
				speedChan <- SPEED_STOP
				orderServedChan <- true
				Open_door() //Dont think we want this select loop to do anything else while the door is open. Solve with go open_door() if its not the case
				//You know you are under/above the current floor
			} else if gtf < currentFloor {
				speedChan <- MAX_SPEED_DOWN
				currentDir = DIR_DOWN
				currentStateChan <- State{currentDir, currentFloor} //Do this for all? Should we send if its already going down (no state changed then)

			} else if gtf > currentFloor {
				speedChan <- MAX_SPEED_UP
				currentDir = DIR_UP
				currentStateChan <- State{currentDir, currentFloor}
				//Your last floor was the current floor, but something may have been pulled, so you dont know where you lie relative to it. Cant use direction.
			} else if gtf == currentFloor { //this may now go all the time
				//Using previousfloor can give you an idea in some cases.
				if previousFloor > currentFloor {
					speedChan <- MAX_SPEED_UP
					currentDir = DIR_UP
					currentStateChan <- State{currentDir, currentFloor}
				} else if previousFloor < currentFloor { //This will also be the case if prevFloor is undefined (-1) They can never be the same.
					speedChan <- MAX_SPEED_DOWN
					currentDir = DIR_DOWN
					currentStateChan <- State{currentDir, currentFloor}
				}
			}
		//New floor is reached and therefore shit is updated
		case cf := <-privateSensorChan:
			previousFloor = currentFloor
			currentFloor = cf
			currentStateChan <- State{currentDir, currentFloor}
			if orderedFloor == currentFloor {
				speedChan <- SPEED_STOP
				orderServedChan <- true
				Open_door()
			}
		}
	}
}

//This function has control over the orderlist and speaks directly to the worker.
func Send_orders_to_worker(orderToManagerChan chan Order, ordersCalculatedChan chan []Order, tempOrderServedChan chan bool) {
	var currentOrderList []Order
	var currentOrderIter int
	for {
		select {
		//New orders sorted, picked. scrap the old one
		case currentOrderList = <-ordersCalculatedChan:
			currentOrderIter = 0
			//fmt.Println("Currentorderlist: ",currentOrderList)
			orderToManagerChan <- currentOrderList[currentOrderIter]
			currentOrderIter++
		case <-tempOrderServedChan:
			if currentOrderList[currentOrderIter].TurnOn { //All unsetted orders have turnOn = false by default {
				orderToManagerChan <- currentOrderList[currentOrderIter]
				currentOrderIter++
			}
		}
	}
}

//Swapping : x,y = y,x
// logic here, no problemo, motherfucker
func Choose_next_order(orderArrayChan chan [driver.N_FLOORS][2]int, commandOrderChan chan [driver.N_FLOORS]int, currentStateChan chan State, ordersCalculatedChan chan []Order) {
	var currentFloor, currentDir int
	var dirIter int //Deciding where to iterate first
	var orderArray [driver.N_FLOORS][2]int
	var commandList [driver.N_FLOORS]int
	var firstPriority, secondPriority int

	for {
		select {
		case currentState := <-currentStateChan:
			currentFloor = currentState.CurrentFloor
			currentDir = currentState.Direction
		case orderArray = <-orderArrayChan:
			commandList = <-commandOrderChan
			//fmt.Println("I CALCULATE FOR YOU BOSS                                   YOLO")
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
				if isCommandOrder := commandList[i]; isCommandOrder == 1 {
					resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
					resultIter++
				}
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
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
					if isCommandOrder := commandList[i]; isCommandOrder == 1 {
						resultOrderSlice[resultIter] = Order{i, driver.COMMAND, true}
						resultIter++
					}
				}
				if i == currentFloor {
					canCheckCommand = true
				}
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
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
				if isOrder := orderArray[i][firstPriority]; isOrder == 1 {
					resultOrderSlice[resultIter] = Order{i, firstPriority, true}
					resultIter++
				}
				i += dirIter
			}
			ordersCalculatedChan <- resultOrderSlice
		}
	}
}

//This one creates the basic button slice for our friends

func Create_button_chan_slice() []chan Order { //A little unsure of this
	//fmt.Println("making slice for you, sir")
	chanSlice := make([]chan Order, driver.N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		chanSlice[i] = make(chan Order)
	}
	return chanSlice
}

/*
func old_Button_updater(buttonSlice []chan Button) { //Sending the struct a level up, to the state machine setting and turning off lights.
	var buttonMatrix [][]int
	buttonMatrix = make([][]int, driver.N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		buttonMatrix[i] = make([]int, driver.N_BUTTONS) //Golang creates a slice of zeros by default
	}
	fmt.Print(buttonMatrix)
	go func
	//Continious checking of buttons, buttonChan is a buffered channel who can fit N_FLOORS*N_BUTTONS elements
	for {
		time.Sleep(time.Millisecond * 20) //Need a proper time to wait.
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if buttonVar := driver.Elev_get_button_signal(i, j); buttonVar != buttonMatrix[i][j] { //Sending the struct if its pushed and hasnt been sent already
					fmt.Println("Here is drivers version of button pressed: ", buttonVar)
					fmt.Println("floor and button:", i, j)
					if buttonVar == 1 {

						buttonSlice[i] <- Button{i, j, true} //This might give an error
					}
					buttonMatrix[i][j] = buttonVar
				}
			}
		}
	}
}

//Slave should only send to buttonUpdated if something comes from above, ie not from button_updater
func Button_updater(externalButtPressChan chan Order, internalButtPressChan chan Order, buttonUpdatedChan chan Order) { //Sending the struct a level up, to the state machine setting and turning off lights.
	buttonMatrix := make([][]int, driver.N_FLOORS)
	for i := 0; i < driver.N_FLOORS; i++ {
		buttonMatrix[i] = make([]int, driver.N_BUTTONS) //Golang creates a slice of zeros by default
	}

	buttonMatrix[driver.N_FLOORS-1][driver.UP] = -1
	buttonMatrix[0][driver.DOWN] = -1

	//fmt.Print(buttonMatrix)
	go func() {
		for {
			butt := <-buttonUpdatedChan //Word from above that some button is updated
			if butt.TurnOn {
				buttonMatrix[butt.Floor][butt.ButtonType] = 1
			} else {
				buttonMatrix[butt.Floor][butt.ButtonType] = 0

			}
		}
	}()
	//Continious checking of buttons, buttonChan is a buffered channel who can fit N_FLOORS*N_BUTTONS elements
	for {
		time.Sleep(time.Millisecond * 40) //Need a proper time to wait.
		for i := 0; i < driver.N_FLOORS; i++ {
			for j := 0; j < driver.N_BUTTONS; j++ {
				if buttonVar := driver.Elev_get_button_signal(i, j); buttonVar != buttonMatrix[i][j] { //Sending the struct if its pushed and hasnt been sent already
					//fmt.Println("Here is drivers version of button pressed: ", buttonVar)
					//fmt.Println("floor and button:", i, j)
					if buttonVar == 1 && j != driver.COMMAND {
						fmt.Println("Button has been pushed")
						externalButtPressChan <- Order{i, j, true} //   YO!   Need to make this one sexier. Maybe one channel for each button
						buttonMatrix[i][j] = 1                     //Confirm press to avoid spamming
					} else if buttonVar == 1 && j == driver.COMMAND {
						fmt.Println("Button has been pushed")
						buttonMatrix[i][j] = 1 //Confirm press to avoid spamming

						internalButtPressChan <- Order{i, j, true}
					}
				}
			}
		}
	}
}

//Try having a dedicated channel for each floor. Light updater will receive a floor command, and set the light on or off
//This will receive commands from two different holds, and only one will be served. I dont think this will be a problem

//If we get time: see how we can make this more dynamic. Yhis also turns off if the bool is false.
func Light_updater(setLightChan chan Order, buttUpdatedChan chan Order) {
	for {
		butt := <-setLightChan
		_ = driver.Elev_set_button_lamp(butt.Floor, butt.ButtonType, butt.TurnOn)
		buttUpdatedChan <- butt
	}
}

func Motor_control(speedChan chan float64) { //I think speedchan should not be buffered
	for {
		speedVal := <-speedChan
		if speedVal == 0 {
			driver.Elev_stop_elevator()
		} else {
			driver.Elev_set_speed(speedVal)
		}
	}
}

// Gets sensor signal and tells which floor is the current
func Is_floor_reached(sensorChan chan int) {
	var previousFloor int = -1
	for {
		if currentFloor := driver.Elev_get_floor_sensor_signal(); currentFloor != -1 && currentFloor != previousFloor {
			driver.Elev_set_floor_indicator(currentFloor)
			previousFloor = currentFloor
			sensorChan <- currentFloor

		}
		time.Sleep(time.Millisecond * 250)
	}
}

func Open_door() {
	driver.Elev_set_door_open_lamp(true)
	time.Sleep(time.Second * 3)
	driver.Elev_set_door_open_lamp(false)
}
*/
