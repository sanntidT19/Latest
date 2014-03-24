package driver

// #cgo LDFLAGS: -lcomedi -lm
// #include <png.h>
// #include "io.h"
import "C"

//C.io_init returns =! 0 if init has been successfull, io_init returns true for the same case
func io_init() bool {
	return bool(int(C.io_init()) != 0)
}

func io_set_bit(channel int) {
	C.io_set_bit(C.int(channel))
}

func io_clear_bit(channel int) {
	C.io_clear_bit(C.int(channel))
}

func io_write_analog(channel int, value int) {
	C.io_write_analog(C.int(channel), C.int(value))
}

//have changed this to return boolean values

func io_read_bit(channel int) bool {
	var i int = int(C.io_read_bit(C.int(channel)))
	if i == 1 {
		return true
	} else {
		return false
	}
}

func Io_read_analog(channel int) int {
	return int(C.io_read_analog(C.int(channel)))
}
