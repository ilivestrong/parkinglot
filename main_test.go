package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ilivestrong/internal/lib"
)

func TestRunFileBasedMode(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		expectedOutput string
	}{
		{
			name: "Filebased - create a slot of 6, park 2 cars",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 White`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2`,
		},
		{
			name: "Filebased - create a slot of 6, park 2 cars, leave slot 2",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 White
			leave 2`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Slot number 2 is free`,
		},
		{
			name: "Filebased - create a slot of 6, park 2 cars, leave slot 2, park new car and get slot 2",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 White
			leave 2
			park KA-01-HH-4444 Red`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Slot number 2 is free
			Allocated slot number: 2`,
		},
		{
			name: "Filebased - create a slot of 6, park 5 cars, get registration numbers for Red cars",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 Red
			park KA-01-HH-1234 White
			park KA-01-HH-8888 Red
			park KA-01-HH-4444 Red
			registration_numbers_for_cars_with_color Red`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Allocated slot number: 3
			Allocated slot number: 4
			Allocated slot number: 5
			KA-01-HH-9999, KA-01-HH-8888, KA-01-HH-4444`,
		},
		{
			name: "Filebased - create a slot of 30k, get error",
			fileContent: `create_parking_lot 30000
			park KA-01-HH-1234 White
			park KA-01-HH-9999 Red
			park KA-01-HH-1234 White
			park KA-01-HH-8888 Red
			park KA-01-HH-4444 Red
			registration_numbers_for_cars_with_color Red`,
			expectedOutput: `cannot create :30000 slots. max slots available: 20000`,
		},
		{
			name: "Filebased - create a slot of 6, park 5 cars, get slot numbers for White cars",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 Red
			park KA-01-HH-1235 White
			park KA-01-HH-8888 Red
			park KA-01-HH-4444 Red
			slot_numbers_for_cars_with_color White`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Allocated slot number: 3
			Allocated slot number: 4
			Allocated slot number: 5
			1, 3`,
		},
		{
			name: "Filebased - create a slot of 6, park 5 cars, get slot numbers for registration number: KA-01-HH-999",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 Red
			park KA-01-HH-1235 White
			park KA-01-HH-8888 Red
			park KA-01-HH-4444 Red
			slot_number_for_registration_number KA-01-HH-9999`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Allocated slot number: 3
			Allocated slot number: 4
			Allocated slot number: 52`,
		},
		{
			name: "Filebased - create a slot of 6, park 5 cars, get slot numbers for invalid registration number",
			fileContent: `create_parking_lot 6
			park KA-01-HH-1234 White
			park KA-01-HH-9999 Red
			park KA-01-HH-1235 White
			park KA-01-HH-8888 Red
			park KA-01-HH-4444 Red
			slot_number_for_registration_number invalid_11`,
			expectedOutput: `Created a parking lot with 6 slots
		    Allocated slot number: 1
		    Allocated slot number: 2
			Allocated slot number: 3
			Allocated slot number: 4
			Allocated slot number: 5Not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tempFile, err := os.CreateTemp("", "testfile_*.txt")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			content := strings.ReplaceAll(tt.fileContent, "\t", "")
			if _, err := tempFile.WriteString(content); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tempFile.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var output bytes.Buffer

			// Run the function
			runFileBasedMode(ctx, tempFile.Name(), &output)

			// Normalize output: Trim spaces, replace \r\n with \n, and remove leading spaces from each line
			actualOutput := normalizeOutput(output.String())
			expectedOutput := strings.ReplaceAll(tt.expectedOutput, "\t", "")
			expectedOutput = normalizeOutput(expectedOutput)

			if actualOutput != expectedOutput {
				t.Errorf("output did not match expected.\nGot:\n%s\nExpected:\n%s", actualOutput, expectedOutput)
			}
		})
	}
}

// normalizeOutput trims leading/trailing spaces and normalizes line breaks
// Also removes leading spaces from each line for consistent formatting
func normalizeOutput(input string) string {
	// Normalize line breaks
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.TrimSpace(input)

	// Remove leading spaces for each line
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeft(line, " ") // Trim leading spaces from each line
	}
	return strings.Join(lines, "\n")
}

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
			name: "Create parking lot with 4 slots and exit without parking",
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
		{
			name: "Create parking lot with 10 slots and park three cars, get status and exit",
			input: `create_parking_lot 10
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		park KA-01-HH-9531 Red
		status
		exit
		`,
			expectedOutput: `Created a parking lot with 10 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Slot No.   Registration No      Color     1          KA-01-HH-1234        White     2          KA-01-HH-9999        White     3          KA-01-HH-9531        Red       `,
		},
		{
			name: "Create parking lot with 30K slots",
			input: `create_parking_lot 30000
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		exit
		`,
			expectedOutput: fmt.Sprintf(`cannot create :30000 slots. %s`, lib.ErrMaxSlotExceeded),
		},
		{
			name: "Create parking lot of 2 cars and try to park 3 cars, get parking lot full error",
			input: `create_parking_lot 2
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		park KA-01-HH-8888 White
		exit
		`,
			expectedOutput: `Created a parking lot with 2 slotsAllocated slot number: 1Allocated slot number: 2Sorry, parking lot is full`,
		},
		{
			name: "Create parking lot of 6 cars and park 4 cars, get registration number for white color",
			input: `create_parking_lot 6
		park KA-01-HH-1234 White
		park KA-01-HH-9999 White
		park KA-01-HH-8888 Red
		park KA-01-GG-1121 White
		registration_numbers_for_cars_with_color White
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Allocated slot number: 4KA-01-HH-1234, KA-01-HH-9999, KA-01-GG-1121`,
		},
		{
			name: "Create parking lot of 6 cars and park 4 cars, get slot number for Red color",
			input: `create_parking_lot 6
		park KA-01-HH-1234 Red
		park KA-01-HH-9999 White
		park KA-01-HH-8888 Red
		park KA-01-GG-1121 White
		slot_numbers_for_cars_with_color Red
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Allocated slot number: 41, 3`,
		},
		{
			name: "Create parking lot of 6 cars and park 4 cars, get slot number for registration no - KA-01-HH-8888",
			input: `create_parking_lot 6
		park KA-01-HH-1234 Red
		park KA-01-HH-9999 White
		park KA-01-HH-8888 Red
		park KA-01-GG-1121 White
		slot_number_for_registration_number KA-01-HH-8888
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Allocated slot number: 43`,
		},
		{
			name: "Create parking lot of 6 cars and park 4 cars, get slot number for invalid registration no - invalid_1111",
			input: `create_parking_lot 6
		park KA-01-HH-1234 Red
		park KA-01-HH-9999 White
		park KA-01-HH-8888 Red
		park KA-01-GG-1121 White
		slot_number_for_registration_number invalid_1111
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Allocated slot number: 4Not found`,
		},
		{
			name: "Create parking lot of 6 cars and park 4 cars, leave slot 3, park 1 car, get last available slot - 3",
			input: `create_parking_lot 6
		park KA-01-HH-1234 Red
		park KA-01-HH-9999 White
		park KA-01-HH-8888 Red
		park KA-01-GG-1121 White
		leave 3
		park KA-01-KK-4532 Blue
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsAllocated slot number: 1Allocated slot number: 2Allocated slot number: 3Allocated slot number: 4Slot number 3 is freeAllocated slot number: 3`,
		},
		{
			name: "Create parking lot of 3 cars, try to park 1 car with only 0 argument provided",
			input: `create_parking_lot 6
		park 
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsargs missing for command: park`,
		},
		{
			name: "Create parking lot of 3 cars, try to park 1 car with only 1 argument provided",
			input: `create_parking_lot 6
		park 
		exit
		`,
			expectedOutput: `Created a parking lot with 6 slotsargs missing for command: park`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the context
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
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
