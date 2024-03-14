# TTK4145 Project

## Authors
* **Gustav Lokna** - [@gustavlokna](https://github.com/gustavlokna)
* **Ian Philip Eglin** - [@ipeglin](https://github.com/ipeglin)
* **Simen Fritzner** - [@simenfritzner](https://github.com/simenfritzner)

## Getting Started

### Pre-requisites
* [Go](https://go.dev/dl/) (v1.21.7 or higher)
* Installing [Hall Request Assigner (HRA)](https://dlang.org/download.html) by [@klasbo](https://github.com/klasbo)

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