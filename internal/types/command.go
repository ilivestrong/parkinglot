package types

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
	parkingLot     *ParkingLot
	tokensWithArgs map[string]struct{} = map[string]struct{}{
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
		Execute(ctx context.Context) (interface{}, error)
	}

	CreateParkingLotCommand struct {
		capacity int
	}
	ParkCommand struct {
		vehicle *Vehicle
	}
	LeaveCommand struct {
		slot int
	}
	StatusCommand                     struct{}
	QueryRegistrationNoByColorCommand struct {
		color string
	}
	QuerySlotNoByRegistrationNoCommand struct {
		registrationNo string
	}
	QuerySlotNoByColorCommand struct {
		color string
	}

	CommandBuilder struct{}
)

/*
Instantiates a command builder object.

It allows to parse command(s) individually or in bulk.
*/
func NewCommandBuilder(mode string) *CommandBuilder {
	if mode == "interactive" {
		newlineOrNothing = ""
	}
	return &CommandBuilder{}
}

/*
Parses command tokens and their args and returns a concrete Command object.

The returned command object is used to execute the command on-demand.
*/
func (cb *CommandBuilder) ParseCommand(token string, args ...string) Commander {
	if token == "" {
		log.Println("token is missing")
		return nil
	}
	if _, ok := tokensWithArgs[token]; ok && len(args) == 0 {
		log.Printf("\nargs missing for token: %s", token)
		return nil
	}

	switch strings.TrimSpace(token) {
	case TokenForCreateParkingLot:
		capacity, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			log.Printf("\ninvalid args provided for token: %s", token)
		}
		return &CreateParkingLotCommand{
			capacity: capacity,
		}
	case TokenForPark:
		return &ParkCommand{
			vehicle: &Vehicle{registrationNumber: strings.TrimSpace(args[0]), color: strings.TrimSpace(args[1])},
		}
	case TokenForLeave:
		slot, _ := strconv.Atoi(strings.TrimSpace(args[0]))
		return &LeaveCommand{
			slot: slot,
		}
	case TokenForStatus:
		return &StatusCommand{}
	case TokenForQueryRegistrationNoByColor:
		return &QueryRegistrationNoByColorCommand{
			color: strings.TrimSpace(args[0]),
		}
	case TokenForQuerySlotNoByRegistrationNo:
		return &QuerySlotNoByRegistrationNoCommand{
			registrationNo: strings.TrimSpace(args[0]),
		}
	case TokenForQuerySlotNoByColor:
		return &QuerySlotNoByColorCommand{
			color: strings.TrimSpace(args[0]),
		}
	case "":
		fmt.Printf("\nno command provided \n\n")
		return nil
	default:
		fmt.Printf("\n invalid command: %s, skipping...\n\n", strings.TrimSpace(token))
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

func (cplCmd *CreateParkingLotCommand) Execute(ctx context.Context) (interface{}, error) {
	if cplCmd.capacity > MaxNumberOfSlots {
		fmt.Printf("cannot create :%d slots. Maximum allowed is: %d", cplCmd.capacity, MaxNumberOfSlots)
		return nil, nil
	}
	fmt.Printf("Created a parking lot with %d slots", cplCmd.capacity)
	parkingLot = NewParkingLot(cplCmd.capacity)
	return nil, nil
}
func (parkCmd *ParkCommand) Execute(ctx context.Context) (interface{}, error) {
	if len(parkingLot.availableSlots) == 0 {
		fmt.Printf("\nSorry, parking lot is full")
		return nil, fmt.Errorf("no slots available")
	}

	slot := int(parkingLot.availableSlots[0])
	parkingLot.availableSlots = parkingLot.availableSlots[1:]
	parkingLot.occupiedSlots[slot] = parkCmd.vehicle
	parkingLot.vehicleToSlotMap[parkCmd.vehicle.registrationNumber] = slot
	parkingLot.colorToVehicleMap[parkCmd.vehicle.color] = append(parkingLot.colorToVehicleMap[parkCmd.vehicle.color], *parkCmd.vehicle)
	fmt.Printf(newlineOrNothing+"Allocated slot number: %d", slot)
	return slot, nil
}
func (leaveCmd *LeaveCommand) Execute(ctx context.Context) (interface{}, error) {
	vehicle, exists := parkingLot.occupiedSlots[leaveCmd.slot]
	if !exists {
		fmt.Printf("slot %d is not occupied", leaveCmd.slot)
		return nil, nil
	}

	delete(parkingLot.occupiedSlots, leaveCmd.slot)
	delete(parkingLot.vehicleToSlotMap, vehicle.registrationNumber)

	// Remove vehicle from colorToVehicles map
	for color, vehicles := range parkingLot.colorToVehicleMap {
		for i, v := range vehicles {
			if v.registrationNumber == vehicle.registrationNumber {
				parkingLot.colorToVehicleMap[color] = append(vehicles[:i], vehicles[i+1:]...)
				break
			}
		}
	}

	// sync the slots - in order
	i := 0
	for i < len(parkingLot.availableSlots) && parkingLot.availableSlots[i] < leaveCmd.slot {
		i++
	}

	parkingLot.availableSlots = append(parkingLot.availableSlots[:i], append([]int{leaveCmd.slot}, parkingLot.availableSlots[i:]...)...)
	fmt.Printf("\nSlot number %d is free", leaveCmd.slot)
	return nil, nil
}
func (statusCmd *StatusCommand) Execute(ctx context.Context) (interface{}, error) {
	slots := make([]int, 0, len(parkingLot.occupiedSlots))
	for slot, _ := range parkingLot.occupiedSlots {
		slots = append(slots, slot)
	}

	sort.Ints(slots)

	fmt.Printf("\n%-10s %-20s %-10s", "Slot No.", "Registration No", "Color")
	for _, slot := range slots {
		vehicle := parkingLot.occupiedSlots[slot]
		fmt.Printf("\n%-10d %-20s %-10s", slot, vehicle.registrationNumber, vehicle.color)
	}
	return nil, nil
}
func (qRegNoByColorCmd *QueryRegistrationNoByColorCommand) Execute(ctx context.Context) (interface{}, error) {
	vehicles, ok := parkingLot.colorToVehicleMap[qRegNoByColorCmd.color]
	// fmt.Println(qRegNoByColorCmd.color, vehicles)
	if !ok {
		fmt.Println("Not found")
		return nil, nil
	}

	output := make([]string, 0)
	for _, vehicle := range vehicles {
		output = append(output, vehicle.registrationNumber)
	}

	fmt.Printf("\n%s", strings.Join(output, ", "))
	return nil, nil
}
func (qSlotNoByRegNoCmd *QuerySlotNoByRegistrationNoCommand) Execute(ctx context.Context) (interface{}, error) {
	slot, exists := parkingLot.vehicleToSlotMap[qSlotNoByRegNoCmd.registrationNo]
	if !exists {
		fmt.Println("Not found")
		return nil, nil
	}

	fmt.Println(slot)
	return nil, nil
}
func (qSlotNoByColorCmd *QuerySlotNoByColorCommand) Execute(ctx context.Context) (interface{}, error) {
	vehicles, exists := parkingLot.colorToVehicleMap[qSlotNoByColorCmd.color]
	if !exists {
		fmt.Println("Not found")
		return nil, nil
	}

	slots := []string{}
	for _, vehicle := range vehicles {
		if slot, exists := parkingLot.vehicleToSlotMap[vehicle.registrationNumber]; exists {
			slots = append(slots, fmt.Sprintf("%d", slot))
		}
	}

	fmt.Printf("\n%s\n", strings.Join(slots, ", "))
	return nil, nil
}
