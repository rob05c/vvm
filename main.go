package main

import (
	"fmt"
	"strconv"
)

func main() {
	cu := NewControlUnit(64, 128, 64);
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
	a := createMatrix()
	b := createMatrix()
	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)
	offset += int64(len(b)) // row length
	loadMatrix(cu, c, offset)
	printMemory(cu)
	printMatrix(a)
//	printMatrix(matrix)

}

func createMatrix() [][]int64 {
	dim := 10
	matrix := make([][]int64, dim, dim)
	for i, _ := range matrix {
		matrix[i] = make([]int64, dim, dim)
		for j, _ := range matrix[i] {
			matrix[i][j] = int64(j+1)
		}
	}
	return matrix
}

func printMatrix(matrix [][]int64) {

	for i, row := range matrix {
		if i == 0 {
			fmt.Printf("⎡")
		} else if i == len(matrix) - 1 {
			fmt.Printf("⎣")
		} else {
			fmt.Printf("⎢")
		}

		for _, col := range row {
			fmt.Printf("%4d ", col)
		}

		if i == 0 {
			fmt.Printf("⎤")
		} else if i == len(matrix) - 1 {
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
		if int64(i) + offset >= int64(len(cu.PE[0].Memory)) { // all PEs have the same amount of memory, so we just grab the first
			break
		}
		for j, _ := range matrix[i] {
			if j >= len(cu.PE) {
				break
			}
			cu.PE[j].Memory[int64(i) + offset] = matrix[i][j]
		}
	}
}

func printMemory(cu *ControlUnit) {
	bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
	fmt.Printf("PE: ")
	for i, _ := range cu.PE {
		fmt.Printf("%3d", i)
	}
	fmt.Printf("\n")

	fmt.Printf("----")
	for i := 0; i < len(cu.PE); i++ {
		fmt.Printf("---")
	}
	fmt.Printf("\n")

	for i := 0; i < bytesPerPe; i++ {
		fmt.Printf("    ")
		for j := 0; j < len(cu.PE); j++ {
			pe := cu.PE[j]
			fmt.Printf("%3d", pe.Memory[i])
		}
		fmt.Printf("\n")
	}

}
