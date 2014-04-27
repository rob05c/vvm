package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"io"
)

const InstructionLength24bit = 3 ///< instructions are 3 bytes wide, or 24 bits

type Program24bit []byte

func NewProgram24bit() *Program24bit {
	p := make(Program24bit, 0)
	return &p
}

/// CU Memory addresses are 12 bits, so they're encoded a little differently
func (p *Program24bit) PushMem(instruction OpCode, param byte, memParam uint16) {
	byte1 := byte(instruction) | param<<6
	byte2 := param>>2 | byte(memParam)<<4
	byte3 := byte(memParam >> 4)
	*p = append(*p, byte1)
	*p = append(*p, byte2)
	*p = append(*p, byte3)
}

/// Do NOT call this for CU Mem instructions - ldx, stx, cload, cstore. Call PushMem instead.
func (p *Program24bit) Push(instruction OpCode, params []byte) {
	byte1 := byte(instruction) | params[0]<<6
	byte2 := params[0]>>2 | params[1]<<4
	byte3 := params[1]>>4 | params[2]<<2
	*p = append(*p, byte1)
	*p = append(*p, byte2)
	*p = append(*p, byte3)
}

// returns the number of instructions. Use for creating Labels and Jump positions
func (p Program24bit) Size() int64 {
	return int64(len(p) / 3)
}

func (p Program24bit) At(index int64) []byte {
	return p[index*InstructionLength24bit:index*InstructionLength24bit+InstructionLength24bit]
}

/// This doesn't really compile. The "compiling" to binary has already been done by the lexer
/// This just writes the byte array to a file
func (p Program24bit) Save(file string) error {
	return ioutil.WriteFile(file, p, 0xFFF)
}

func LoadProgram24bit(file string) (Program, error) {
	p, err := ioutil.ReadFile(file)
	pp := Program24bit(p)
	return Program(&pp), err
}


/// Data Pseudo-Operation
///
/// This puts the given data in a memory location, returns the address for that location,
/// and the operations necessary to store the data there.
/// The operations MUST be executed before any ops which reference the data.
/// It is HIGHLY recommended to execute all DataOps first.
///
/// @param cu necessary to get the initial data position, and to ensure we haven't exceeded memory
var nextDataPos int

func (p *Program24bit) DataOp(cu *ControlUnitData, data byte) (address uint16) {
	// init next data position
	if nextDataPos == 0 {
		bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
		nextDataPos = len(cu.PE) * bytesPerPe
		//		fmt.Printf("DataOp() Init nextDataPos: bytesperpe: %d pelen: %d pos: %d\n", bytesPerPe, len(cu.PE), nextDataPos) // debug
	}
	if nextDataPos == len(cu.Memory) {
		panic("too much data, not enough memory") /// @todo handle error
	}
	if nextDataPos > 4095 {
		fmt.Printf("Error: nextDataPos is greater than 12 bits: %d\n", nextDataPos)
		panic("data address exceeds 12 bits") // @todo handle error. CU Memory addresses are 12 bits.
	}
	p.Push(isLdxi, []byte{0, data, 0})
	p.PushMem(isStx, 0, uint16(nextDataPos))
	nextDataPos++
	return uint16(nextDataPos - 1) // return the value before it was incremented
}

type ProgramReader24bit os.File

func NewProgramReader24bit(file string) (ProgramReader, error) {
	f, err := os.Open(file)
	pr := (*ProgramReader24bit)(f)
	return pr, err
}

func (pr *ProgramReader24bit) ReadInstruction(num int64) ([]byte, error) {
	instruction := make([]byte, InstructionLength24bit, InstructionLength24bit)
	_, err := (*os.File)(pr).ReadAt(instruction, num*InstructionLength24bit)
	return instruction, err
}

type ProgramReader24bitMem struct {
	Program
}

func NewProgramReader24bitMem(program Program) (ProgramReader, error) {
	pr := ProgramReader24bitMem{program}
	return &pr, nil
}

func (pr *ProgramReader24bitMem) ReadInstruction(num int64) ([]byte, error) {
	if num > pr.Program.Size() {
		return nil, io.EOF
	}
	return pr.Program.At(num), nil
}
