package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ilivestrong/internal/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	cmdBuilder := types.CommandBuilder{}

	args := os.Args[1:]
	if len(args) > 0 {
		inputFileName := args[0]
		commands, err := cmdBuilder.BuildCommands(ctx, inputFileName)
		if err != nil {
			log.Fatalf("failed to execute commands from input. Error: %v", err)
		}
		executeCommands(ctx, commands)
		return
	}
}

func executeCommands(ctx context.Context, commands []types.Commander) {
	for _, command := range commands {
		command.Execute(ctx)
	}
}
