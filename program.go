package main

type Program interface {
	PushMem(instruction OpCode, param byte, memParam uint16)
	Push(instruction OpCode, params []byte)
	Size() int64 /// @todo fix Ldxi to take more than a byte. This means we're limited to 255-inst programs :(
	Save(file string) error
	DataOp(cu *ControlUnitData, data byte) (address uint16)
	At(index int64) []byte
}

type ProgramReader interface {
	ReadInstruction(num int64) ([]byte, error)
}
