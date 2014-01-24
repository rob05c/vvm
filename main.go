package main

import (
	"fmt"
	"strconv"
)

func main() {
	cu := NewControlUnit(64, 16, 16)
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
	n := 10
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)
	printMemory(cu)
	fmt.Println("Multiplying...\n")
	matrixMultiply(cu, int64(n))
	printMemory(cu)
	//	printMatrix(matrix)

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

// multiplies matrices of the given dimension, assuming they're stored by-column,
// starting in memory location 0 and consecutive,
// and stores the result in the immediately following memory
//
/// @param n dimension of the matrices

func matrixMultiply(cu *ControlUnit, n int64) {
	// indices into the CU Index Register
	lim := int64(1)
	i := int64(2)
	j := int64(3)

	// indices into memory
	a := int64(0)
	b := int64(a + n)
	c := int64(b + n)

	var program []int64
	program = append(program, isLdx)
	program = append(program, j)
	program = append(program, 0)
	labelLoop := int64(len(program))
	program = append(program, isLod)
	program = append(program, a)
	program = append(program, i)
	program = append(program, isMov)
	program = append(program, peArithmetic)
	program = append(program, peRouting)
	program = append(program, isBcast)
	program = append(program, j)
	program = append(program, isLod)
	program = append(program, b)
	program = append(program, j)
	program = append(program, isRmul)
	program = append(program, isAdd)
	program = append(program, c)
	program = append(program, i)
	program = append(program, isSto)
	program = append(program, c)
	program = append(program, i)
	program = append(program, isIncx)
	program = append(program, j)
	program = append(program, 1)
	program = append(program, isCmpx)
	program = append(program, j)
	program = append(program, lim)
	program = append(program, labelLoop)
	cu.Run(program)
}
