package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

const InstructionLength32bit = 4 ///< instructions are 4 bytes wide, or 32 bits

type Program32bit []byte

func NewProgram32bit() *Program32bit {
	p := make(Program32bit, 0)
	return &p
}

/// CU Memory addresses are 12 bits, so they're encoded a little differently
func (p *Program32bit) PushMem(instruction OpCode, param byte, memParam uint16) {
	*p = append(*p, byte(instruction))
	*p = append(*p, param)
	*p = append(*p, byte(memParam))
	*p = append(*p, byte(memParam >> 8))
}

/// Do NOT call this for CU Mem instructions - ldx, stx, cload, cstore. Call PushMem instead.
func (p *Program32bit) Push(instruction OpCode, params []byte) {
	if len(params) < InstructionLength32bit - 1 {
		panic("not enough params") /// @todo error?
	}
	*p = append(*p, byte(instruction))
	*p = append(*p, params[0])
	*p = append(*p, params[1])
	*p = append(*p, params[2])
}

// returns the number of instructions. Use for creating Labels and Jump positions
func (p Program32bit) Size() byte {
	return byte(len(p) / InstructionLength32bit)
}

func (p Program32bit) At(index int64) []byte {
	return p[index*InstructionLength32bit:index*InstructionLength32bit+InstructionLength32bit]
}

/// This doesn't really compile. The "compiling" to binary has already been done by the lexer
/// This just writes the byte array to a file
func (p Program32bit) Save(file string) error {
	return ioutil.WriteFile(file, p, 0xFFF)
}

func LoadProgram32bit(file string) (Program, error) {
	p, err := ioutil.ReadFile(file)
	pp := Program32bit(p)
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
var nextDataPos32bit int ///< @todo get rid of this, with the magic of FP
func (p *Program32bit) DataOp(cu *ControlUnitData, data byte) (address uint16) {
	// init next data position
	if nextDataPos32bit == 0 {
		bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)
		nextDataPos32bit = len(cu.PE) * bytesPerPe
		//		fmt.Printf("DataOp() Init nextDataPos: bytesperpe: %d pelen: %d pos: %d\n", bytesPerPe, len(cu.PE), nextDataPos) // debug
	}
	if nextDataPos32bit == len(cu.Memory) {
		panic("too much data, not enough memory") /// @todo handle error
	}
	if nextDataPos32bit > 65535 {
		fmt.Printf("Error: nextDataPos is greater than 16 bits: %d\n", nextDataPos32bit)
		panic("data address exceeds 16 bits") // @todo handle error. CU Memory addresses are 12 bits.
	}
	p.Push(isLdxi, []byte{0, data, 0})
	p.PushMem(isStx, 0, uint16(nextDataPos32bit))
	nextDataPos32bit++
	return uint16(nextDataPos32bit - 1) // return the value before it was incremented
}

type ProgramReader32bit os.File

func NewProgramReader32bit(file string) (ProgramReader, error) {
	f, err := os.Open(file)
	pr := (*ProgramReader32bit)(f)
	return pr, err
}

func (pr *ProgramReader32bit) ReadInstruction(num int64) ([]byte, error) {
	instruction := make([]byte, InstructionLength32bit, InstructionLength32bit)
	_, err := (*os.File)(pr).ReadAt(instruction, num*InstructionLength32bit)
	return instruction, err
}
