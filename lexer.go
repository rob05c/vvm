package main

import (
	"strings"
)

func LexProgram(source string) (Program, error) {
	//	aliases := make(map[string]int) // jump labels, EQUIV pseudo-ops, DATA positions, and BSS positions

	var program []byte
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		instruction, err := LexLine(line)
		if err != nil {
			return nil, err
		}
		program = append(program, instruction...)
	}
	return program, nil
}

/// @todo implement this
func LexLine(line string) ([]byte, error) {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return nil, nil
	}

	op := StringToInstruction(tokens[0])
	if op == isInvalid { // invalid op => assume pseudo-operation

	} else {

	}
	return nil, nil
}
