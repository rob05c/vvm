package main

import (
	"fmt"
	"strconv"
)

func testLoadMatrices(cu *ControlUnitData) {
	n := 3
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)
}

// @todo fix movA toR
func testLexer() {
	input := `
lim equiv 0
i equiv 1
j equiv 2
n data 3
zero data 0
a bss 3x3
b bss 3x3
c bss 3x3

ldxi i,0

ldxi j,0
ldx lim,n
loop:
lod a,i
mov 2,1
bcast j
lod b,j
rmul
add c,i
sto c,i
incx j,1
cmpx j,lim,loop
ldxi j,0
incx i,1
cmpx i,lim,loop
`
	cu := NewControlUnit24bitPipelined(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)

	//	lines, program, err := ParsePseudoOperations(cu,
	//	lines)
	program := NewProgram24bit()
	err := LexProgram(cu.Data(), input, program)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print("program: ")
	fmt.Println(program)

	fmt.Println("RUNNING")

	n := 3
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu.Data(), a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu.Data(), b, offset)

	cu.Run(input)
	//	lines = ReplacePseudoOpAliases(lines, aliases)
}

func testMatrixMultiply() {
	cu := NewControlUnit24bitPipelined(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.Data().PE)) + " processing elements.") //debug
	n := 3
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu.Data(), a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu.Data(), b, offset)

	fmt.Println("main() Start State")
	cu.PrintMachine()
	fmt.Println("main() Multiplying...\n")
	matrixMultiply(cu, byte(n))
	fmt.Println("main() Final State")
	cu.PrintMachine()
}

func createMatrix(n int) [][]int64 {
	matrix := make([][]int64, n, n)
	for i, _ := range matrix {
		matrix[i] = make([]int64, n, n)
		for j, _ := range matrix[i] {
			matrix[i][j] = int64(j + 2)
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
func loadMatrix(cu *ControlUnitData, matrix [][]int64, offset int64) {
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
func matrixMultiply(cu ControlUnit, matrixDimension byte) {
	// indices into the CU Index Register
	nextRegister := byte(0)
	lim := nextRegister
	nextRegister++

	i := nextRegister
	nextRegister++

	j := nextRegister
	nextRegister++

	// indices into memory
	a := byte(0)
	b := a + matrixDimension
	c := b + matrixDimension

	var program Program24bit
	n := program.DataOp(cu.Data(), matrixDimension)

	//	zero :=
	program.DataOp(cu.Data(), 0)

	program.Push(isLdxi, []byte{i, 0, 0})
	program.Push(isLdxi, []byte{j, 0, 0})

	program.PushMem(isLdx, lim, n)

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
	program.Push(isLdxi, []byte{j, 0, 0})
	program.Push(isIncx, []byte{i, 1, 0})
	program.Push(isCmpx, []byte{i, lim, labelLoop})

	fmt.Print("matrixMultiply() Program: ")
	fmt.Println(program)
	fmt.Print("\n")
	file := "program.simd"
	err := program.Save(file)
	if err != nil {
		fmt.Println(err)
	}
	/*
		program2, err := LoadProgram(file)
		if err != nil {
			fmt.Println(err)
			return
		}
	*/
	err = cu.Run(file)
	if err != nil {
		fmt.Println(err)
	}
}
