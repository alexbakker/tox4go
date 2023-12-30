package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alexbakker/tox4go/state"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: state-tool -s/-d\n")
		return
	}

	var serialize bool

	switch os.Args[1] {
	case "-s":
		serialize = true
	case "-d":
		serialize = false
	default:
		fmt.Printf("error: unknown option '%s'\n", os.Args[1])
		return
	}

	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("error reading input: %s\n", err.Error())
		return
	}

	if serialize {
		stateFile := state.State{}
		err = stateFile.UnmarshalBinary(inputBytes)

		if err != nil {
			fmt.Printf("error parsing input: %s\n", err.Error())
			return
		}

		stateFileAlias := stateAlias(stateFile)
		output, err := json.MarshalIndent(&stateFileAlias, "", "    ")

		if err != nil {
			fmt.Printf("error marshalling to json: %s\n", err.Error())
			return
		}

		os.Stdout.Write([]byte(output))
	} else {
		stateFileAlias := stateAlias(state.State{})
		err := json.Unmarshal(inputBytes, &stateFileAlias)

		if err != nil {
			fmt.Printf("error parsing input: %s\n", err.Error())
			return
		}

		stateFile := state.State(stateFileAlias)
		output, err := stateFile.MarshalBinary()

		if err != nil {
			fmt.Printf("error marshalling to .tox: %s\n", err.Error())
			return
		}

		os.Stdout.Write(output)
	}
}
