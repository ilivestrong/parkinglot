package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	ParkingLotCapacityEnvKey      = "PARKING_LOT_CAPACITY"
	DefaultParkingLotCapacity int = 100
)

func main() {

	log.Println("Initializing Car Parking Lot system")

	parkingLotCapacity := DefaultParkingLotCapacity
	v, exists := os.LookupEnv(ParkingLotCapacityEnvKey)
	if exists {
		if capacity, err := strconv.Atoi(v); err != nil {
			parkingLotCapacity = capacity
		}
	}
	fmt.Println(parkingLotCapacity)

}
