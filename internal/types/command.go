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

	pm "github.com/ilivestrong/internal/types/parking_manager"
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
	}
	ParkCommand struct {
		vehicle *pm.Vehicle
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
func (cb *CommandBuilder) ParseCommand(commandName string, args ...string) Commander {
	if _, ok := commandsWithArgs[commandName]; ok && len(args) == 0 {
		log.Printf("\nargs missing for command: %s", commandName)
		return nil
	}

	switch commandName {
	case TokenForCreateParkingLot:
		capacity, err := strconv.Atoi(args[0])
		if err != nil {
			log.Printf("\ninvalid args provided for command: %s", commandName)
		}
		return &CreateParkingLotCommand{
			capacity: capacity,
		}
	case TokenForPark:
		return &ParkCommand{
			vehicle: pm.NewVehicle(args[0], args[1]),
		}
	case TokenForLeave:
		slot, _ := strconv.Atoi(args[0])
		return &LeaveCommand{
			slot: slot,
		}
	case TokenForStatus:
		return &StatusCommand{}
	case TokenForQueryRegistrationNoByColor:
		return &QueryRegistrationNoByColorCommand{
			color: args[0],
		}
	case TokenForQuerySlotNoByRegistrationNo:
		return &QuerySlotNoByRegistrationNoCommand{
			registrationNo: args[0],
		}
	case TokenForQuerySlotNoByColor:
		return &QuerySlotNoByColorCommand{
			color: args[0],
		}
	case "":
		fmt.Printf("\nno command provided \n\n")
		return nil
	default:
		fmt.Printf("\n invalid command: %s, skipping...\n\n", commandName)
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
		fmt.Printf("cannot create :%d slots. Maximum allowed is: %d", cplCmd.capacity, MaxNumberOfSlots)
		return
	}

	if cplCmd.capacity <= 0 {
		fmt.Printf("invalid slot number: %d", cplCmd.capacity)
		return
	}

	fmt.Printf("Created a parking lot with %d slots", cplCmd.capacity)
	parkingLot = pm.NewParkingLot(cplCmd.capacity)
}
func (parkCmd *ParkCommand) Execute(ctx context.Context) {
	if len(parkingLot.GetAvailableSlots()) == 0 {
		fmt.Printf("%s", newlineOrNothing+"Sorry, parking lot is full")
		return
	}

	slot := int(parkingLot.GetAvailableSlots()[0])
	parkingLot.UpdateAvailableSlots(parkingLot.GetAvailableSlots()[1:])
	parkingLot.UpdateOccupiedSlot(slot, parkCmd.vehicle)
	parkingLot.UpdateSlotByRegistrationNo(parkCmd.vehicle.GetRegistrationNo(), slot)
	parkingLot.UpdateVehiclesByColor(parkCmd.vehicle.GetColor(), parkCmd.vehicle)

	fmt.Printf(newlineOrNothing+"Allocated slot number: %d", slot)
}

func (leaveCmd *LeaveCommand) Execute(ctx context.Context) {
	vehicle, exists := parkingLot.GetOccupiedSlots()[leaveCmd.slot]
	if !exists {
		fmt.Printf("slot %d is not occupied", leaveCmd.slot)
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

	fmt.Printf(newlineOrNothing+"Slot number %d is free", leaveCmd.slot)
}
func (statusCmd *StatusCommand) Execute(ctx context.Context) {
	slots := make([]int, 0, len(parkingLot.GetOccupiedSlots()))
	for slot, _ := range parkingLot.GetOccupiedSlots() {
		slots = append(slots, slot)
	}

	sort.Ints(slots)

	fmt.Printf(newlineOrNothing+"%-10s %-20s %-10s", "Slot No.", "Registration No", "Color")
	for _, slot := range slots {
		vehicle := parkingLot.GetOccupiedSlots()[slot]
		fmt.Printf("\n%-10d %-20s %-10s", slot, vehicle.GetRegistrationNo(), vehicle.GetColor())
	}
}
func (qRegNoByColorCmd *QueryRegistrationNoByColorCommand) Execute(ctx context.Context) {
	vehicles, ok := parkingLot.GetVehiclesByColor(qRegNoByColorCmd.color)
	if !ok {
		fmt.Println("Not found")
		return
	}

	output := make([]string, 0)
	for _, vehicle := range vehicles {
		output = append(output, vehicle.GetRegistrationNo())
	}

	fmt.Printf(newlineOrNothing+"%s", strings.Join(output, ", "))
}
func (qSlotNoByRegNoCmd *QuerySlotNoByRegistrationNoCommand) Execute(ctx context.Context) {
	slot, exists := parkingLot.GetSlotByRegistrationNo(qSlotNoByRegNoCmd.registrationNo)
	if !exists {
		fmt.Printf("%s", "Not found"+newlineOrNothing)
		return
	}

	fmt.Println(slot)
}
func (qSlotNoByColorCmd *QuerySlotNoByColorCommand) Execute(ctx context.Context) {
	vehicles, exists := parkingLot.GetVehiclesByColor(qSlotNoByColorCmd.color)
	if !exists {
		fmt.Printf("%s", "Not found"+newlineOrNothing)
		return
	}

	slots := []string{}
	for _, vehicle := range vehicles {
		if slot, exists := parkingLot.GetSlotByRegistrationNo(vehicle.GetRegistrationNo()); exists {
			slots = append(slots, fmt.Sprintf("%d", slot))
		}
	}

	fmt.Printf(newlineOrNothing+"%s\n", strings.Join(slots, ", "))
}
