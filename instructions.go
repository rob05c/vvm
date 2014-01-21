package main

type RegisterType int
const (
	peIndex = iota
	peRouting
	peArithmetic
)

type ControlUnit struct {
	ProgramCounter int64
	IndexRegister []int64
	ArithmeticRegister int64
	Mask []bool
	LengthRegister int64 // necessary?
	PE []ProcessingElement
	Memory []int64
}

type ProcessingElement struct {
	ArithmeticRegister int64
	RoutingRegister int64
	Index int64
	Enabled bool
	Memory []int64
}

/*
The Memory is 1 "BytesPerElement" larger than the number of PEs. This is so the CU may have its own memory.
*/
func NewControlUnit(indexRegisters int, processingElements int, memoryBytesPerElement int) *ControlUnit {
	var cu ControlUnit
	memory := (memoryBytesPerElement + 1) * processingElements // +1 so the CU has its own memory
	cu.Memory = make([]int64, memory, memory)
	cu.IndexRegister = make([]int64, indexRegisters, indexRegisters)
	cu.Mask = make([]bool, processingElements, processingElements)
	cu.PE = make([]ProcessingElement, processingElements, processingElements)
	for i, pe := range cu.PE {
		mpos := i * memoryBytesPerElement
		mlen := mpos + memoryBytesPerElement
		pe.Memory = cu.Memory[mpos:mlen]
	}
	return &cu
}




func (cu ControlUnit) Mov(from RegisterType, to RegisterType) {
	if from == to {
		return
	}
	for _, pe := range cu.PE {
		pe.Mov(from, to)
	}
}

// @todo make this more efficient
func (pe ProcessingElement) Mov(from RegisterType, to RegisterType) {
	if(!pe.Enabled) {
		return
	}
	var fromVal int64
	switch from {
	case peIndex:
		fromVal = pe.Index
	case peRouting:
		fromVal = pe.RoutingRegister
	case peArithmetic:
		fromVal = pe.ArithmeticRegister
	}
	switch to {
	case peIndex:
		pe.Index = fromVal
	case peRouting:
		pe.RoutingRegister = fromVal
	case peArithmetic:
		pe.ArithmeticRegister = fromVal
	}
}


// control broadcast. Broadcasts the Control's Arithmetic Register to every PE's Routing Register
func (cu ControlUnit) Cbcast() {
	for _, pe := range cu.PE {
		pe.RoutingRegister = cu.ArithmeticRegister
	}
}


// control load
func (cu ControlUnit) Cload(index int64) {
	cu.ArithmeticRegister = cu.Memory[index];
}

// vector load
func (cu ControlUnit) Lod(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Lod(a, i)
	}
}

// vector store
func (cu ControlUnit) Sto(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Sto(a, i)
	}
}


// lod operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) Lod(a int64, i int64) {
	if(!pe.Enabled) {
		return
	}
	pe.ArithmeticRegister = pe.Memory[a+i]
}

// sto operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) Sto(a int64, i int64) {
	if(!pe.Enabled) {
		return
	}
	pe.Memory[a+i] = pe.ArithmeticRegister
}

