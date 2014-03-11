package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

const DefaultIndexRegisters = 64
const DefaultProcessingElements = 16
const DefaultMemoryPerElement = 32

var compileFile string
var outputFile string
var verbose bool

func init() {
	const (
		compileDefault = ""
		compileUsage   = "file to compile"
		outputDefault  = "output.simd"
		outputUsage    = "output file for compiled binary"
		verboseDefault = false
		verboseUsage   = "Verbose output, prints the state of the machine after each instruction"
	)
	flag.StringVar(&compileFile, "compile", compileDefault, compileUsage)
	flag.StringVar(&compileFile, "c", compileDefault, compileUsage+" (shorthand)")
	flag.StringVar(&outputFile, "output", outputDefault, outputUsage)
	flag.StringVar(&outputFile, "o", outputDefault, outputUsage+" (shorthand)")
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseUsage)
	flag.BoolVar(&verbose, "v", verboseDefault, verboseUsage+" (shorthand)")
}

func printUsage() {
	exeName := os.Args[0]
	fmt.Println("usage: ")
	fmt.Println("\t" + exeName + " -c file-to-compile.sasm -o output-file")
	fmt.Println("\t" + exeName + " file-to-execute.simd")
	fmt.Println("flags:")
	flag.PrintDefaults()
	fmt.Println("example:\n\t" + exeName + " -c input.sasm")
	fmt.Println("\t" + exeName + " -c input.sasm -o matrix_multiply.simd")
	fmt.Println("\t" + exeName + " output.simd")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	cu := NewControlUnit(DefaultIndexRegisters, DefaultProcessingElements, DefaultMemoryPerElement)
	cu.Verbose = verbose
	if len(compileFile) != 0 {
		compile(cu)
		return
	}
	if flag.NArg() <= 0 {
		printUsage()
		return
	}
	run(cu)
}

func compile(cu *ControlUnit) {
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
}

func run(cu *ControlUnit) {
	testLoadMatrices(cu) ///< @todo change sample input to load matrices within instructions, so this is unnecessary

	programFile := flag.Arg(0)
	program, err := LoadProgram(programFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	start := time.Now()
	cu.Run(program)
	executionTime := time.Now().Sub(start)
	fmt.Print("Program executed in ")
	fmt.Print(executionTime)
	fmt.Print(" on ")
	fmt.Print(runtime.GOMAXPROCS(0))
	fmt.Println(" cores.")
	cu.PrintMachine()
}
