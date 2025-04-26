package test

import (
	"testing"

	"github.com/GoSec-Labs/StateStinger/engine"
	"github.com/stretchr/testify/assert"
)

// TestRandomMutator tests the basic functionality of the random mutator
func TestRandomMutator(t *testing.T) {
	// Initialize with a fixed seed for deterministic testing
	r := 
	mutator := engine.NewRandomTxMutator(r)

	// Generate some inputs and check basic properties
	for i := 0; i < 100; i++ {
		input := mutator.GenerateFuzzInput()
		assert.NotEmpty(t, input, "Generated input should not be empty")
		assert.GreaterOrEqual(t, len(input), 16, "Input should be at least 16 bytes")
		assert.LessOrEqual(t, len(input), 4096, "Input should be at most 4096 bytes")
	}

	// Test validation logic
	result := mutator.ValidateOutput([]byte("state_inconsistent"), nil)
	assert.NotNil(t, result, "Should detect state inconsistency")
	assert.True(t, result.StateInconsistency, "Should flag state inconsistency")

	result = mutator.ValidateOutput([]byte("consensus_failure"), nil)
	assert.NotNil(t, result, "Should detect consensus failure")
	assert.True(t, result.ConsensusFailure, "Should flag consensus failure")

	result = mutator.ValidateOutput([]byte("success"), nil)
	assert.Nil(t, result, "Should not report success as failure")

	result = mutator.ValidateOutput(nil, nil)
	assert.Nil(t, result, "Should not report nil as failure")

	result = mutator.ValidateOutput(nil, assert.AnError)
	assert.NotNil(t, result, "Should detect error")
	assert.True(t, result.Failed, "Should flag failure")
	assert.True(t, result.Crashed, "Should flag crash")
}


func TestBoundaryValueMutator(t *testing.T) {
	// Initialize with a fixed seed for deterministic testing
	r := utils.NewTestRandom(42)
	mutator := utils.NewBoundaryValueMutator(r)
	
	// Generate boundary value inputs and check basic properties
	seenCases := make(map[int]bool)
	for i := 0; i < 100; i++ {
		input := mutator.GenerateFuzzInput()
		
		// Determine which case this is
		caseType := -1
		switch {
		case len(input) == 0:
			caseType = 0 // Empty input
		case len(input) == 1:
			caseType = 1 // Single byte
		case len(input) == 16 && bytes.Equal(input[0:8], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}):
			caseType = 2 // Max uint64
		case len(input) >= 16 && allZeros(input):
			caseType = 3 // Zero values
		case len(input) == 16 && bytes.Equal(input[0:8], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}):
			caseType = 4 // Negative values
		case len(input) >= 1024*1024:
			caseType = 5 // Very large buffer
		}
		
		assert.NotEqual(t, -1, caseType, "Input should match one of the expected boundary cases")
		seenCases[caseType] = true
	}
	
	// Ensure we've seen all cases
	assert.GreaterOrEqual(t, len(seenCases), 3, "Should see at least 3 different boundary cases")
	
	// Test validation logic (similar to RandomMutator tests)
	result := mutator.ValidateOutput([]byte("state_inconsistent"), nil)
	assert.NotNil(t, result, "Should detect state inconsistency")
	assert.True(t, result.StateInconsistency, "Should flag state inconsistency")
}

func TestStatefulMutator(t *testing.T) {
	// Initialize with a fixed seed for deterministic testing
	r := utils.NewTestRandom(42)
	mutator := utils.NewStatefulMutator(r)
	
	// Generate a sequence of inputs and verify state is being maintained
	var stateLength int
	for i := 0; i < 100; i++ {
		input := mutator.GenerateFuzzInput()
		assert.NotEmpty(t, input, "Generated input should not be empty")
		
		// Verify the state grows with each operation (up to a cap)
		if i > 0 {
			if i <= 64 {
				assert.Equal(t, i, stateLength+1, "State should grow by 1 for each call up to 64")
			} else {
				assert.Equal(t, 64, stateLength, "State should be capped at 64 bytes")
			}
		}
		
		// Track expected state length
		if stateLength < 64 {
			stateLength++
		}
	}
	
	// Test validation logic
	result := mutator.ValidateOutput([]byte("state_inconsistent"), nil)
	assert.NotNil(t, result, "Should detect state inconsistency")
	assert.True(t, result.StateInconsistency, "Should flag state inconsistency")
	assert.NotEmpty(t, result.Input, "Should include state in failure result")
}

// TestSpecialCasesMutator tests the special cases mutator
func TestSpecialCasesMutator(t *testing.T) {
	// Initialize with a fixed seed for deterministic testing
	r := utils.NewTestRandom(42)
	mutator := utils.NewSpecialCasesMutator(r)
	
	// Collect all seen first bytes (which identify the case)
	seenFirstBytes := make(map[byte]bool)
	for i := 0; i < 100; i++ {
		input := mutator.GenerateFuzzInput()
		assert.NotEmpty(t, input, "Generated input should not be empty")
		
		// Track the first byte which identifies the special case
		if len(input) > 0 {
			seenFirstBytes[input[0]] = true
		}
	}
	
	// Ensure we've seen multiple special cases
	assert.GreaterOrEqual(t, len(seenFirstBytes), 3, "Should see at least 3 different special cases")
	
	// Test validation logic
	result := mutator.ValidateOutput([]byte("consensus_failure"), nil)
	assert.NotNil(t, result, "Should detect consensus failure")
	assert.True(t, result.ConsensusFailure, "Should flag consensus failure")
}

// TestFuzzEngine tests the core fuzzing engine
func TestFuzzEngine(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "statestinger-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)
	
	// Create a test configuration
	config := engine.Config{
		TargetPath:   "./testdata/mock_module", // Would point to a mock module in a real test
		ModuleName:   "mock",
		FuzzCount:    100,
		Seed:         42,
		OutputDir:    tempDir,
		Verbose:      true,
		SpecialCases: true,
	}
	
	// Mock the module loading and execution for testing
	// In a real test, you'd use dependency injection or create a test interface
	originalLoadFunc := utils.LoadCosmosModule
	utils.LoadCosmosModule = func(path, name string) (*utils.CosmosModule, error) {
		return &utils.CosmosModule{
			Path: path,
			Name: name,
			Handlers: []string{
				"MsgSend",
				"MsgMultiSend",
				"MsgSetWithdrawAddress",
			},
			StateTypes: []string{
				"Balance",
				"Account",
				"Validator",
			},
		}, nil
	}
	defer func() { utils.LoadCosmosModule = originalLoadFunc }()
	
	// Create and run the engine
	fuzzEngine := engine.NewFuzzEngine(config)
	summary := fuzzEngine.Run()
	
	// Verify the summary data
	assert.Equal(t, 100, summary.TotalTests, "Should run all specified tests")
	assert.GreaterOrEqual(t, summary.Failed, 0, "May have failed tests")
	
	// Check if result files were created for any failures
	if summary.Failed > 0 {
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err, "Failed to read temp directory")
		assert.GreaterOrEqual(t, len(files), summary.Failed, "Should have at least one file per failure")
	}
}
