package main

import (
	"errors"
	//	"fmt"
	"strconv"
	"strings"
)

/// NOTE Programs must be run on the same CU they are compiled for.
///      That is, with the same registers, elements, and memory.
///      Otherwise, memory layouts will not line up and the program will explode.
func LexProgram(cu *ControlUnit, source string) (Program, error) {
	lines := RemoveBlanks(strings.Split(source, "\n"))
	lines, program, err := ParsePseudoOperations(cu, lines)
	if err != nil {
		return nil, err
	}
	lines, labels := ParseLabels(lines)
	return ReplaceLabels(lines, labels, program)

	//	input := strings.Join(lines, "\n") // debug
	//	fmt.Println("ppo input: X" + input + "X")

	//	output := strings.Join(lines, "\n")         // debug
	//	fmt.Println("ppo output: X" + output + "X") // debug

	//	output = strings.Join(lines, "\n")         // debug
	//	fmt.Println("pl output: X" + output + "X") // debug

	//	fmt.Print("Labels: ")
	//	fmt.Println(labels)
}

func RemoveBlanks(lines []string) []string {
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
		if len(lines[i]) == 0 {
			lines = append(lines[:i], lines[i+1:]...)
			i--
		}
	}
	return lines
}

/// @todo handle movR toA etc.
func ReplaceLabels(lines []string, labels map[string]int, program Program) (Program, error) {
	//	instructionSize := 3 ///< @todo don't hardcode instruction size here. Magic numbers bad!
	realLabels := make(map[string]int)
	for i, _ := range lines {
		for key, val := range labels {
			if val == i {
				realLabels[key] = int(program.Size()) + i
				//				fmt.Printf("RealLabel %s is %d + %d + %d = %d\n", key, program.Size(), i, instructionSize, realLabels[key])
			}
		}
	}
	for i, _ := range lines {
		var params []int

		tokens := strings.Fields(lines[i])
		if len(tokens) == 0 {
			return nil, errors.New("malformed line a " + strconv.Itoa(i))
		}

		op := StringToInstruction(strings.ToLower(tokens[0]))
		if op == isInvalid {
			return nil, errors.New("malformed line b " + strconv.Itoa(i))
		}

		tokens = tokens[1:]
		for j, _ := range tokens {

			subtokens := strings.Split(tokens[j], ",")
			for k, _ := range subtokens {
				subtokens[k] = strings.ToLower(subtokens[k])
				for key, val := range realLabels {
					if subtokens[k] == key {
						subtokens[k] = strconv.Itoa(val)
						//						fmt.Printf("Label Usage Replaced at line %d with %d\n", i, val)
					}
				}

				val, err := strconv.Atoi(subtokens[k])
				if err != nil {
					return nil, errors.New("malformed line c " + strconv.Itoa(i) + " : " + subtokens[k])
				}
				params = append(params, val)
			}
		}

		for len(params) < 3 {
			params = append(params, 0)
		}

		if isMem(op) {
			program.PushMem(op, byte(params[0]), uint16(params[1]))
		} else {
			var bytes []byte
			bytes = append(bytes, byte(params[0]))
			bytes = append(bytes, byte(params[1]))
			bytes = append(bytes, byte(params[2]))
			program.Push(op, bytes)
		}
	}
	return program, nil
}

func ParseLabels(lines []string) (parsed []string, labels map[string]int) {
	labels = make(map[string]int)
	for i := 0; i < len(lines); i++ {
		for labelEnd := strings.Index(lines[i], ":"); labelEnd != -1; labelEnd = strings.Index(lines[i], ":") {
			labelStart := strings.LastIndex(lines[i][:labelEnd], "\n")
			if labelStart == -1 {
				labelStart = 0
			}
			label := lines[i][labelStart:labelEnd]
			label = strings.TrimSpace(label)
			//			fmt.Printf("Found Label: %s at %d\n", label, i)
			labels[label] = i
			lines[i] = strings.TrimSpace(lines[i][labelEnd+1:])

			if len(lines[i]) == 0 {
				lines = append(lines[:i], lines[i+1:]...) // if the line only had this label, remove the blank line
				i--
			}

			//			fmt.Printf("Label Line Replaced: X%sX\n", lines[i])
		}
	}
	return RemoveBlanks(lines), labels
}

/// @todo accomodate BSS matrices larger than len(cu.PE)
func ParsePseudoOperations(cu *ControlUnit, lines []string) (parsed []string, program Program, err error) {
	bytesPerPe := len(cu.Memory) / (len(cu.PE) + 1)

	data := make(map[string]int) // map[alias] cu_memory_location
	//	equiv := make(map[string]int) //map[alias] constant (usually to a CU IndexRegister location)
	//	bss := make(map[string]int) //map[alias] pe_memory_location

	var lastLine int
	var pprogram PseudoProgram

	nextBssLocation := 0

	for i, line := range lines {
		lastLine = i
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		op := StringToInstruction(tokens[0])
		if op != isInvalid { // valid op means we're done with pseudo-ops and have reached real instructions
			break
		}

		if len(tokens) < 3 {
			return nil, nil, errors.New("malformed line d " + strconv.Itoa(i))
		}

		alias := strings.ToLower(tokens[0])
		opType := strings.ToLower(tokens[1])
		strVal := tokens[2]

		switch opType {
		case "data":
			val, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, nil, errors.New("malformed line e " + strconv.Itoa(i))
			}
			location := pprogram.DataOp(cu, byte(val))
			data[alias] = int(location)
		case "equiv":
			val, err := strconv.Atoi(strVal)
			if err != nil {
				return nil, nil, errors.New("malformed line f " + strconv.Itoa(i))
			}
			data[alias] = val
		case "bss":
			vals := strings.Split(strVal, "x")
			if len(vals) != 2 {
				return nil, nil, errors.New("malformed line g " + strconv.Itoa(i))
			}
			width, err := strconv.Atoi(vals[0])
			if err != nil {
				return nil, nil, errors.New("malformed line h " + strconv.Itoa(i))
			}
			height, err := strconv.Atoi(vals[1])
			if err != nil {
				return nil, nil, errors.New("malformed line i " + strconv.Itoa(i))
			}
			if width > len(cu.PE) {
				return nil, nil, errors.New("line " + strconv.Itoa(i) + " exceeds number of Vector Processing Elements") /// @todo accomodate BSS matrices wider than len(cu.PE)
			}
			if height+nextBssLocation > bytesPerPe {
				//				fmt.Printf("Error exceeds width: %d, nbss: %d, bytesPerPe: %d\n", width, nextBssLocation, bytesPerPe)
				return nil, nil, errors.New("line " + strconv.Itoa(i) + " exceeds memory of Vector Processing Elements")
			}

			data[alias] = nextBssLocation
			nextBssLocation += height

		default:
			return nil, nil, errors.New("malformed line j " + strconv.Itoa(i))
		}

	}
	/*
		fmt.Println("ParsedPseudoOps:")
		fmt.Print("Lines: ")
		fmt.Println(lines)
		fmt.Print("linesPostPseu: ")
		fmt.Println(lines[lastLine:])
		fmt.Print("aliases: ")
		fmt.Println(data)
	*/
	lines = ReplacePseudoOpAliases(lines[lastLine:], data)
	return lines, Program(pprogram), nil
}

func ReplacePseudoOpAliases(lines []string, aliases map[string]int) []string {
	for i, line := range lines {
		tokens := strings.Fields(line)
		for j, token := range tokens {
			subtokens := strings.Split(token, ",")
			for k, _ := range subtokens {
				subtokens[k] = strings.ToLower(subtokens[k])
				if val, ok := aliases[subtokens[k]]; ok {
					// alias found!
					subtokens[k] = strconv.Itoa(val)
					tokens[j] = strings.Join(subtokens, ",")
					lines[i] = strings.Join(tokens, " ")
				}
			}
		}
	}
	return lines
}
