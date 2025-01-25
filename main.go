package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ilivestrong/internal/types"
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
		runFileBasedMode(ctx, inputFileName)
	} else {
		runInteractiveMode(ctx)
	}
}

func runFileBasedMode(ctx context.Context, inputFileName string) {
	cmdBuilder := types.NewCommandBuilder(ModeFileBased)
	commands, err := cmdBuilder.BuildCommands(ctx, inputFileName)
	if err != nil {
		log.Fatalf("failed to execute commands from input. Error: %v", err)
	}
	executeCommands(ctx, commands)
}

func runInteractiveMode(ctx context.Context) {
	cmdBuilder := types.NewCommandBuilder(ModeInteractive)
	reader := bufio.NewReader(os.Stdin)

	parkingLotCreated := false
	for {
		input, _ := reader.ReadString('\n')
		commandName, args := tokenize(input)
		if commandName == "" {
			fmt.Println("invalid command.....")
			continue
		}

		if strings.ToLower(commandName) == CommandExit {
			return
		}

		cmd := cmdBuilder.ParseCommand(commandName, args...)
		if cmd == nil {
			continue
		}

		if commandName == types.TokenForCreateParkingLot {
			parkingLotCreated = true
		} else if !parkingLotCreated {
			fmt.Printf("\nPlease create a parking lot first\n\n")
			continue
		}

		cmd.Execute(ctx)
		fmt.Printf("\n\n")
	}
}

func executeCommands(ctx context.Context, commands []types.Commander) {
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
