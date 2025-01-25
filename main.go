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
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmdBuilder := types.NewCommandBuilder(ModeFileBased)

	args := os.Args[1:]
	if len(args) > 0 {
		inputFileName := args[0]
		commands, err := cmdBuilder.BuildCommands(ctx, inputFileName)
		if err != nil {
			log.Fatalf("failed to execute commands from input. Error: %v", err)
		}
		executeCommands(ctx, commands)
		return
	} else {
		cmdBuilder := types.NewCommandBuilder(ModeInteractive)
		startInteractiveMode(ctx, cmdBuilder)
	}
}

func executeCommands(ctx context.Context, commands []types.Commander) {
	for _, command := range commands {
		command.Execute(ctx)
	}
}

func startInteractiveMode(ctx context.Context, cb *types.CommandBuilder) {
	parkingLotCreated := false
	reader := bufio.NewReader(os.Stdin)

	for {
		input, _ := reader.ReadString('\n')
		tokens := strings.Split(input, " ")
		if len(tokens) < 1 {
			fmt.Println("invalid command.....")
			continue
		}

		trimmedToken := strings.TrimSpace(tokens[0])
		if strings.ToLower(trimmedToken) == "exit" {
			return
		}

		cmd := cb.ParseCommand(trimmedToken, tokens[1:]...)
		if cmd == nil {
			continue
		}

		if trimmedToken == types.TokenForCreateParkingLot {
			parkingLotCreated = true
		}

		if !parkingLotCreated && trimmedToken != types.TokenForCreateParkingLot {
			fmt.Printf("\ninvalid, please create a parking lot first\n\n")
			continue
		}

		cmd.Execute(ctx)
		fmt.Printf("\n\n")
	}
}
