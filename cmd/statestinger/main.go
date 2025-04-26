package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/GoSec-Labs/StateStinger/engine"
)

// Version information (set during build)
var Version = "dev"

func main() {
	// Print version info
	fmt.Printf("StateStinger v%s (Cosmos SDK State Machine Fuzzer)\n", Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Println("--------------------------------------------------")

	// Execute the main command
	engine.Execute()

	// Exit with appropriate status code
	// This is managed by the Execute function, we just need
	// to make sure we return to OS after it's done
	os.Exit(0)
}
