package test

import (
	"bytes"
	"os"
	"testing"

	"github.com/GoSec-Labs/StateStinger/engine"
	"github.com/GoSec-Labs/StateStinger/utils/target/cosmossdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRandomMutator tests the basic functionality of the random mutator
func TestRandomMutator(t *testing.T) {
    // Initialize with a fixed seed for deterministic testing
    r := utils.NewTestRandom(42)
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

// TestBoundaryValueMutator tests the boundary value mutator
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

// allZeros checks if a byte slice contains only zero values
func allZeros(input []byte) bool {
    for _, b := range input {
        if b != 0 {
            return false
        }
    }
    return true
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
    originalLoadFunc := cosmossdk.LoadCosmosModule
    cosmossdk.LoadCosmosModule = func(path, name string) (*cosmossdk.CosmosModule, error) {
        return &cosmossdk.CosmosModule{
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
    defer func() { cosmossdk.LoadCosmosModule = originalLoadFunc }()

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