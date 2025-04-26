package engine

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

/*
This is will be an intery point for the fuzzing engine
*/

// Config hlds the global configuration for stateStinger
type Config struct {
	TargetPath   string // Path to the target binary
	ModuleName   string // Name of the module to be fuzzed
	FuzzCount    int
	Seed         int64
	OutputDir    string
	Verbose      bool
	StateMutator []string
	SpecialCases bool
}

var GlobalConfig Config

func Execute() {
	flag.StringVar(&GlobalConfig.TargetPath, "target", "", "Path to the Cosmos SDK module directory to fuzz")
	flag.StringVar(&GlobalConfig.ModuleName, "module", "", "Name of the module to target")
	flag.IntVar(&GlobalConfig.FuzzCount, "count", 5000, "Number of fuzzing iterations")
	flag.Int64Var(&GlobalConfig.Seed, "seed", 0, "Random seed (0 for time-based)")
	flag.StringVar(&GlobalConfig.OutputDir, "output", "./fuzz_results", "Directory to store results")
	flag.BoolVar(&GlobalConfig.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&GlobalConfig.SpecialCases, "special", true, "Enable special case testing")

	flag.Parse()

	if GlobalConfig.TargetPath == "" {
		fmt.Println("Error: Target path is required")
		flag.Usage()
		os.Exit(1)
	}

	if GlobalConfig.ModuleName == "" {
		GlobalConfig.ModuleName = filepath.Base(GlobalConfig.TargetPath)
		log.Printf("Module name not provided, using directory name: %s", GlobalConfig.ModuleName)
	}

	if err := os.MkdirAll(GlobalConfig.OutputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Initialize and start the fuzzing engine
	engine := NewFuzzerEngine(GlobalConfig)
	results := engine.Run()

	//Report results
	fmt.Printf("\n=== StateStinger Fuzzing Results ===\n")
	fmt.Printf("Total tests run: %d\n", results.TotalTests)
	fmt.Printf("Failed tests: %d\n", results.Failed)
	fmt.Printf("State inconsistencies: %d\n", results.StateInconsistencies)
	fmt.Printf("Consensus failures: %d\n", results.ConsensusFailures)
	fmt.Printf("Crashes detected: %d\n", results.Crashes)

	if results.Failed > 0 {
		fmt.Printf("\nDetailed failure reports saved to: %s\n", GlobalConfig.OutputDir)
	}
}
