package types

type (
	ParkingLot struct {
		capacity       uint32
		availableSlots []int

		// O(1) constant time access
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
		capacity:          uint32(capacity),
		availableSlots:    availableSlots,
		occupiedSlots:     map[int]*Vehicle{},
		vehicleToSlotMap:  map[string]int{},
		colorToVehicleMap: map[string][]Vehicle{},
	}
}
