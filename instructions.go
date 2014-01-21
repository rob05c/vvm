package main

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

// vector load
func (cu ControlUnit) lod(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.lod(a, i)
	}
}

// vector load
func (cu ControlUnit) sto(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.sto(a, i)
	}
}


// lod operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) lod(a int64, i int64) {
	if(!pe.Enabled) {
		return
	}
	pe.ArithmeticRegister = pe.Memory[a+i]
}

// lod operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) sto(a int64, i int64) {
	if(!pe.Enabled) {
		return
	}
	pe.Memory[a+i] = pe.ArithmeticRegister
}
