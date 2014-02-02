package main

import (
	"fmt"
	"strconv"
	"flag"
	"io/ioutil"
//	"strings"
)

const DefaultIndexRegisters = 64
const DefaultProcessingElements = 16
const DefaultMemoryPerElement = 32

var compileFile string
var outputFile string
func init() {
	const (
		compileDefault = ""
		compileUsage = "file to compile"
		outputDefault = "output.simd"
		outputUsage = "output file for compiled binary"

	)
	flag.StringVar(&compileFile, "compile", compileDefault, compileUsage)
	flag.StringVar(&compileFile, "c", compileDefault, compileUsage+" (shorthand)")
	flag.StringVar(&outputFile, "output", outputDefault, outputUsage)
	flag.StringVar(&outputFile, "o", outputDefault, outputUsage+" (shorthand)")
}

func main() {
	flag.Parse()

	cu := NewControlUnit(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)

	if len(compileFile) != 0 {
		bytes, err := ioutil.ReadFile(compileFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		
		input := string(bytes)
		program, err := LexProgram(cu, input)
		if err != nil {
			fmt.Println(err)
			return
		}
		
		err = program.Save(outputFile)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	if flag.NArg() > 0 {
		testLoadMatrices(cu) // debug

		programFile := flag.Arg(0)
		program, err := LoadProgram(programFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		cu.Run(program)
		return
	}

	flag.PrintDefaults()
	return
}

func testLoadMatrices(cu *ControlUnit) {
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
	cu := NewControlUnit(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)

	//	lines, program, err := ParsePseudoOperations(cu, lines)
	program, err := LexProgram(cu, input)
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
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)

	cu.Run(program)
	//	lines = ReplacePseudoOpAliases(lines, aliases)
}

func testMatrixMultiply() {
	cu := NewControlUnit(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)
	fmt.Println("You made a vector VM with " + strconv.Itoa(len(cu.PE)) + " processing elements.") //debug
	n := 3
	a := createMatrix(n)
	b := createMatrix(n)
	//	c := createMatrix()
	offset := int64(0)
	loadMatrix(cu, a, offset)
	offset += int64(len(a)) // row length
	loadMatrix(cu, b, offset)

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
func matrixMultiply(cu *ControlUnit, matrixDimension byte) {
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

	var pprogram PseudoProgram
	n := pprogram.DataOp(cu, matrixDimension)

	//	zero :=
	pprogram.DataOp(cu, 0)

	program := CreateProgram(pprogram)
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
	program2, err := LoadProgram(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	cu.Run(program2)
}
