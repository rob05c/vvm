package main

type RegisterType int64

const (
	peIndex = iota
	peRouting
	peArithmetic
)

type ControlUnit struct {
	ProgramCounter     int64
	IndexRegister      []int64
	ArithmeticRegister int64
	Mask               []bool
	LengthRegister     int64 // necessary?
	PE                 []ProcessingElement
	Memory             []int64
}

type ProcessingElement struct {
	ArithmeticRegister int64
	RoutingRegister    int64
	Index              int64
	Enabled            bool
	Memory             []int64
}

/*
The Memory is 1 "BytesPerElement" larger than the number of PEs. This is so the CU may have its own memory.
*/
func NewControlUnit(indexRegisters int, processingElements int, memoryBytesPerElement int) *ControlUnit {
	var cu ControlUnit
	memory := memoryBytesPerElement * (processingElements + 1) // +1 so the CU has its own memory
	cu.Memory = make([]int64, memory, memory)
	cu.IndexRegister = make([]int64, indexRegisters, indexRegisters)
	cu.Mask = make([]bool, processingElements, processingElements)
	cu.PE = make([]ProcessingElement, processingElements, processingElements)
	for i, _ := range cu.PE {
		mpos := i * memoryBytesPerElement
		mlen := mpos + memoryBytesPerElement
		cu.PE[i].Memory = cu.Memory[mpos:mlen]
	}
	return &cu
}

type InstructionType int64

const (
	// control instructions
	isLdx = iota
	isStx
	isLdxi
	isIncx
	isDecx
	isMulx
	isCload
	isCstore
	isCmpx
	isCbcast
	// vector instructions
	isLod
	isSto
	isAdd
	isSub
	isMul
	isDiv
	isBcast
	isMov
	isRadd
	isRsub
	isRmul
	isRdiv
)

var InstructionParams = map[InstructionType]byte{
	isLdx:    2,
	isStx:    2,
	isLdxi:   2,
	isIncx:   2,
	isDecx:   2,
	isMulx:   2,
	isCload:  1,
	isCstore: 1,
	isCmpx:   3,
	isCbcast: 0,
	isLod:    2,
	isSto:    2,
	isAdd:    2,
	isSub:    2,
	isMul:    2,
	isDiv:    2,
	isBcast:  1,
	isMov:    2,
	isRadd:   0,
	isRsub:   0,
	isRmul:   0,
	isRdiv:   0,
}

func (cu ControlUnit) Run(program []int64) {
	cu.ProgramCounter = 0
	for cu.ProgramCounter != int64(len(program)) {
		pc := cu.ProgramCounter
		instruction := InstructionType(program[pc])
		params := program[pc : pc+int64(InstructionParams[instruction])]
		cu.Execute(instruction, params)
		cu.ProgramCounter++
	}
}

/// @param params must have as many members as the instruction takes parameters
func (cu ControlUnit) Execute(instruction InstructionType, params []int64) {
	switch instruction {
	case isLdx:
		cu.Ldx(params[0], params[1])
	case isStx:
		cu.Stx(params[0], params[1])
	case isLdxi:
		cu.Ldxi(params[0], params[1])
	case isIncx:
		cu.Incx(params[0], params[1])
	case isDecx:
		cu.Decx(params[0], params[1])
	case isMulx:
		cu.Mulx(params[0], params[1])
	case isCload:
		cu.Cload(params[0])
	case isCstore:
		cu.Cstore(params[0])
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

//
// control instructions
//
func (cu ControlUnit) Ldx(index int64, a int64) {
	//	fmt.Printf("ldx: cu.index[%d] = cu.Memory[%d] (%d)"
	cu.IndexRegister[index] = cu.Memory[a]
}
func (cu ControlUnit) Stx(index int64, a int64) {
	cu.Memory[a] = cu.IndexRegister[index]
}
func (cu ControlUnit) Ldxi(index int64, a int64) {
	cu.IndexRegister[index] = a
}
func (cu ControlUnit) Incx(index int64, a int64) {
	cu.IndexRegister[index] += a
}
func (cu ControlUnit) Decx(index int64, a int64) {
	cu.IndexRegister[index] -= a
}
func (cu ControlUnit) Mulx(index int64, a int64) {
	cu.IndexRegister[index] *= a
}
func (cu ControlUnit) Cload(index int64) {
	cu.ArithmeticRegister = cu.Memory[index]
}
func (cu ControlUnit) Cstore(index int64) {
	cu.Memory[index] = cu.ArithmeticRegister
}
func (cu ControlUnit) Cmpx(index int64, ix2 int64, a int64) {
	if cu.IndexRegister[index] <= cu.IndexRegister[ix2] {
		cu.ProgramCounter = a
	}
}

// control broadcast. Broadcasts the Control's Arithmetic Register to every PE's Routing Register
func (cu ControlUnit) Cbcast() {
	for _, pe := range cu.PE {
		pe.RoutingRegister = cu.ArithmeticRegister
	}
}

//
// vector instructions
//
func (cu ControlUnit) Lod(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Lod(a, i)
	}
}
func (cu ControlUnit) Sto(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Sto(a, i)
	}
}
func (cu ControlUnit) Add(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Add(a, i)
	}
}
func (cu ControlUnit) Sub(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Sub(a, i)
	}
}
func (cu ControlUnit) Mul(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Mul(a, i)
	}
}
func (cu ControlUnit) Div(a int64, i int64) {
	for _, pe := range cu.PE {
		pe.Div(a, i)
	}
}
func (cu ControlUnit) Bcast(i int64) {
	for _, pe := range cu.PE {
		if !pe.Enabled {
			continue
		}
		pe.RoutingRegister = cu.PE[i].RoutingRegister
	}
}
func (cu ControlUnit) Mov(from RegisterType, to RegisterType) {
	if from == to {
		return
	}
	for _, pe := range cu.PE {
		pe.Mov(from, to)
	}
}

func (cu ControlUnit) Radd() {
	for _, pe := range cu.PE {
		pe.Radd()
	}
}
func (cu ControlUnit) Rsub() {
	for _, pe := range cu.PE {
		pe.Rsub()
	}
}
func (cu ControlUnit) Rmul() {
	for _, pe := range cu.PE {
		pe.Rmul()
	}
}
func (cu ControlUnit) Rdiv() {
	for _, pe := range cu.PE {
		pe.Rdiv()
	}
}

///
/// PE (vector) instructions
///
func (pe ProcessingElement) Add(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister += pe.Memory[a+i]
}
func (pe ProcessingElement) Sub(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister -= pe.Memory[a+i]
}
func (pe ProcessingElement) Mul(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister *= pe.Memory[a+i]
}
func (pe ProcessingElement) Div(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	if pe.ArithmeticRegister == 0 {
		return
	}
	pe.ArithmeticRegister /= pe.Memory[a+i]
}

// lod operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) Lod(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister = pe.Memory[a+i]
}

// sto operation for individual PE
// @todo change this to be signalled by a channel
func (pe ProcessingElement) Sto(a int64, i int64) {
	if !pe.Enabled {
		return
	}
	pe.Memory[a+i] = pe.ArithmeticRegister
}

// @todo make this more efficient
func (pe ProcessingElement) Mov(from RegisterType, to RegisterType) {
	if !pe.Enabled {
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
func (pe ProcessingElement) Radd() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister += pe.RoutingRegister
}
func (pe ProcessingElement) Rsub() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister -= pe.RoutingRegister
}
func (pe ProcessingElement) Rmul() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister *= pe.RoutingRegister
}
func (pe ProcessingElement) Rdiv() {
	if !pe.Enabled {
		return
	}
	pe.ArithmeticRegister /= pe.RoutingRegister
}
