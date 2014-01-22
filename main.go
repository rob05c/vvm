package main

import (
	"fmt"
	"strconv"
)

func main() {
	cu := NewControlUnit(64, 128, 64);
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
	matrix := createMatrix()
	loadMatrix(cu, matrix)
	printMatrix(matrix)
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

func loadMatrix(cu *ControlUnit, matrix [][]int64) {
	for i, pe := range cu.PE {
		if i >= len(matrix) {
			break
		}
		for j, _ := range pe.Memory {
			if j >= len(matrix[i]) {
				break
			}
			pe.Memory[j] = matrix[i][j]
		}
	}
}
