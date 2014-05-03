///
/// @todo split this into multiple files
///
package main

type RegisterType int64

const (
	peIndex = iota
	peRouting
	peArithmetic
)

type ByteTuple struct {
	First  byte
	Second byte
}

type OpCode byte

const (
	// control instructions
	isLdx OpCode = iota
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
	isMov // debug: 18
	isRadd
	isRsub
	isRmul
	isRdiv

	isInvalid OpCode = ^OpCode(0)
)

func StringToInstruction(s string) OpCode {
	switch s {
	case "ldx":
		return isLdx
	case "stx":
		return isStx
	case "ldxi":
		return isLdxi
	case "incx":
		return isIncx
	case "decx":
		return isDecx
	case "mulx":
		return isMulx
	case "cload":
		return isCload
	case "cstore":
		return isCstore
	case "cmpx":
		return isCmpx
	case "cbcast":
		return isCbcast
	case "lod":
		return isLod
	case "sto":
		return isSto
	case "add":
		return isAdd
	case "sub":
		return isSub
	case "mul":
		return isMul
	case "div":
		return isDiv
	case "bcast":
		return isBcast
	case "mov":
		return isMov
	case "radd":
		return isRadd
	case "rsub":
		return isSub
	case "rmul":
		return isRmul
	case "rdiv":
		return isRdiv
	}
	return isInvalid
}

func (i OpCode) String() string {
	switch i {
	case isLdx:
		return "ldx"
	case isStx:
		return "stx"
	case isLdxi:
		return "ldxi"
	case isIncx:
		return "incx"
	case isDecx:
		return "decx"
	case isMulx:
		return "mulx"
	case isCload:
		return "cload"
	case isCstore:
		return "cstore"
	case isCmpx:
		return "cmpx"
	case isCbcast:
		return "cbcast"
	case isLod:
		return "lod"
	case isSto:
		return "sto"
	case isAdd:
		return "add"
	case isSub:
		return "sub"
	case isMul:
		return "mul"
	case isDiv:
		return "div"
	case isBcast:
		return "bcast"
	case isMov:
		return "mov"
	case isRadd:
		return "radd"
	case isRsub:
		return "rsub"
	case isRmul:
		return "rmul"
	case isRdiv:
		return "rdiv"
	}
	return "NUL"
}

var InstructionParams = map[OpCode]byte{
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

/// @return whether the given instruction is a CU Memory instruction, i.e. using a 12-bit memory address
func isMem(i OpCode) bool {
	return i == isLdx || i == isStx || i == isCload || i == isCstore
}
