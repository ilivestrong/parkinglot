package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ilivestrong/internal/lib"
)

const (
	ModeInteractive = "interactive"
	ModeFileBased   = "filebased"
	CommandExit     = "exit"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	if len(os.Args) > 1 {
		inputFileName := os.Args[1]
		runFileBasedMode(ctx, inputFileName, os.Stdout)
	} else {
		runInteractiveMode(ctx, os.Stdin, os.Stdout)
	}
}

func runFileBasedMode(ctx context.Context, inputFileName string, output io.Writer) {
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	cmdBuilder := lib.NewCommandBuilder(ModeFileBased, writer)
	commands, err := cmdBuilder.BuildCommands(ctx, inputFileName)
	if err != nil {
		return
	}
	executeCommands(ctx, commands)
}

func runInteractiveMode(ctx context.Context, input io.Reader, output io.Writer) {
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	cmdBuilder := lib.NewCommandBuilder(ModeInteractive, writer)

	parkingLotCreated := false
	for {
		input, _ := reader.ReadString('\n')
		commandName, args := tokenize(input)
		if commandName == "" {
			writeToOutput(writer, "invalid command.....\n")
			continue
		}

		if strings.ToLower(commandName) == CommandExit {
			return
		}

		cmd := cmdBuilder.ParseCommand(commandName, args...)
		if cmd == nil {
			continue
		}

		if commandName == lib.TokenForCreateParkingLot {
			parkingLotCreated = true
		} else if !parkingLotCreated {
			writeToOutput(writer, "\nPlease create a parking lot first\n\n")
			continue
		}

		if err := cmd.Execute(ctx); err != nil {
			return
		}
		fmt.Printf("\n\n")
	}
}

func executeCommands(ctx context.Context, commands []lib.Commander) {
	for _, command := range commands {
		command.Execute(ctx)
	}
}

func tokenize(input string) (string, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}
	tokens := strings.Fields(input)
	return tokens[0], tokens[1:]
}

func writeToOutput(writer *bufio.Writer, message string) {
	fmt.Fprintf(writer, "%s", message)
	writer.Flush()
}
