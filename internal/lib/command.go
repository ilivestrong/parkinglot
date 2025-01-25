package lib

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	pm "github.com/ilivestrong/internal/lib/parking_manager"
)

const (
	TokenForCreateParkingLot            = "create_parking_lot"
	TokenForPark                        = "park"
	TokenForLeave                       = "leave"
	TokenForStatus                      = "status"
	TokenForQueryRegistrationNoByColor  = "registration_numbers_for_cars_with_color"
	TokenForQuerySlotNoByRegistrationNo = "slot_number_for_registration_number"
	TokenForQuerySlotNoByColor          = "slot_numbers_for_cars_with_color"

	MaxNumberOfSlots = 20000
)

var (
	ErrCreateParkingLotCommandMissing = errors.New("invalid input file, require 'create_parking_lot' as the first command")
	ErrInvalidCreateParkingLotCommand = errors.New("invalid 'create_parking_lot' command")
	ErrInvalidInputFile               = errors.New("invalid input file")
)

var (
	parkingLot       *pm.ParkingLot
	commandsWithArgs map[string]struct{} = map[string]struct{}{
		TokenForCreateParkingLot:            {},
		TokenForPark:                        {},
		TokenForLeave:                       {},
		TokenForQueryRegistrationNoByColor:  {},
		TokenForQuerySlotNoByRegistrationNo: {},
	}
	newlineOrNothing = "\n"
)

type (
	Commander interface {
		Execute(ctx context.Context)
	}

	CreateParkingLotCommand struct {
		capacity int
		owriter  *bufio.Writer
	}
	ParkCommand struct {
		vehicle *pm.Vehicle
		owriter *bufio.Writer
	}
	LeaveCommand struct {
		slot    int
		owriter *bufio.Writer
	}
	StatusCommand struct {
		owriter *bufio.Writer
	}
	QueryRegistrationNoByColorCommand struct {
		color   string
		owriter *bufio.Writer
	}
	QuerySlotNoByRegistrationNoCommand struct {
		registrationNo string
		owriter        *bufio.Writer
	}
	QuerySlotNoByColorCommand struct {
		color   string
		owriter *bufio.Writer
	}

	CommandBuilder struct {
		owriter *bufio.Writer
	}
)

/*
Instantiates a command builder object.

It allows to parse command(s) individually or in bulk.
*/
func NewCommandBuilder(mode string, oWriter *bufio.Writer) *CommandBuilder {
	if mode == "interactive" {
		newlineOrNothing = ""
	}
	return &CommandBuilder{owriter: oWriter}
}

/*
Parses command tokens and their args and returns a concrete Command object.

The returned command object is used to execute the command on-demand.
*/
func (cb *CommandBuilder) ParseCommand(commandName string, args ...string) Commander {
	if _, ok := commandsWithArgs[commandName]; ok && len(args) == 0 {
		writeToOutput(cb.owriter, fmt.Sprintf("\nargs missing for command: %s", commandName))
		return nil
	}

	switch commandName {
	case TokenForCreateParkingLot:
		capacity, err := strconv.Atoi(args[0])
		if err != nil {
			writeToOutput(cb.owriter, fmt.Sprintf("\ninvalid args provided for command: %s", commandName))
		}
		return &CreateParkingLotCommand{
			capacity: capacity,
			owriter:  cb.owriter,
		}
	case TokenForPark:
		return &ParkCommand{
			vehicle: pm.NewVehicle(args[0], args[1]),
			owriter: cb.owriter,
		}
	case TokenForLeave:
		slot, _ := strconv.Atoi(args[0])
		return &LeaveCommand{
			slot:    slot,
			owriter: cb.owriter,
		}
	case TokenForStatus:
		return &StatusCommand{owriter: cb.owriter}
	case TokenForQueryRegistrationNoByColor:
		return &QueryRegistrationNoByColorCommand{
			color:   args[0],
			owriter: cb.owriter,
		}
	case TokenForQuerySlotNoByRegistrationNo:
		return &QuerySlotNoByRegistrationNoCommand{
			registrationNo: args[0],
			owriter:        cb.owriter,
		}
	case TokenForQuerySlotNoByColor:
		return &QuerySlotNoByColorCommand{
			color:   args[0],
			owriter: cb.owriter,
		}
	case "":
		writeToOutput(cb.owriter, "\nno command provided \n\n")
		return nil
	default:
		writeToOutput(cb.owriter, fmt.Sprintf("\n invalid command: %s, skipping...\n\n", commandName))
		return nil
	}
}

/*
Used with file based mode and accepts an input file containing command instructions.

Reads input line by line (stream mode) to be memory efficient.
Parse each token in the file and instantites a command object.

Ultimately aggregates all those command instances and returns as a slice of interfaces.
*/
func (cb *CommandBuilder) BuildCommands(ctx context.Context, inputFileName string) ([]Commander, error) {
	file, err := os.Open(inputFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	tokens := strings.Split(scanner.Text(), " ")
	if tokens[0] != TokenForCreateParkingLot {
		/*
			Assumption: The first command must be create_parking_lot command.
			Without a parking lot, no command/operation would make sense and allowed.
		*/
		log.Fatal(ErrCreateParkingLotCommandMissing) // cannot proceed, panic!
	}

	// intilialise a parking lot
	cb.ParseCommand(TokenForCreateParkingLot, tokens[1]).Execute(ctx)
	// parkingLot = cmd.(*ParkingLot)

	// build rest of the commands
	commands := make([]Commander, 0)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")

		cmd := cb.ParseCommand(tokens[0], tokens[1:]...)
		if cmd == nil {
			continue
		}
		commands = append(commands, cmd)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func (cplCmd *CreateParkingLotCommand) Execute(ctx context.Context) {
	if cplCmd.capacity > MaxNumberOfSlots {
		writeToOutput(cplCmd.owriter, fmt.Sprintf("cannot create :%d slots. Maximum allowed is: %d", cplCmd.capacity, MaxNumberOfSlots))
		return
	}

	if cplCmd.capacity <= 0 {
		writeToOutput(cplCmd.owriter, fmt.Sprintf("invalid slot number: %d", cplCmd.capacity))
		return
	}

	writeToOutput(cplCmd.owriter, fmt.Sprintf("Created a parking lot with %d slots", cplCmd.capacity))
	parkingLot = pm.NewParkingLot(cplCmd.capacity)
}
func (parkCmd *ParkCommand) Execute(ctx context.Context) {
	if len(parkingLot.GetAvailableSlots()) == 0 {
		writeToOutput(parkCmd.owriter, newlineOrNothing+"Sorry, parking lot is full")
		return
	}

	slot := int(parkingLot.GetAvailableSlots()[0])
	parkingLot.UpdateAvailableSlots(parkingLot.GetAvailableSlots()[1:])
	parkingLot.UpdateOccupiedSlot(slot, parkCmd.vehicle)
	parkingLot.UpdateSlotByRegistrationNo(parkCmd.vehicle.GetRegistrationNo(), slot)
	parkingLot.UpdateVehiclesByColor(parkCmd.vehicle.GetColor(), parkCmd.vehicle)

	writeToOutput(parkCmd.owriter, fmt.Sprintf(newlineOrNothing+"Allocated slot number: %d", slot))
}

func (leaveCmd *LeaveCommand) Execute(ctx context.Context) {
	vehicle, exists := parkingLot.GetOccupiedSlots()[leaveCmd.slot]
	if !exists {
		writeToOutput(leaveCmd.owriter, fmt.Sprintf("slot %d is not occupied", leaveCmd.slot))
		return
	}

	parkingLot.RemoveSlotFromOccupiedSlots(leaveCmd.slot)
	parkingLot.RemoveRegistrationNoFromOccupiedSlots(vehicle.GetRegistrationNo())
	parkingLot.RemoveVehicleFromColorToVehicleMapping(vehicle)

	// sync the slots - in order
	i := 0
	for i < len(parkingLot.GetAvailableSlots()) && parkingLot.GetAvailableSlots()[i] < leaveCmd.slot {
		i++
	}

	updatedAvailableSlots := append(parkingLot.GetAvailableSlots()[:i], append([]int{leaveCmd.slot}, parkingLot.GetAvailableSlots()[i:]...)...)
	parkingLot.UpdateAvailableSlots(updatedAvailableSlots)

	writeToOutput(leaveCmd.owriter, fmt.Sprintf(newlineOrNothing+"Slot number %d is free", leaveCmd.slot))
}
func (statusCmd *StatusCommand) Execute(ctx context.Context) {
	slots := make([]int, 0, len(parkingLot.GetOccupiedSlots()))
	for slot := range parkingLot.GetOccupiedSlots() {
		slots = append(slots, slot)
	}

	sort.Ints(slots)

	writeToOutput(statusCmd.owriter, fmt.Sprintf(newlineOrNothing+"%-10s %-20s %-10s", "Slot No.", "Registration No", "Color"))
	for _, slot := range slots {
		vehicle := parkingLot.GetOccupiedSlots()[slot]
		writeToOutput(statusCmd.owriter, fmt.Sprintf("\n%-10d %-20s %-10s", slot, vehicle.GetRegistrationNo(), vehicle.GetColor()))
	}
}
func (qRegNoByColorCmd *QueryRegistrationNoByColorCommand) Execute(ctx context.Context) {
	vehicles, ok := parkingLot.GetVehiclesByColor(qRegNoByColorCmd.color)
	if !ok {
		writeToOutput(qRegNoByColorCmd.owriter, "Not found")
		return
	}

	output := make([]string, 0)
	for _, vehicle := range vehicles {
		output = append(output, vehicle.GetRegistrationNo())
	}

	writeToOutput(qRegNoByColorCmd.owriter, fmt.Sprintf(newlineOrNothing+"%s", strings.Join(output, ", ")))
}
func (qSlotNoByRegNoCmd *QuerySlotNoByRegistrationNoCommand) Execute(ctx context.Context) {
	slot, exists := parkingLot.GetSlotByRegistrationNo(qSlotNoByRegNoCmd.registrationNo)
	if !exists {
		writeToOutput(qSlotNoByRegNoCmd.owriter, "Not found"+newlineOrNothing)
		return
	}

	writeToOutput(qSlotNoByRegNoCmd.owriter, fmt.Sprintf("%d\n", slot))
}
func (qSlotNoByColorCmd *QuerySlotNoByColorCommand) Execute(ctx context.Context) {
	vehicles, exists := parkingLot.GetVehiclesByColor(qSlotNoByColorCmd.color)
	if !exists {
		writeToOutput(qSlotNoByColorCmd.owriter, "Not found"+newlineOrNothing)
		return
	}

	slots := []string{}
	for _, vehicle := range vehicles {
		if slot, exists := parkingLot.GetSlotByRegistrationNo(vehicle.GetRegistrationNo()); exists {
			slots = append(slots, fmt.Sprintf("%d", slot))
		}
	}

	writeToOutput(qSlotNoByColorCmd.owriter, fmt.Sprintf(newlineOrNothing+"%s\n", strings.Join(slots, ", ")))
}

func writeToOutput(writer *bufio.Writer, message string) {
	fmt.Fprintf(writer, "%s", message)
	writer.Flush()
}
