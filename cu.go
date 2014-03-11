package main

import (
	"fmt"
)

type ControlUnit struct {
	ProgramCounter     int64
	IndexRegister      []int64
	ArithmeticRegister int64
	Mask               []bool
	LengthRegister     int64 // necessary?
	PE                 []ProcessingElement
	Memory             []int64
	Done               chan bool
	Verbose            bool ///< whether to print verbose details during execution
}


/*
nThe Memory is 1 "BytesPerElement" larger than the number of PEs. This is so the CU may have its own memory.
*/
func NewControlUnit(indexRegisters int, processingElements int, memoryBytesPerElement int) *ControlUnit {
	var cu ControlUnit
	memory := memoryBytesPerElement * (processingElements + 1) // +1 so the CU has its own memory
	cu.Memory = make([]int64, memory, memory)
	cu.IndexRegister = make([]int64, indexRegisters, indexRegisters)
	cu.Mask = make([]bool, processingElements, processingElements)
	cu.PE = make([]ProcessingElement, processingElements, processingElements)
	cu.Done = make(chan bool, processingElements)
	for i, _ := range cu.PE {
		mpos := i * memoryBytesPerElement
		mlen := mpos + memoryBytesPerElement
		pe := &cu.PE[i]
		pe.Memory = cu.Memory[mpos:mlen]
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
		pe.Done = cu.Done
		go pe.Run()
	}
	return &cu
}


func (cu *ControlUnit) Run(program Program) {
	cu.ProgramCounter = 0
	for cu.ProgramCounter != int64(len(program)/3) {
		pc := cu.ProgramCounter
		instruction := OpCode(program[pc*InstructionLength]) & 63 // 63 = 00111111
		if !isMem(instruction) {
			param1 := program[pc*InstructionLength]>>6 | program[pc*InstructionLength+1]<<2&63
			param2 := program[pc*InstructionLength+1]>>4 | program[pc*InstructionLength+2]<<4&63
			param3 := program[pc*InstructionLength+2] >> 2

			if cu.Verbose {
				fmt.Printf("Run() PC: %3d  IS: %5s  P1: %d  P2: %d  P3: %d\n", cu.ProgramCounter, instruction.String(), param1, param2, param3) // debug
			}
			cu.Execute(instruction, []byte{param1, param2, param3})
			if cu.Verbose {
				cu.PrintMachine() // debug
			}
		} else {
			param := program[pc*InstructionLength]>>6 | program[pc*InstructionLength+1]<<2&63
			memParam := uint16(program[pc*InstructionLength+1]>>4) | uint16(program[pc*InstructionLength+2])<<4
			if cu.Verbose {
				fmt.Printf("Run() PC: %3d  IS: %5s  P: %d  MP: %d\n", cu.ProgramCounter, instruction.String(), param, memParam) // debug
			}
			cu.ExecuteMem(instruction, param, memParam)
			if cu.Verbose {
				cu.PrintMachine() // debug
			}
		}
		cu.ProgramCounter++
	}
}

func (cu *ControlUnit) ExecuteMem(instruction OpCode, param byte, memParam uint16) {
	switch instruction {
	case isLdx:
		cu.Ldx(param, memParam)
	case isStx:
		cu.Stx(param, memParam)
	case isCload:
		cu.Cload(memParam)
	case isCstore:
		cu.Cstore(memParam)
	}
}

/// @param params must have as many members as the instruction takes parameters
func (cu *ControlUnit) Execute(instruction OpCode, params []byte) {
	switch instruction {
	case isLdxi:
		cu.Ldxi(params[0], params[1])
	case isIncx:
		cu.Incx(params[0], params[1])
	case isDecx:
		cu.Decx(params[0], params[1])
	case isMulx:
		cu.Mulx(params[0], params[1])
	case isCmpx:
		cu.Cmpx(params[0], params[1], params[2])
	case isCbcast:
		cu.Cbcast()
	case isLod:
		cu.Lod(params[0], params[1])
	case isSto:
		cu.Sto(params[0], params[1])
	case isAdd:
		cu.Add(params[0], params[1])
	case isSub:
		cu.Sub(params[0], params[1])
	case isMul:
		cu.Mul(params[0], params[1])
	case isDiv:
		cu.Div(params[0], params[1])
	case isBcast:
		cu.Bcast(params[0])
	case isMov:
		cu.Mov(RegisterType(params[0]), RegisterType(params[1])) ///< @todo change to be multiple 'instructions' ?
	case isRadd:
		cu.Radd()
	case isRsub:
		cu.Rsub()
	case isRmul:
		cu.Rmul()
	case isRdiv:
		cu.Rdiv()
	}
}

func (cu *ControlUnit) PrintMachine() {
	cu.printCu()
	cu.printPe()
	cu.printMemory()
}

func (cu *ControlUnit) printMemory() {
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
		fmt.Printf("    ")
		for j := 0; j < len(cu.PE); j++ {
			pe := cu.PE[j]
			fmt.Printf("%3d", pe.Memory[i])
		}
		fmt.Printf("\n")
	}

}

func (cu *ControlUnit) printPe() {
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

func (cu *ControlUnit) printCu() {
	/// @todo print Mask, Memory?
	fmt.Printf("PC: %d  AR: %d  LR: %d\nIR: %d\nMask: ", cu.ProgramCounter, cu.ArithmeticRegister, cu.LengthRegister, cu.IndexRegister)
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

//
// control instructions
//
func (cu *ControlUnit) Ldx(index byte, a uint16) {
	//	fmt.Printf("ldx: cu.index[%d] = cu.Memory[%d] (%d)"
	cu.IndexRegister[index] = cu.Memory[a]
}
func (cu *ControlUnit) Stx(index byte, a uint16) {
	cu.Memory[a] = cu.IndexRegister[index]
}
func (cu *ControlUnit) Ldxi(index byte, a byte) {
	cu.IndexRegister[index] = int64(a)
}
func (cu *ControlUnit) Incx(index byte, a byte) {
	cu.IndexRegister[index] += int64(a)
}
func (cu *ControlUnit) Decx(index byte, a byte) {
	cu.IndexRegister[index] -= int64(a)
}
func (cu *ControlUnit) Mulx(index byte, a byte) {
	cu.IndexRegister[index] *= int64(a)
}
func (cu *ControlUnit) Cload(index uint16) {
	cu.ArithmeticRegister = cu.Memory[index]
}
func (cu *ControlUnit) Cstore(index uint16) {
	cu.Memory[index] = cu.ArithmeticRegister
}

/// @todo fix this to take a larger jump (a). Byte only allows for 256 instructions. That's not a very big program
func (cu *ControlUnit) Cmpx(index byte, ix2 byte, a byte) {
	if cu.IndexRegister[index] < cu.IndexRegister[ix2] {
		cu.ProgramCounter = int64(a) - 1 // -1 because the PC will be incremented.
	}
}

// control broadcast. Broadcasts the Control's Arithmetic Register to every PE's Routing Register
func (cu *ControlUnit) Cbcast() {
	for i, _ := range cu.PE {
		cu.PE[i].RoutingRegister = cu.ArithmeticRegister
	}
}

// Block until all PE's are done
func (cu *ControlUnit) Barrier() {
	for i := 0; i != len(cu.PE); i++ {
		<-cu.Done
	}
}

//
// vector instructions
//
func (cu *ControlUnit) Lod(a byte, idx byte) {
	//	fmt.Printf("PE-Lod %d + %d (%d)\n", a, cu.IndexRegister[idx], idx)
	for i, _ := range cu.PE {
		cu.PE[i].Lod <- ByteTuple{a, byte(cu.IndexRegister[idx])} ///< @todo is this ok? Should we be loading the index register somewhere else?
	}
	cu.Barrier()
}
func (cu *ControlUnit) Sto(a byte, idx byte) {
	for i, _ := range cu.PE {
		cu.PE[i].Sto <- ByteTuple{a, byte(cu.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit) Add(a byte, idx byte) {
	for i, _ := range cu.PE {
		cu.PE[i].Add <- ByteTuple{a, byte(cu.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit) Sub(a byte, idx byte) {
	for i, _ := range cu.PE {
		cu.PE[i].Sub <- ByteTuple{a, byte(cu.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit) Mul(a byte, idx byte) {
	for i, _ := range cu.PE {
		cu.PE[i].Mul <- ByteTuple{a, byte(cu.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit) Div(a byte, idx byte) {
	for i, _ := range cu.PE {
		cu.PE[i].Div <- ByteTuple{a, byte(cu.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit) Bcast(idx byte) {
	idx = byte(cu.IndexRegister[idx]) ///< @todo is this ok? Should we be loading the index register here?
	for i, _ := range cu.PE {
		if !cu.PE[i].Enabled {
			continue
		}
		cu.PE[i].RoutingRegister = cu.PE[idx].RoutingRegister
	}
}
func (cu *ControlUnit) Mov(from RegisterType, to RegisterType) {
	/// @todo remove this? speed for safety?
	if from == to {
		return
	}
	for i, _ := range cu.PE {
		cu.PE[i].Mov <- ByteTuple{byte(from), byte(to)}
	}
	cu.Barrier()
}

func (cu *ControlUnit) Radd() {
	for i, _ := range cu.PE {
		cu.PE[i].Radd <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit) Rsub() {
	for i, _ := range cu.PE {
		cu.PE[i].Rsub <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit) Rmul() {
	for i, _ := range cu.PE {
		cu.PE[i].Rmul <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit) Rdiv() {
	for i, _ := range cu.PE {
		cu.PE[i].Rdiv <- true
	}
	cu.Barrier()
}
