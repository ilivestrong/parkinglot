package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunInteractiveMode(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name: "Create parking lot and park two cars",
			input: `create_parking_lot 6
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2`,
		},
		{
			name: "Create parking lot and park three cars, leave slot 2 and exit",
			input: `create_parking_lot 4
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		park KA-01-HH-9531 Red
		leave 2
		exit
		`,
			expectedOutput: `Created a parking lot with 4 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Slot number 2 is free`,
		},
		{
			name: "Create parking lot and exit without parking",
			input: `create_parking_lot 4
				exit
				`,
			expectedOutput: `Created a parking lot with 4 slots`,
		},
		{
			name: "Invalid command",
			input: `invalid_command
		exit
		`,
			expectedOutput: `invalid command: invalid_command, skipping...`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the context
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Prepare input and output
			input := strings.NewReader(tt.input)
			var output bytes.Buffer

			// Call the function
			runInteractiveMode(ctx, input, &output)

			actual := strings.ReplaceAll(output.String(), "\n", "")

			if actual != tt.expectedOutput {
				t.Errorf("output did not match expected. Got:\n%s\nExpected:\n%s", actual, tt.expectedOutput)
			}
		})
	}
}

// package main

// import (
// 	"bytes"
// 	"context"
// 	"strings"
// 	"testing"
// 	"time"
// )

// func TestRunInteractiveMode(t *testing.T) {
// 	// Prepare the context
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	// Simulate user input
// 	userInput := `create_parking_lot 6
// park KA-01-HH-1234 White
// park KA-01-HH-9999 White
// exit
// `
// 	input := strings.NewReader(userInput)

// 	// Capture output
// 	var output bytes.Buffer

// 	// Call the function
// 	runInteractiveMode(ctx, input, &output)

// 	actual := strings.ReplaceAll(output.String(), "\r\n", "\n")

// 	// Expected output
// 	expectedOutput := `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2`

// 	// Assert output
// 	if actual != expectedOutput {
// 		t.Errorf("output did not match expected. Got:\n%s\nExpected:\n%s", output.String(), expectedOutput)
// 	}
// }
