package main

import (
"math"
)

// [ElevatorNumber[[Flornumber, Is elevator at this floor],Vector of[number in que, direction (down, up, cabin-call)]]]
// [Elevator = 1[ // [Elevator = 2[ ] // [Elevator = 3 []
// [4, X ==1],[1, down]]
// [[3.5, 0],[]]
// [[3, 0],],[]]
// [[2.5, 0],[]]
// [[2, 0],],[]]
// [[1.5,0],[]]
// [[1, 0],[]]]

type FloorQue struct {
NumberInQueue int
Direction string // "down", "up", "cabin-call"
Weight float64 // Total cost of the order (note the capital W for exporting)
}

type Order struct {
Floor float64
Direction string // "down", "up", "cabin-call"
}

type FloorInfo struct {
FloorNumber float64
AtFloor bool
CurrentDirection string
Calls []FloorQue
}

type Elevator struct {
ElevatorNumber int
Floors []FloorInfo
Direction string
}

// IsElevatorQueueEmpty sjekker om alle Calls for alle Floors er tomme for en enkelt Elevator
func IsElevatorQueueEmpty(elevator Elevator) bool {
for _, floor := range elevator.Floors {
if len(floor.Calls) > 0 {
return false
}
}
// Alle Calls for alle Floors i denne Elevator er tomme
return true
}

func OrderInQueue(elevator Elevator, order Order) bool {
for _, floor := range elevator.Floors {
if floor.FloorNumber == order.Floor {
for _, call := range floor.Calls {
if call.Direction == order.Direction {
// Ordren finnes allerede i køen, returnerer true og vekten av kallet
return true
}
}
}
}
// Ordren finnes ikke i køen, returnerer false og en vekt på 0 som standard
return false
}

func CostOrderInQueue(elevator Elevator, order Order) float64 {
for _, floor := range elevator.Floors {
if floor.FloorNumber == order.Floor {
for _, call := range floor.Calls {
if call.Direction == order.Direction {
// Ordren finnes allerede i køen, returnerer true og vekten av kallet
return call.Weight
}
}
}
}
// Ordren finnes ikke i køen, returnerer false og en vekt på 0 som standard
return 1000
}

func CostWorstCase(newOrder Order) float64 {
if newOrder.Direction == "down" {
return float64(newOrder.Floor) - 1
}
return 4 - float64(newOrder.Floor)
}

func findBiggestWeightInQue(elevator Elevator) float64 {
weight := 0.0
for _, floor := range elevator.Floors {
for _, call := range floor.Calls {
if call.Weight > weight {
weight = call.Weight
}
}
}
return weight
}


func findMaxWeightInInterval(elevator Elevator, it_fra float64, it_til float64) (maxWeight float64, maxWeightFloor float64) {
maxWeight = -1.0 // Initialize with -1 to indicate that no weight has been found yet
maxWeightFloor = -1.0 // Similar reasoning for the floor

// Iterate over each floor in the elevator
for _, floor := range elevator.Floors {
// Check if the current floor is within the specified interval
if floor.FloorNumber >= it_fra && floor.FloorNumber <= it_til {
// Iterate over each call in the current floor
for _, call := range floor.Calls {
// Update maxWeight and maxWeightFloor if this call's weight is greater than the current maxWeight
if call.Weight > maxWeight {
maxWeight = call.Weight
maxWeightFloor = floor.FloorNumber
}
}
}
}

return maxWeight, maxWeightFloor
}


func costFunction(elevators []Elevator, newOrder Order) []float64 {
var cost []float64 = make([]float64, len(elevators))

for i, elevator := range elevators {
CurrentElevatorFloor := 2.0 // Example current floor, replace with actual current floor retrieval

if IsElevatorQueueEmpty(elevator) {
cost[i] = math.Abs(CurrentElevatorFloor - newOrder.Floor)
} else {
var it_fra, it_til float64

if elevator.Direction != newOrder.Direction {
if elevator.Direction == "up" {
it_fra = CurrentElevatorFloor
it_til = 4 

} else { 
it_fra = 1 
it_til = CurrentElevatorFloor
}
maxWeight, maxWeightFloor := findMaxWeightInInterval(elevator, it_fra, it_til)
cost[i]= maxWeight + math.Abs(maxWeightFloor - newOrder.Floor)
} else { 
if (elevator.Direction == "up" && CurrentElevatorFloor <= newOrder.Floor) || (elevator.Direction == "down" && CurrentElevatorFloor >= newOrder.Floor) {
// If the elevator is already moving towards the order in the correct direction
cost[i] = math.Abs(CurrentElevatorFloor - newOrder.Floor)
} else {
// If the elevator will have to change direction or pass the floor to reach the order
it_fra = 1 // Assuming min floor is 1
it_til = 4 // Assuming max floor is 4
maxWeight, maxWeightFloor := findMaxWeightInInterval(elevator, it_fra, it_til)
cost[i]= maxWeight + math.Abs(maxWeightFloor - newOrder.Floor)
}
}
}
}

return cost
}



func main() {
elevators := []Elevator{
{
ElevatorNumber: 1,
Floors: []FloorInfo{
{FloorNumber: 4, AtFloor: true, Calls: []FloorQue{{NumberInQueue: 1, Direction: "down", Weight: 4}}},
{FloorNumber: 3.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 3, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 2.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 2, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 1.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 1, AtFloor: false, Calls: []FloorQue{}},
},
},
{
ElevatorNumber: 2,
Floors: []FloorInfo{
{FloorNumber: 4, AtFloor: true, Calls: []FloorQue{{NumberInQueue: 1, Direction: "down"}}},
{FloorNumber: 3.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 3, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 2.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 2, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 1.5, AtFloor: false, Calls: []FloorQue{}},
{FloorNumber: 1, AtFloor: false, Calls: []FloorQue{}},
},
},
// You can add more Elevator definitions here as needed
}
print(elevators)
// Use 'elevators' in your logic...
}
