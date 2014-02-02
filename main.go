package main

import (
	"fmt"
	"flag"
	"io/ioutil"
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
		compile(cu)
		return
	}

	if flag.NArg() > 0 {
		run(cu)
		return
	}

	flag.PrintDefaults()
	return
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
	cu.Run(program)
}
