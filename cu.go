package main

import (
	"fmt"
)

/// @todo rename this, and ducks
type ControlUnit interface {
	Run(file string) error
	RunProgram(program Program) error
	PrintMachine()
	Data() *ControlUnitData
}

type ControlUnitData struct {
	IndexRegister      []int64
	ArithmeticRegister int64
	Mask               []bool
	LengthRegister     int64 // necessary?
	PE                 []ProcessingElement
	Memory             []int64
	Verbose            bool ///< whether to print verbose details during execution
	Done               chan bool
}

/*
The Memory is 1 "BytesPerElement" larger than the number of PEs. This is so the CU may have its own memory.
*/
func NewControlUnitData(indexRegisters uint, processingElements uint, memoryBytesPerElement uint) *ControlUnitData {
	var d ControlUnitData
	d.Verbose = true;
	memory := memoryBytesPerElement * (processingElements + 1) // +1 so the CU has its own memory
	d.Memory = make([]int64, memory, memory)
	d.IndexRegister = make([]int64, indexRegisters, indexRegisters)
	d.Mask = make([]bool, processingElements, processingElements)
	d.PE = make([]ProcessingElement, processingElements, processingElements)
	d.Done = make(chan bool, processingElements)

	for i, _ := range d.PE {
		mpos := i * int(memoryBytesPerElement)
		mlen := mpos + int(memoryBytesPerElement)
		pe := &d.PE[i]
		pe.Memory = d.Memory[mpos:mlen]
		pe.Enabled = true
		pe.Lod = make(chan ByteTuple)
		pe.Sto = make(chan ByteTuple)
		pe.Add = make(chan ByteTuple)
		pe.Sub = make(chan ByteTuple)
		pe.Mul = make(chan ByteTuple)
		pe.Div = make(chan ByteTuple)
		pe.Mov = make(chan ByteTuple)
		pe.Radd = make(chan bool)
		pe.Rsub = make(chan bool)
		pe.Rmul = make(chan bool)
		pe.Rdiv = make(chan bool)
		pe.Done = d.Done
		go pe.Run()
	}
	return &d
}

func (cu *ControlUnitData) PrintMachine() {
	cu.printCu()
	cu.printPe()
	cu.printMemory()
}

func (cu *ControlUnitData) printMemory() {
	bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
	/*
		fmt.Printf("PE: ")
		for i, _ := range cu.PE {
			fmt.Printf("%3d", i)
		}
		fmt.Printf("\n")
	*/
	fmt.Printf("----")
	for i := 0; i < len(cu.PE); i++ {
		fmt.Printf("---")
	}
	fmt.Printf("\n")

	for i := 0; i < bytesPerPe; i++ {
//		if i > 8 {
//			break //debug
//		}
		fmt.Printf("    ")
		for j := 0; j < len(cu.PE); j++ {
			pe := cu.PE[j]
			fmt.Printf("%3d", pe.Memory[i])
		}
		fmt.Printf("\n")
	}

}

func (cu *ControlUnitData) printPe() {
	//	bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
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

	fmt.Printf("AR: ")
	for j := 0; j < len(cu.PE); j++ {
		pe := cu.PE[j]
		fmt.Printf("%3d", pe.ArithmeticRegister)
	}
	fmt.Printf("\n")

	fmt.Printf("RR: ")
	for j := 0; j < len(cu.PE); j++ {
		pe := cu.PE[j]
		fmt.Printf("%3d", pe.RoutingRegister)
	}
	fmt.Printf("\n")

	fmt.Printf("Ix: ")
	for j := 0; j < len(cu.PE); j++ {
		pe := cu.PE[j]
		fmt.Printf("%3d", pe.Index)
	}
	fmt.Printf("\n")

	fmt.Printf("En: ")
	for j := 0; j < len(cu.PE); j++ {
		pe := cu.PE[j]
		if pe.Enabled {
			fmt.Printf("%3d", 1)
		} else {
			fmt.Printf("%3d", 0)
		}
	}
	fmt.Printf("\n")

}

func (cu *ControlUnitData) printCu() {
	/// @todo print Mask, Memory?
	/// @todo print Program Counter
	fmt.Printf("AR: %d  LR: %d\nIR: %d\nMask: ", cu.ArithmeticRegister, cu.LengthRegister, cu.IndexRegister)
	for i := 0; i < len(cu.Mask); i++ {
		if cu.Mask[i] {
			fmt.Printf("1  ")
		} else {
			fmt.Printf("0  ")
		}
	}
	fmt.Print("\nMem:")

	bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
	cuMemoryBegin := len(cu.PE) * bytesPerPe
	for i := cuMemoryBegin; i < len(cu.Memory); i++ {
		if i != cuMemoryBegin && i%len(cu.PE) == 0 {
			fmt.Print("\n    ")
		}
		fmt.Printf("%3d", cu.Memory[i])
	}

	bar := "----"
	for i := 0; i < len(cu.PE); i++ {
		bar += "---"
	}
	fmt.Println("\n" + bar)
}
