package main

type Program interface {
	PushMem(instruction OpCode, param byte, memParam uint16)
	Push(instruction OpCode, params []byte)
	Size() byte
	Save(file string) error
	DataOp(cu *ControlUnitData, data byte) (address uint16)
	At(index int64) []byte
}
