package types

type (
	slotID uint32

	ParkingLot struct {
		capacity       uint32
		availableSlots []uint32

		// O(1) constant time access
		occupiedSlots     map[slotID]Vehicle
		vehicleToSlotMap  map[string]slotID
		colorToVehicleMap map[string][]Vehicle
	}

	Vehicle struct {
		registrationNumber string
		color              string
	}
)
