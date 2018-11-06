# Introduction

This is a WebAssembly VM which can run wasm program built by golang.

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


