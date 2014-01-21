package main

import (
	"fmt"
	"strconv"
)

func main() {
	cu := NewControlUnit(64, 128, 64);
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
}
