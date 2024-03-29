# TTK4145 Project

## Getting Started

### Pre-requisites
* [Go](https://go.dev/dl/) (v1.21.7 or higher)
* [hall_request_assigner (HRA)](https://github.com/TTK4145/Project-resources/releases/tag/v1.1.1) (v1.1.1) by [@klasbo](https://github.com/klasbo)

### Installation

Download the source repository as zip, and extract in desired directory.

Navigate into the project directory

```bash
cd <yourpath>/TTK4145/Project
```

Add HRA dependency to the `elevator` module

```bash
mv ~/Downloads/hall_request_assigner ./elevator/
```

### Build and Run

Build the project with:

```bash
# Nagivate to module
cd ./init

# elevator argument strictly required
go build -o elevator
```

Run the executable
```bash
./elevator
```
