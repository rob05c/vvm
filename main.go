package main

import (
	"fmt"
	"strconv"
)

func main() {
	cu := NewControlUnit(64, 16, 32)
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
	n := 10
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)

	cu.PrintMachine()
	fmt.Println("Multiplying...\n")
	matrixMultiply(cu, byte(n))
	cu.PrintMachine()

	//	program := addInstruction(nil, isMov, []byte{4, 9, 15})
}

func createMatrix(n int) [][]int64 {
	matrix := make([][]int64, n, n)
	for i, _ := range matrix {
		matrix[i] = make([]int64, n, n)
		for j, _ := range matrix[i] {
			matrix[i][j] = int64(j + 1)
		}
	}
	return matrix
}

func printMatrix(matrix [][]int64) {

	for i, row := range matrix {
		if i == 0 {
			fmt.Printf("⎡")
		} else if i == len(matrix)-1 {
			fmt.Printf("⎣")
		} else {
			fmt.Printf("⎢")
		}

		for _, col := range row {
			fmt.Printf("%4d ", col)
		}

		if i == 0 {
			fmt.Printf("⎤")
		} else if i == len(matrix)-1 {
			fmt.Printf("⎦")
		} else {
			fmt.Printf("⎥")
		}
		fmt.Printf("\n")
	}
}

/// @todo account for matrices larger than the number of PEs
/// @param offset the memory offset to begin at,
func loadMatrix(cu *ControlUnit, matrix [][]int64, offset int64) {
	for i, _ := range matrix {
		if int64(i)+offset >= int64(len(cu.PE[0].Memory)) { // all PEs have the same amount of memory, so we just grab the first
			break
		}
		for j, _ := range matrix[i] {
			if j >= len(cu.PE) {
				break
			}
			cu.PE[j].Memory[int64(i)+offset] = matrix[i][j]
		}
	}
}

// multiplies matrices of the given dimension, assuming they're stored by-column,
// starting in memory location 0 and consecutive,
// and stores the result in the immediately following memory
//
/// @param n dimension of the matrices
func matrixMultiply(cu *ControlUnit, n byte) {
	// indices into the CU Index Register
	lim := byte(1)
	i := byte(2)
	j := byte(3)

	// indices into memory
	a := byte(0)
	b := byte(a + n)
	c := byte(b + n)

	var program Program
	program.Push(isLdx, []byte{j, 0, 0})
	labelLoop := program.Size()
	program.Push(isLod, []byte{a, i, 0})
	program.Push(isMov, []byte{peArithmetic, peRouting, 0})
	program.Push(isBcast, []byte{j, 0, 0})
	program.Push(isLod, []byte{b, j, 0})
	program.Push(isRmul, []byte{0, 0, 0})
	program.Push(isAdd, []byte{c, i, 0})
	program.Push(isSto, []byte{c, i, 0})
	program.Push(isIncx, []byte{j, 1, 0})
	program.Push(isCmpx, []byte{j, lim, labelLoop})
	fmt.Print("Program: ")
	fmt.Println(program)
	fmt.Print("\n")
	cu.Run(program)
}
