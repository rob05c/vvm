package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

type ArchitectureType uint
const (
	at24bit = ArchitectureType(iota)
	at24bitpipelined
	at32bit
)

const version = "1.0.1"
var compileFile string
var outputFile string
var verbose bool
var script bool
var archString string
var arch ArchitectureType
var memoryPerPe uint
var numPe uint
var numIndexRegisters uint

func init() {
	const (
		compileDefault = ""
		compileUsage   = "file to compile"
		outputDefault  = "output.simd"
		outputUsage    = "output file for compiled binary"
		verboseDefault = false
		verboseUsage   = "Verbose output, prints the state of the machine after each instruction"
		scriptDefault = false
		scriptUsage   = "Act as script, immediately output the execution result of the given assembly file."
		archDefault    = "24bit"
		archUsage      = "Machine architecture: 24bit, 24bitpipelined, 32bit."
		peMemDefault   = 64
		peMemUsage     = `Memory per processing element. 
        CAUTION: setting more than the instruction set can address will result in undefined behavior.`
		numPeDefault   = 32
		numPeUsage     = `Number of processing elements. 
        CAUTION: setting more than the instruction set can address will result in undefined behavior.`
		numIndexRegistersDefault   = 64
		numIndexRegistersUsage = `Number of index registers. 
        CAUTION: setting more than the instruction set can address will result in undefined behavior.`
	)
	flag.StringVar(&compileFile, "compile", compileDefault, compileUsage)
	flag.StringVar(&compileFile, "c", compileDefault, compileUsage+" (shorthand)")
	flag.StringVar(&outputFile, "output", outputDefault, outputUsage)
	flag.StringVar(&outputFile, "o", outputDefault, outputUsage+" (shorthand)")
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseUsage)
	flag.BoolVar(&verbose, "v", verboseDefault, verboseUsage+" (shorthand)")
	flag.BoolVar(&script, "script", scriptDefault, scriptUsage)
	flag.BoolVar(&script, "s", scriptDefault, scriptUsage+" (shorthand)")
	flag.StringVar(&archString, "arch", archDefault, archUsage)
	flag.StringVar(&archString, "a", archDefault, archUsage+" (shorthand)")
	flag.UintVar(&memoryPerPe, "pemem", peMemDefault, peMemUsage)
	flag.UintVar(&numPe, "numpe", numPeDefault, numPeUsage)
	flag.UintVar(&numIndexRegisters, "indexregisters", numIndexRegistersDefault, numIndexRegistersUsage)
}

func printUsage() {
	exeName := os.Args[0]
	fmt.Println(exeName + " " + version + " usage: ")
	fmt.Println("\t" + exeName + " -c file-to-compile.sasm -o output-file")
	fmt.Println("\t" + exeName + " file-to-execute.simd")
	fmt.Println("\t" + exeName + " -s file-to-execute.sasm")
	fmt.Println("flags:")
	flag.PrintDefaults()
	fmt.Println("example:\n\t" + exeName + " -c input.sasm")
	fmt.Println("\t" + exeName + " -c input.sasm -o matrix_multiply.simd")
	fmt.Println("\t" + exeName + " output.simd")
}

func parseEnumArgs() {
	switch archString {
	case "24bit":
		arch = at24bit
	case "24bitpipelined":
		arch = at24bitpipelined
	case "32bit":
		arch = at32bit
	default:
		arch = at24bit
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	parseEnumArgs()

	var cu ControlUnit
	switch arch {
	case at24bit:
		cu = NewControlUnit24bit(numIndexRegisters, numPe, memoryPerPe)
	case at24bitpipelined:
		cu = NewControlUnit24bitPipelined(numIndexRegisters, numPe, memoryPerPe)
	case at32bit:
		cu = NewControlUnit32bit(numIndexRegisters, numPe, memoryPerPe)
	default: 
		cu = NewControlUnit24bit(numIndexRegisters, numPe, memoryPerPe)
	}
	cu.Data().Verbose = verbose

	if script {
		compileFile = flag.Arg(0)
		program, err := compile(cu, arch)
		if err != nil {
			fmt.Println(err)
			return
		}
		runProgram(cu, program)
		return
	}

	if len(compileFile) != 0 {
		program, err := compile(cu, arch)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = program.Save(outputFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	if flag.NArg() <= 0 {
		printUsage()
		return
	}
	run(cu)
}

func compile(cu ControlUnit, arch ArchitectureType) (Program, error) {
	bytes, err := ioutil.ReadFile(compileFile)
	if err != nil {
		return nil, err
	}

	input := string(bytes)
	var program Program
	switch arch {
	case at24bit:
	fallthrough
	case at24bitpipelined:
		program = NewProgram24bit()
	case at32bit:
		program = NewProgram32bit()
	default:
		program = NewProgram24bit()
	}
	err = LexProgram(cu.Data(), input, program)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func run(cu ControlUnit) {
	testLoadMatrices(cu.Data()) ///< @todo change sample input to load matrices within instructions, so this is unnecessary

	programFile := flag.Arg(0)
	start := time.Now()
	cu.Run(programFile)
	executionTime := time.Now().Sub(start)
	fmt.Print("Program executed in ")
	fmt.Print(executionTime)
	fmt.Print(" on ")
	fmt.Print(runtime.GOMAXPROCS(0))
	fmt.Println(" cores.")
	cu.PrintMachine()
	/*
		pr, err := NewProgramReader(programFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		for instruction, err := pr.ReadInstruction(); err == nil; instruction, err = pr.ReadInstruction() {
			fmt.Println(instruction)
		}
	*/
}

func runProgram(cu ControlUnit, program Program) {
	testLoadMatrices(cu.Data()) ///< @todo change sample input to load matrices within instructions, so this is unnecessary
	start := time.Now()
	cu.RunProgram(program)
	executionTime := time.Now().Sub(start)
	fmt.Print("Program executed in ")
	fmt.Print(executionTime)
	fmt.Print(" on ")
	fmt.Print(runtime.GOMAXPROCS(0))
	fmt.Println(" cores.")
	cu.PrintMachine()
	/*
		pr, err := NewProgramReader(programFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		for instruction, err := pr.ReadInstruction(); err == nil; instruction, err = pr.ReadInstruction() {
			fmt.Println(instruction)
		}
	*/
}
