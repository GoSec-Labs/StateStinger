package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/GoSec-Labs/StateStinger/utils/target/cosmossdk"
)

// FuzzResult tracks the outcome of a single fuzzing test
type FuzzResult struct {
	ID                 string
	Input              []byte
	ErrorMessage       string
	Failed             bool
	StateInconsistency bool
	ConsensusFailure   bool
	Crashed            bool
}

// FuzzSummary contains aggregate results from a fuzzing run
type FuzzSummary struct {
	TotalTests           int
	Failed               int
	StateInconsistencies int
	ConsensusFailures    int
	Crashes              int
}

// FuzzEngine is the core fuzzing implementation
type FuzzEngine struct {
	config   Config
	rand     *rand.Rand
	mutators []StateMutator
	results  []FuzzResult
	summary  FuzzSummary
}

// StateMutator defines an interface for state mutation strategies
type StateMutator interface {
	Name() string
	GenerateFuzzInput() []byte
	ValidateOutput(output []byte, err error) *FuzzResult
}

// NewFuzzEngine creates a new fuzzing engine with the given configuration
func NewFuzzerEngine(config Config) *FuzzEngine {
	var seed int64
	if config.Seed == 0 {
		seed = time.Now().UnixNano()
	} else {
		seed = config.Seed
	}

	rng := rand.New(rand.NewSource(seed))

	log.Printf("Initializing StateStinger with seed: %d", seed)

	engine := &FuzzEngine{
		config:   config,
		rand:     rng,
		mutators: make([]StateMutator, 0),
		results:  make([]FuzzResult, 0),
	}

	// Register default mutators
	engine.registerDefaultMutators()

	return engine
}

// registerDefaultMutators adds the standard mutation strategies
func (f *FuzzEngine) registerDefaultMutators() {
	f.mutators = append(f.mutators, NewRandomTxMutator(f.rand))
	f.mutators = append(f.mutators, NewBoundaryValueMutator(f.rand))
	f.mutators = append(f.mutators, NewStatefulMutator(f.rand))

	if f.config.SpecialCases {
		f.mutators = append(f.mutators, NewSpecialCasesMutator(f.rand))
	}

	log.Printf("Registered %d mutation strategies", len(f.mutators))
}

// Run executes the fuzzing process
func (f *FuzzEngine) Run() FuzzSummary {
	log.Printf("Starting fuzzing run with %d iterations on module '%s'",
		f.config.FuzzCount, f.config.ModuleName)

	// Set up target module
	targetModule, err := cosmossdk.LoadCosmosModule(f.config.TargetPath, f.config.ModuleName)
	if err != nil {
		log.Fatalf("Failed to load target module: %v", err)
	}

	// Main fuzzing loop
	for i := 0; i < f.config.FuzzCount; i++ {
		if i > 0 && i%1000 == 0 {
			log.Printf("Progress: %d/%d iterations completed", i, f.config.FuzzCount)
		}

		// Choose a random mutator
		mutator := f.mutators[f.rand.Intn(len(f.mutators))]

		// Generate fuzz input
		input := mutator.GenerateFuzzInput()

		// Execute on target
		output, err := targetModule.ExecuteFuzz(input)

		// Validate result
		result := mutator.ValidateOutput(output, err)

		// Track result
		if result != nil {
			f.trackResult(result)
		}

		f.summary.TotalTests++
	}

	return f.summary
}

// trackResult processes and tracks the result of a fuzzing iteration
func (f *FuzzEngine) trackResult(result *FuzzResult) {
	if result.Failed {
		f.recordFailure(*result)
		f.summary.Failed++

		if result.StateInconsistency {
			f.summary.StateInconsistencies++
		}
		if result.ConsensusFailure {
			f.summary.ConsensusFailures++
		}
		if result.Crashed {
			f.summary.Crashes++
		}
	}
}

// recordFailure saves detailed information about a failed test
func (f *FuzzEngine) recordFailure(result FuzzResult) {
	// Create a unique filename
	filename := filepath.Join(f.config.OutputDir,
		fmt.Sprintf("failure_%s.json", result.ID))

	// Save failure details to file
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Warning: Failed to save failure details: %v", err)
		return
	}
	defer file.Close()

	// Serialize the result to JSON
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(result); err != nil {
		log.Printf("Warning: Failed to serialize failure details: %v", err)
	}

	if f.config.Verbose {
		log.Printf("Failure detected [%s]: %s", result.ID, result.ErrorMessage)
	}
}
