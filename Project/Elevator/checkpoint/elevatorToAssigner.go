package checkpoint


import (
    "heislab/Elevator/elev"
    "heislab/Elevator/elevio"
	"fmt"
)



func indexToWord(index int) string {
    numWords := map[int]string{
        1: "one",
        2: "two",
        3: "three",
        // Extend this map as needed
    }
    // Default to a numeric fallback if the map doesn't cover the index
    if word, exists := numWords[index]; exists {
        return word
    }
    // Fallback to numeric string if index is not in map
    return fmt.Sprintf("%d", index)
}


func convertElvToAssignerFormat(elevators []elev.Elevator) HRAInput {
    var hraInput HRAInput
    hraInput.States = make(map[string]HRAElevState)
    hraInput.HallRequests = make([][2]bool, elevio.NFloors)

    for i, elv := range elevators {
        // Convert Elevator.CurrentBehaviour to a string for HRAElevState.Behavior
        var behavior string
        switch elv.CurrentBehaviour {
        case elev.EBIdle:
            behavior = "idle"
        case elev.EBMoving:
            behavior = "moving"
        case elev.EBDoorOpen:
            behavior = "door open"
        }

        // Convert Elevator.Dirn to a string for HRAElevState.Direction
        var direction string
        switch elv.Dirn {
        case elevio.DirUp:
            direction = "up"
        case elevio.DirDown:
            direction = "down"
        default:
            direction = "stop"
        }

        // Convert Elevator.Requests to CabRequests and HallRequests
        cabRequests := make([]bool, elevio.NFloors)
        for f := 0; f < elevio.NFloors; f++ {
            cabRequests[f] = elv.Requests[f][elevio.BCab]
            hraInput.HallRequests[f][0] = hraInput.HallRequests[f][0] || elv.Requests[f][elevio.BHallDown]
            hraInput.HallRequests[f][1] = hraInput.HallRequests[f][1] || elv.Requests[f][elevio.BHallUp]
        }


        // Populate HRAInput.States with the converted Elevator states
        hraInput.States[indexToWord(i+1)] = HRAElevState{
            Behavior:    behavior,
            Floor:       elv.CurrentFloor,
            Direction:   direction,
            CabRequests: cabRequests,
        }
    }

    return hraInput
}
