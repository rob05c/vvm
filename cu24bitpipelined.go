package main

import (
	"fmt"
)

const NoJump = int64(-1)

// used as a "union" for ExecuteChan
type ExecuteParam interface {
	IsMem() bool
	Op() OpCode
	Params() []byte
	Param() byte
	MemParam() uint16
}

type ExecuteParams struct {
	op     OpCode
	params []byte
}
func (p ExecuteParams) IsMem() bool {
	return false
}
func (p ExecuteParams) Op() OpCode {
	return p.op
}
func (p ExecuteParams) Params() []byte {
	return p.params
}
func (p ExecuteParams) Param() byte {
	return byte(0)
}
func (p ExecuteParams) MemParam() uint16 {
	return uint16(0)
}

type ExecuteMemParams struct {
	op       OpCode
	param    byte
	memParam uint16
}
func (p ExecuteMemParams) IsMem() bool {
	return true
}
func (p ExecuteMemParams) Op() OpCode {
	return p.op
}
func (p ExecuteMemParams) Params() []byte {
	return nil
}
func (p ExecuteMemParams) Param() byte {
	return p.param
}
func (p ExecuteMemParams) MemParam() uint16 {
	return p.memParam
}

type ControlUnit24bitPipelined struct {
	data *ControlUnitData

	FetchWaitForPcChange chan bool
	FetchPcChangeChan        chan int64
	FetchFinished  chan bool
	FetchStop chan bool

	DecodeChan         chan []byte
	DecodeFinished chan bool
	DecodeStop     chan bool
	DecodePause        chan bool
	DecodeResume       chan bool

	ExecuteChan     chan ExecuteParam
	ExecuteStopChan chan bool
	ExecuteFinishedChan chan bool

	Finished        chan bool
}

func tryGetPcChange(fetchWaitForPcChange <-chan bool, fetchPcChange <-chan int64, pc *int64) {
	select {
	case <-fetchWaitForPcChange:
		*pc = <-fetchPcChange
	default:
	}
}
func getPcChange(fetchWaitForPcChange <-chan bool, fetchPcChange <-chan int64, pc *int64, fetchStop <-chan bool) bool {
	select {
	case <-fetchWaitForPcChange:
		*pc = <-fetchPcChange
		return true
	case <-fetchStop:
		return false
	}
}
/// INVARIANT fetchWaitForPcChange MUST be passed BEFORE decodePause
func Fetcher(pr ProgramReader,
	decode chan<- []byte,
	fetchWaitForPcChange <-chan bool,
	fetchPcChange <-chan int64,
	fetchFinished chan<- bool,
	fetchStop <-chan bool) {

	pc := int64(0)
	cache := make(map[int64][]byte)

	for {
		tryGetPcChange(fetchWaitForPcChange, fetchPcChange, &pc)
		if instruction, ok := cache[pc]; ok {
			decode<- instruction
			pc++
		}
		instruction, err := pr.ReadInstruction(pc)
		if err == nil {
			cache[pc] = instruction
			decode<- instruction
			pc++
			continue
		}
		fetchFinished<- true
		if !getPcChange(fetchWaitForPcChange, fetchPcChange, &pc, fetchStop) {
			return
		}
	}
}

/// if we recieve a Pause while sending an Execute, throw away the Execute, don't block on it,
/// and Pause until Resume
/// @return whether to keep going. False => stop
func trySendExecute(execute chan<- ExecuteParam, 
	decodePause <-chan bool, 
	decodeResume <-chan bool, 
	params ExecuteParam, 
	decodeStop <-chan bool) bool {

	select {
	case execute <- params:
		return true
	case <-decodePause:
		<-decodeResume
		return true
	case <-decodeStop:
		return false
		// don't execute params, pause => pipeline flush
	}
}
func trySendDecodeFinished(decodeFinished chan<- bool, decodePause <-chan bool, decodeResume <-chan bool) {
	select {
	case decodeFinished<- true:
	case <-decodePause:
		<-decodeResume
		// don't send decodeFinished. Pause => pipeline flush
	}
}
func Decoder(decode chan []byte,
	execute chan<- ExecuteParam,
	decodePause <-chan bool,
	decodeResume <-chan bool,
	fetchFinished <-chan bool,
	decodeFinished chan<- bool,
	decodeStop <-chan bool) {

	for {
		select {
		case instruction := <-decode:
			op := OpCode(instruction[0] & 63) // 63 = 00111111
			if !isMem(op) {
				param1 := instruction[0]>>6 | instruction[1]<<2&63
				param2 := instruction[1]>>4 | instruction[2]<<4&63
				param3 := instruction[2] >> 2
				if(!trySendExecute(execute, decodePause, decodeResume, (ExecuteParams{op, []byte{param1, param2, param3}}), decodeStop)) {
					return
				}
				
			} else {
				param := instruction[0]>>6 | instruction[1]<<2&63
				memParam := uint16(instruction[1]>>4) | uint16(instruction[2])<<4
				params := ExecuteMemParams{op, param, memParam}
				if(!trySendExecute(execute, decodePause, decodeResume, params, decodeStop)) {
					return
				}
			}
		case <-decodePause:
			<-decodeResume
		case <-decodeStop:
			return
		case <-fetchFinished:
			trySendDecodeFinished(decodeFinished, decodePause, decodeResume)
		}
	}
}

func drainDecode(decode <-chan []byte, decodePause chan<- bool, decodeResume chan<- bool, fetchFinished <-chan bool) {
	decodePause <- true
	for {
		select {
		case <-decode:
		case <-fetchFinished:
		default:
			decodeResume<- true
			return
		}
	}
}
func drainExecute(execute <-chan ExecuteParam, decodeFinished <-chan bool) {
	for {
		select {
		case <-execute:
		case <-decodeFinished:
		default:
			return
	}
	}
}
/// sends on the Wait chan, while throwing away any executes we receive.
/// this is necessary, because the Fetcher listening for the Wait may be trying to write to decode,
/// while the Decoder is trying to write to Execute.
func sendWaitForPc(fetchWaitForPcChange chan<- bool, execute <-chan ExecuteParam, decodeFinished <-chan bool) {
	for {
		select {
		case <-execute:
		case <-decodeFinished:
		case fetchWaitForPcChange <- true:
			return
		}
	}
}
func Executor(cu *ControlUnit24bitPipelined,
	execute <-chan ExecuteParam,
	fetchWaitForPcChange chan<- bool,
	fetchPcChange chan<- int64,
	decodePause chan<- bool,
	decodeResume chan<- bool,
	decode <-chan []byte,
	fetchFinished <-chan bool,
	decodeFinished <-chan bool,
	fetchStop chan<- bool,
	decodeStop chan<- bool,
	finished chan<- bool) {

	for {
		select {
		case params := <-execute:
			if params.IsMem() {
				cu.ExecuteMem(params.Op(), params.Param(), params.MemParam())
				continue
			}
			jumpPos := cu.Execute(params.Op(), params.Params())
			if jumpPos == NoJump {
				continue
			}
			sendWaitForPc(fetchWaitForPcChange, execute, decodeFinished)
			drainDecode(decode, decodePause, decodeResume, fetchFinished)
			drainExecute(execute, decodeFinished)
			fetchPcChange <- jumpPos
		case <-decodeFinished:
			fetchStop<- true
			decodeStop<- true
			finished<- true
			return
		}
	}
}

func NewControlUnit24bitPipelined(indexRegisters uint, processingElements uint, memoryBytesPerElement uint) ControlUnit {
	var cu ControlUnit24bitPipelined
	cu.data = NewControlUnitData(indexRegisters, processingElements, memoryBytesPerElement)
	cu.FetchPcChangeChan = make(chan int64)
	cu.FetchWaitForPcChange = make(chan bool) ///< MUST be unbuffered, to force synchronisation
	cu.FetchFinished = make(chan bool)
	cu.DecodeChan = make(chan []byte)
	cu.DecodeFinished = make(chan bool)
	cu.DecodeStop = make(chan bool)
	cu.FetchStop = make(chan bool)
	cu.DecodePause = make(chan bool)
	cu.DecodeResume = make(chan bool)
	cu.ExecuteChan = make(chan ExecuteParam)
	cu.ExecuteStopChan = make(chan bool)
	cu.ExecuteFinishedChan = make(chan bool)
	cu.Finished = make(chan bool)
	return &cu
}

func (cu *ControlUnit24bitPipelined) Data() *ControlUnitData {
	return cu.data
}

/// @todo remove?
func (cu *ControlUnit24bitPipelined) PrintMachine() {
	fmt.Println("Machine: 24bit pipelined")
	cu.data.PrintMachine()
}

func (cu *ControlUnit24bitPipelined) RunProgram(program Program) error {
	pr, err := NewProgramReader24bitMem(program)
	if err != nil {
		return err
	}
	return cu.run(pr)
}
func (cu *ControlUnit24bitPipelined) Run(programFile string) error {
	pr, err := NewProgramReader24bit(programFile)
	if err != nil {
		return err
	}
	return cu.run(pr)
}

func (cu *ControlUnit24bitPipelined) run(pr ProgramReader) error {
	go Fetcher(pr, 
		cu.DecodeChan, 
		cu.FetchWaitForPcChange, 
		cu.FetchPcChangeChan,
		cu.FetchFinished,
		cu.FetchStop)
	go Decoder(cu.DecodeChan, 
		cu.ExecuteChan, 
		cu.DecodePause, 
		cu.DecodeResume,
		cu.FetchFinished,
		cu.DecodeFinished,
		cu.DecodeStop)
	go Executor(cu,
		cu.ExecuteChan,
		cu.FetchWaitForPcChange,
		cu.FetchPcChangeChan,
		cu.DecodePause,
		cu.DecodeResume,
		cu.DecodeChan,
		cu.FetchFinished,
		cu.DecodeFinished,
		cu.FetchStop,
		cu.DecodeStop,
		cu.Finished)

	<-cu.Finished
	return nil
}	

func (cu *ControlUnit24bitPipelined) ExecuteMem(instruction OpCode, param byte, memParam uint16) {
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
func (cu *ControlUnit24bitPipelined) Execute(instruction OpCode, params []byte) (jumpPos int64) {
	jumpPos = NoJump
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
		jumpPos = cu.Cmpx(params[0], params[1], params[2])
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
	return
}


//
// control instructions
//
func (cu *ControlUnit24bitPipelined) Ldx(index byte, a uint16) {
	//	fmt.Printf("ldx: cu.index[%d] = cu.data.Memory[%d] (%d)"
	cu.data.IndexRegister[index] = cu.data.Memory[a]
}
func (cu *ControlUnit24bitPipelined) Stx(index byte, a uint16) {
	cu.data.Memory[a] = cu.data.IndexRegister[index]
}
func (cu *ControlUnit24bitPipelined) Ldxi(index byte, a byte) {
	cu.data.IndexRegister[index] = int64(a)
}
func (cu *ControlUnit24bitPipelined) Incx(index byte, a byte) {
	cu.data.IndexRegister[index] += int64(a)
}
func (cu *ControlUnit24bitPipelined) Decx(index byte, a byte) {
	cu.data.IndexRegister[index] -= int64(a)
}
func (cu *ControlUnit24bitPipelined) Mulx(index byte, a byte) {
	cu.data.IndexRegister[index] *= int64(a)
}
func (cu *ControlUnit24bitPipelined) Cload(index uint16) {
	cu.data.ArithmeticRegister = cu.data.Memory[index]
}
func (cu *ControlUnit24bitPipelined) Cstore(index uint16) {
	cu.data.Memory[index] = cu.data.ArithmeticRegister
}

/// @todo fix this to take a larger jump (a). Byte only allows for 256 instructions. That's not a very big program
func (cu *ControlUnit24bitPipelined) Cmpx(index byte, ix2 byte, a byte) (jumpPos int64) {
	if cu.data.IndexRegister[index] < cu.data.IndexRegister[ix2] {
		jumpPos = int64(a)
	} else {
		jumpPos = NoJump
	}
	return
}

// control broadcast. Broadcasts the Control's Arithmetic Register to every PE's Routing Register
func (cu *ControlUnit24bitPipelined) Cbcast() {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].RoutingRegister = cu.data.ArithmeticRegister
	}
}

// Block until all PE's are done
func (cu *ControlUnit24bitPipelined) Barrier() {
	for i := 0; i != len(cu.data.PE); i++ {
		<-cu.data.Done
	}
}

//
// vector instructions
//
func (cu *ControlUnit24bitPipelined) Lod(a byte, idx byte) {
	//	fmt.Printf("PE-Lod %d + %d (%d)\n", a, cu.data.IndexRegister[idx], idx)
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Lod <- ByteTuple{a, byte(cu.data.IndexRegister[idx])} ///< @todo is this ok? Should we be loading the index register somewhere else?
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Sto(a byte, idx byte) {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Sto <- ByteTuple{a, byte(cu.data.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Add(a byte, idx byte) {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Add <- ByteTuple{a, byte(cu.data.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Sub(a byte, idx byte) {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Sub <- ByteTuple{a, byte(cu.data.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Mul(a byte, idx byte) {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Mul <- ByteTuple{a, byte(cu.data.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Div(a byte, idx byte) {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Div <- ByteTuple{a, byte(cu.data.IndexRegister[idx])}
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Bcast(idx byte) {
	idx = byte(cu.data.IndexRegister[idx]) ///< @todo is this ok? Should we be loading the index register here?
	for i, _ := range cu.data.PE {
		if !cu.data.PE[i].Enabled {
			continue
		}
		cu.data.PE[i].RoutingRegister = cu.data.PE[idx].RoutingRegister
	}
}
func (cu *ControlUnit24bitPipelined) Mov(from RegisterType, to RegisterType) {
	/// @todo remove this? speed for safety?
	if from == to {
		return
	}
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Mov <- ByteTuple{byte(from), byte(to)}
	}
	cu.Barrier()
}

func (cu *ControlUnit24bitPipelined) Radd() {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Radd <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Rsub() {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Rsub <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Rmul() {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Rmul <- true
	}
	cu.Barrier()
}
func (cu *ControlUnit24bitPipelined) Rdiv() {
	for i, _ := range cu.data.PE {
		cu.data.PE[i].Rdiv <- true
	}
	cu.Barrier()
}
