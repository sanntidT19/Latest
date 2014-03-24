package driver

/*
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
#include "io.c"
*/
import (
	"C"
	"errors"
	"fmt"
	"math"
	"time"
)

// Number of signals and lamps on a per-floor basis (excl sensor)
const N_BUTTONS = 3
const N_FLOORS = 4 //is this required here? should this be all caps?

//constants making it easier to read the code, will prob be used
const UP = 0
const DOWN = 1
const COMMAND = 2

var lamp_channel_matrix [][]int
var button_channel_matrix [][]int

func Elev_make_std_l_matrix() [][]int { //stupid name? we need to agree on a stanard name convention
	std_matrix := make([][]int, 4, 8)
	std_matrix[0] = make([]int, N_BUTTONS)
	std_matrix[0] = []int{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1}
	std_matrix[1] = make([]int, N_BUTTONS)
	std_matrix[1] = []int{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2}
	std_matrix[2] = make([]int, N_BUTTONS)
	std_matrix[2] = []int{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3}
	std_matrix[3] = make([]int, N_BUTTONS)
	std_matrix[3] = []int{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4}
	return std_matrix
}

func Elev_make_std_b_matrix() [][]int {
	std_matrix := make([][]int, 4, 8) //Find out if the capacity here is valid
	std_matrix[0] = make([]int, N_BUTTONS)
	std_matrix[0] = []int{FLOOR_UP1, FLOOR_DOWN1, FLOOR_COMMAND1}
	std_matrix[1] = make([]int, N_BUTTONS)
	std_matrix[1] = []int{FLOOR_UP2, FLOOR_DOWN2, FLOOR_COMMAND2}
	std_matrix[2] = make([]int, N_BUTTONS)
	std_matrix[2] = []int{FLOOR_UP3, FLOOR_DOWN3, FLOOR_COMMAND3}
	std_matrix[3] = make([]int, N_BUTTONS)
	std_matrix[3] = []int{FLOOR_UP4, FLOOR_DOWN4, FLOOR_COMMAND4}
	return std_matrix
}

//Possibility for extension of floors, but not used in this project. capacity of the matrix is extended if necessary
//It is assumed that the added "floor" is either on the top or the bottom of the elevator-shaft.

/*func elev_extend_matrix(matrix [][] int,light_up int ,light_down int,light_command int, int floor) ([][] int, int) {
	if len(matrix) == cap(matrix) {
		temp_matrix := make([][]int, len(matrix), (cap(matrix)+1)*2)
		copy(temp_matrix,matrix)
		matrix := temp_matrix
		}
	}
	return new_matrix, floors + 1 // is N_FLOORS available here?
}
*/

//We havent used goroutines nor channels here, since this is a one time event for the life of an elevator
func Elev_init() error {
	if !io_init() {
		return errors.New("IO initialization failed")
	}
	lamp_channel_matrix = Elev_make_std_l_matrix()
	button_channel_matrix = Elev_make_std_b_matrix()
	//Zero all floor button lamps
	//var i int = 0
	for i := 0; i < N_FLOORS; i++ {
		if i != 0 {
			Elev_set_button_lamp(i, DOWN, false)
		}
		if i != N_FLOORS-1 {
			Elev_set_button_lamp(i, UP, false)
		}
		Elev_set_button_lamp(i, COMMAND, false)
	}
	//Clear stop lamp, foor open lamp, and set floor indicatior and ground floor.
	Elev_set_stop_lamp(false)
	Elev_set_door_open_lamp(false)
	Elev_set_floor_indicator(0)

	//Running down until it reaches a floor
	currentFloor := Elev_get_floor_sensor_signal()
	if currentFloor == -1 {
		Elev_set_speed(-300) //Can we write constants here?
		fmt.Println("current direction: ", Elev_get_direction())
		for currentFloor == -1 {
			currentFloor = Elev_get_floor_sensor_signal()
		}
		Elev_stop_elevator()
	}

	//Should current floo

	return nil
}

//I think that 0 is up and 1 is down. 				Needs to be checked.        FIIIIIXFIIIIIX
func Elev_get_direction() int {
	if io_read_bit(MOTORDIR) {
		return 1
	} else {
		return 0
	}

}

// This looks good, isnt +-300 the desired speeds?
func Elev_set_speed(speed float64) { //Float64 may be a problem later
	//In order to sharply stop the elevator, the direction bit is toggled, before setting speed to zero
	var last_speed float64 = 0 //Where is this otherwise, is it being used? Is this even necessary??? May be a bug here.
	//if to start (speed > 0)
	if speed > 0 {
		io_clear_bit(MOTORDIR)
	} else if speed < 0 {
		io_set_bit(MOTORDIR) // if to stop (speed == 0)
	} else if last_speed < 0 {
		io_clear_bit(MOTORDIR)
	} else if last_speed > 0 {
		io_set_bit(MOTORDIR)
	}
	last_speed = speed //last_speed commented out
	//Write new setting to motor
	io_write_analog(MOTOR, 2048+4*int(math.Abs(speed)))
}

// Check this motherfucker
func Elev_get_speed() float64 {
	return float64(Io_read_analog(MOTOR) - 2048)
}

func Elev_stop_elevator() {
	var toggleDir float64
	if x := math.Abs(Elev_get_speed()); x > 10 {
		//Toggles the direction in the opposite way. Down is 1
		toggleDir = 1
		if Elev_get_direction() != 1 {
			//fmt.Println("toggledir:", toggleDir)
			toggleDir = -1
		}

		//If the speed is over some value, you should brake that shit.
		Elev_set_speed(300 * toggleDir)
		time.Sleep(time.Millisecond * 10)
		Elev_set_speed(0)
		//Toggling the direction bit, since we have gone the other way for some time
		if Elev_get_direction() == 1 {
			io_clear_bit(MOTORDIR)
		} else if Elev_get_direction() == 0 {
			io_set_bit(MOTORDIR)
		}
	}
}

// This looks good. I suggest we find a smart use of channels to read and write light-bits
func Elev_get_floor_sensor_signal() int {
	if io_read_bit(SENSOR1) {
		return 0
	} else if io_read_bit(SENSOR2) {
		return 1
	} else if io_read_bit(SENSOR3) {
		return 2
	} else if io_read_bit(SENSOR4) {
		return 3
	}
	//by convention in go no else is used, the function will not continue if one of the previous returns is called(?)  Yngve: I think else is used plenty
	return -1
}

//This will return -1 if something fails  L1 if button is pressed ut should return 1
func Elev_get_button_signal(floor int, button int) int {
	if floor < 0 || floor >= N_FLOORS {
		return -1
	}
	if button < 0 || button >= N_BUTTONS {
		return -1
	}
	if button == UP && floor == N_FLOORS-1 {
		return -1
	}
	if button == DOWN && floor == 0 {
		return -1
	}
	if io_read_bit(button_channel_matrix[floor][button]) {
		return 1
	}
	return 0
}

// This looks good
func Elev_set_floor_indicator(floor int) error {
	if floor < 0 || floor >= N_FLOORS {
		return errors.New("Floor value not in valid region")
	}
	// Binary encoding. One light must always be on.  Yngve: This is no good.
	if floor >= 2 {
		io_set_bit(FLOOR_IND1)
	} else {
		io_clear_bit(FLOOR_IND1)
	}
	if floor == 1 || floor == 3 {
		io_set_bit(FLOOR_IND2)
	} else {
		io_clear_bit(FLOOR_IND2)
	}
	return nil
}

// There is no assert in go, Looks good
func Elev_set_button_lamp(floor int, button int, turnOn bool) error {

	if floor < 0 || floor >= N_FLOORS {
		return errors.New("Floor value not in valid region")
	}
	if button < 0 || button >= N_BUTTONS {
		return errors.New("Button value not in valid region")
	}
	if button == UP && floor == N_FLOORS-1 {
		return errors.New("This floor has no defined up button")
	}
	if button == DOWN && floor == 0 {
		return errors.New("This floor has no defined down button")
	}
	if turnOn {
		io_set_bit(lamp_channel_matrix[floor][button])
	} else {
		io_clear_bit(lamp_channel_matrix[floor][button])
	}
	return nil
}

//This looks good
func Elev_set_door_open_lamp(turn_on bool) {
	if turn_on {
		io_set_bit(DOOR_OPEN)
	} else {
		io_clear_bit(DOOR_OPEN)
	}
}

func Elev_set_stop_lamp(turn_on bool) {
	if turn_on {
		io_set_bit(LIGHT_STOP)
	} else {
		io_clear_bit(LIGHT_STOP)
	}
}
