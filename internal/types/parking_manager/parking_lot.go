package parkingmanager

type (
	ParkingLot struct {
		capacity          int
		availableSlots    []int
		occupiedSlots     map[int]*Vehicle
		vehicleToSlotMap  map[string]int
		colorToVehicleMap map[string][]Vehicle
	}

	Vehicle struct {
		registrationNumber string
		color              string
	}
)

func NewParkingLot(capacity int) *ParkingLot {
	availableSlots := make([]int, capacity)
	for i := range capacity {
		availableSlots[i] = i + 1
	}

	return &ParkingLot{
		capacity:          capacity,
		availableSlots:    availableSlots,
		occupiedSlots:     map[int]*Vehicle{},
		vehicleToSlotMap:  map[string]int{},
		colorToVehicleMap: map[string][]Vehicle{},
	}
}

func NewVehicle(registrationNo string, color string) *Vehicle {
	return &Vehicle{
		registrationNumber: registrationNo,
		color:              color,
	}
}

func (vehicle *Vehicle) GetRegistrationNo() string {
	return vehicle.registrationNumber
}
func (vehicle *Vehicle) GetColor() string {
	return vehicle.color
}

func (pl *ParkingLot) GetAvailableSlots() []int {
	return pl.availableSlots
}
func (pl *ParkingLot) GetOccupiedSlots() map[int]*Vehicle {
	return pl.occupiedSlots
}
func (pl *ParkingLot) GetVehicleToSlotMapping() map[string]int {
	return pl.vehicleToSlotMap
}
func (pl *ParkingLot) GetColorToVehicleMapping() map[string][]Vehicle {
	return pl.colorToVehicleMap
}

func (pl *ParkingLot) UpdateAvailableSlots(latestAvailableSlots []int) {
	pl.availableSlots = latestAvailableSlots
}
func (pl *ParkingLot) UpdateOccupiedSlot(slot int, vehicle *Vehicle) {
	pl.occupiedSlots[slot] = vehicle
}
func (pl *ParkingLot) UpdateSlotByRegistrationNo(registrationNo string, slot int) {
	pl.vehicleToSlotMap[registrationNo] = slot
}
func (pl *ParkingLot) UpdateVehiclesByColor(color string, vehicle *Vehicle) {
	pl.colorToVehicleMap[color] = append(pl.colorToVehicleMap[color], *vehicle)
}

func (pl *ParkingLot) RemoveSlotFromOccupiedSlots(slot int) {
	delete(pl.GetOccupiedSlots(), slot)
}
func (pl *ParkingLot) RemoveRegistrationNoFromOccupiedSlots(registrationNo string) {
	delete(pl.vehicleToSlotMap, registrationNo)
}
func (pl *ParkingLot) RemoveVehicleFromColorToVehicleMapping(vehicle *Vehicle) {
	for color, vehicles := range pl.colorToVehicleMap {
		for i, v := range vehicles {
			if v.registrationNumber == vehicle.GetRegistrationNo() {
				pl.colorToVehicleMap[color] = append(vehicles[:i], vehicles[i+1:]...)
				break
			}
		}
	}
}

func (pl *ParkingLot) GetVehiclesByColor(color string) ([]Vehicle, bool) {
	vehicles, ok := pl.colorToVehicleMap[color]
	if !ok {
		return nil, false
	}
	return vehicles, true
}

func (pl *ParkingLot) GetSlotByRegistrationNo(registrationNo string) (int, bool) {
	slot, ok := pl.vehicleToSlotMap[registrationNo]
	if !ok {
		return -1, false
	}
	return slot, true
}
