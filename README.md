# Introduction

Golang WebAssembly VM. WebAssembly program written in golang can be run by this VM.

Implemented Golang Webassembly JS Runtime for the Webassembly VM written by Golang.  
So the VM can run the wasm files built by Golang, instead of only the wasm files built by C/C++ or Rust.

## Getting Started

### Build the test wasm if you want

```bash
# go to the test folder
cd gowasm/testdata/fmt

# build the test wasm program
GOOS=js GOARCH=wasm go build -o main.wasm

# go back
cd -
```

### Build the WebAssembly VM program

```bash
# download the dependencies
go mod download

# build main program
go build

# run your wasm program
./life -entry run ../../gowasm/testdata/fmt/main.wasm # entry point is `run`
```


