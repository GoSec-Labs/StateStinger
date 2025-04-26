package engine

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

// Base mutator that others can embed
type BaseMutator struct {
	rand *rand.Rand
	name string
}

func (b BaseMutator) Name() string {
	return b.name
}

// RandomTxMutator generates completely random transaction data
type RandomTxMutator struct {
	BaseMutator
}

func NewRandomTxMutator(r *rand.Rand) *RandomTxMutator {
	return &RandomTxMutator{
		BaseMutator: BaseMutator{
			rand: r,
			name: "RandomTxMutator",
		},
	}
}

func (m *RandomTxMutator) GenerateFuzzInput() []byte {
	// Generate random length between 16 and 4096 bytes
	length := 16 + m.rand.Intn(4080)
	input := make([]byte, length)

	// Fill with random bytes
	m.rand.Read(input)

	return input
}

func (m *RandomTxMutator) ValidateOutput(output []byte, err error) *FuzzResult {
	if err != nil {
		// Only report specific errors (not general validation failures)
		if err.Error() != "invalid arguments" && err.Error() != "permission denied" {
			return &FuzzResult{
				ID:           fmt.Sprintf("random_tx_%d", time.Now().UnixNano()),
				Failed:       true,
				ErrorMessage: err.Error(),
				Crashed:      true,
			}
		}
	}

	if output != nil && len(output) > 0 {
		// Check for specific error indicators in output
		if string(output) == "state_inconsistent" {
			return &FuzzResult{
				ID:                 fmt.Sprintf("random_tx_%d", time.Now().UnixNano()),
				Failed:             true,
				StateInconsistency: true,
				ErrorMessage:       "State inconsistency detected",
			}
		} else if string(output) == "consensus_failure" {
			return &FuzzResult{
				ID:               fmt.Sprintf("random_tx_%d", time.Now().UnixNano()),
				Failed:           true,
				ConsensusFailure: true,
				ErrorMessage:     "Consensus failure detected",
			}
		}
	}

	return nil
}

// BoundaryValueMutator focuses on edge cases and boundary values
type BoundaryValueMutator struct {
	BaseMutator
}

func NewBoundaryValueMutator(r *rand.Rand) *BoundaryValueMutator {
	return &BoundaryValueMutator{
		BaseMutator: BaseMutator{
			rand: r,
			name: "BoundaryValueMutator",
		},
	}
}

func (m *BoundaryValueMutator) GenerateFuzzInput() []byte {
	// Choose a boundary test case
	testCase := m.rand.Intn(6)

	var input []byte

	switch testCase {
	case 0:
		// Empty input
		input = []byte{}
	case 1:
		// Single byte
		input = []byte{byte(m.rand.Intn(256))}
	case 2:
		// Max uint64
		input = make([]byte, 16)
		binary.LittleEndian.PutUint64(input[0:8], 0xFFFFFFFFFFFFFFFF)
		binary.LittleEndian.PutUint64(input[8:16], 0xFFFFFFFFFFFFFFFF)
	case 3:
		// Zero values
		length := 16 + m.rand.Intn(100)
		input = make([]byte, length)
		// Keep all zeros
	case 4:
		// Negative values (for int64)
		input = make([]byte, 16)
		binary.LittleEndian.PutUint64(input[0:8], 0x8000000000000000) // -2^63
	case 5:
		// Very large buffer
		length := 1024 * 1024 * (1 + m.rand.Intn(5)) // 1-5 MB
		input = make([]byte, length)
		m.rand.Read(input)
	}

	return input
}

func (m *BoundaryValueMutator) ValidateOutput(output []byte, err error) *FuzzResult {
	// Similar validation logic as RandomTxMutator
	if err != nil {
		if err.Error() != "invalid arguments" && err.Error() != "permission denied" {
			return &FuzzResult{
				ID:           fmt.Sprintf("boundary_%d", time.Now().UnixNano()),
				Failed:       true,
				ErrorMessage: err.Error(),
				Crashed:      true,
			}
		}
	}

	if output != nil && len(output) > 0 {
		if string(output) == "state_inconsistent" {
			return &FuzzResult{
				ID:                 fmt.Sprintf("boundary_%d", time.Now().UnixNano()),
				Failed:             true,
				StateInconsistency: true,
				ErrorMessage:       "State inconsistency detected with boundary value",
			}
		} else if string(output) == "consensus_failure" {
			return &FuzzResult{
				ID:               fmt.Sprintf("boundary_%d", time.Now().UnixNano()),
				Failed:           true,
				ConsensusFailure: true,
				ErrorMessage:     "Consensus failure detected with boundary value",
			}
		}
	}

	return nil
}

// StatefulMutator generates sequences of operations to test state transitions
type StatefulMutator struct {
	BaseMutator
	currentState []byte
}

func NewStatefulMutator(r *rand.Rand) *StatefulMutator {
	return &StatefulMutator{
		BaseMutator: BaseMutator{
			rand: r,
			name: "StatefulMutator",
		},
		currentState: make([]byte, 0),
	}
}

func (m *StatefulMutator) GenerateFuzzInput() []byte {
	// Generate a stateful sequence
	// First byte is the operation type
	opType := byte(m.rand.Intn(4))

	// Create an input that includes current state
	var input []byte

	// Add operation type
	input = append(input, opType)

	// Add a random header (4 bytes)
	header := make([]byte, 4)
	m.rand.Read(header)
	input = append(input, header...)

	// Add current state if we have any
	if len(m.currentState) > 0 {
		input = append(input, m.currentState...)
	}

	// Add some random data
	extraLen := 8 + m.rand.Intn(92)
	extra := make([]byte, extraLen)
	m.rand.Read(extra)
	input = append(input, extra...)

	// Update the current state based on this input
	// In a real implementation, this would be more sophisticated
	// based on the actual response from the module
	m.currentState = append(m.currentState, input[0])
	if len(m.currentState) > 64 {
		// Prevent state from growing too large
		m.currentState = m.currentState[len(m.currentState)-64:]
	}

	return input
}

func (m *StatefulMutator) ValidateOutput(output []byte, err error) *FuzzResult {
	// For stateful tests, focus on state inconsistencies
	if output != nil && len(output) > 0 {
		if string(output) == "state_inconsistent" {
			return &FuzzResult{
				ID:                 fmt.Sprintf("stateful_%d", time.Now().UnixNano()),
				Failed:             true,
				StateInconsistency: true,
				ErrorMessage:       "State transition inconsistency detected",
				Input:              m.currentState, // Save the state that caused the issue
			}
		} else if string(output) == "consensus_failure" {
			return &FuzzResult{
				ID:               fmt.Sprintf("stateful_%d", time.Now().UnixNano()),
				Failed:           true,
				ConsensusFailure: true,
				ErrorMessage:     "Consensus failure in state transition sequence",
				Input:            m.currentState,
			}
		}
	}

	if err != nil && err.Error() != "invalid arguments" && err.Error() != "permission denied" {
		return &FuzzResult{
			ID:           fmt.Sprintf("stateful_%d", time.Now().UnixNano()),
			Failed:       true,
			ErrorMessage: fmt.Sprintf("Unexpected error in state sequence: %s", err.Error()),
			Crashed:      true,
			Input:        m.currentState,
		}
	}

	return nil
}

// SpecialCasesMutator implements known edge cases for Cosmos SDK
type SpecialCasesMutator struct {
	BaseMutator
	cases [][]byte
	index int
}

func NewSpecialCasesMutator(r *rand.Rand) *SpecialCasesMutator {
	// Predefined test cases for common Cosmos SDK issues
	specialCases := [][]byte{
		// Empty message with valid signature
		{0x1, 0x0, 0x0, 0x0, 0xAA, 0xBB, 0xCC, 0xDD},

		// Invalid account sequence
		{0x2, 0xFF, 0xFF, 0xFF, 0xFF, 0x0, 0x0, 0x0},

		// Double spend attempt
		{0x3, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0},

		// State rollback attempt
		{0x4, 0xF0, 0xF0, 0xF0, 0xF0, 0x0, 0x0, 0x0, 0x0},

		// Governance proposal with bad parameters
		{0x5, 0x1, 0x0, 0x0, 0x0, 0xFF, 0xFF, 0xFF, 0xFF},

		// Validator power overflow
		{0x6, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},

		// Module account misuse
		{0x7, 0xA, 0xB, 0xC, 0xD, 0x1, 0x0, 0x0, 0x0},
	}

	return &SpecialCasesMutator{
		BaseMutator: BaseMutator{
			rand: r,
			name: "SpecialCasesMutator",
		},
		cases: specialCases,
		index: 0,
	}
}

func (m *SpecialCasesMutator) GenerateFuzzInput() []byte {
	// Either use a predefined case or mutate one
	useRaw := m.rand.Intn(100) < 75 // 75% chance to use raw special case

	var baseCase []byte
	if useRaw {
		// Select a special case
		baseCase = m.cases[m.rand.Intn(len(m.cases))]
	} else {
		// Mutate a special case
		baseCase = make([]byte, len(m.cases[m.rand.Intn(len(m.cases))]))
		copy(baseCase, m.cases[m.rand.Intn(len(m.cases))])

		// Apply some mutations
		mutations := 1 + m.rand.Intn(5)
		for i := 0; i < mutations; i++ {
			pos := m.rand.Intn(len(baseCase))
			baseCase[pos] = byte(m.rand.Intn(256))
		}
	}

	// Add some padding
	paddingLen := m.rand.Intn(100)
	padding := make([]byte, paddingLen)
	m.rand.Read(padding)

	return append(baseCase, padding...)
}

func (m *SpecialCasesMutator) ValidateOutput(output []byte, err error) *FuzzResult {
	// For special cases, we're particularly interested in crashes
	if err != nil {
		// Ignore expected validation errors
		if err.Error() != "invalid arguments" && err.Error() != "permission denied" {
			return &FuzzResult{
				ID:           fmt.Sprintf("special_%d", time.Now().UnixNano()),
				Failed:       true,
				ErrorMessage: fmt.Sprintf("Special case triggered error: %s", err.Error()),
				Crashed:      true,
			}
		}
	}

	if output != nil && len(output) > 0 {
		if string(output) == "state_inconsistent" {
			return &FuzzResult{
				ID:                 fmt.Sprintf("special_%d", time.Now().UnixNano()),
				Failed:             true,
				StateInconsistency: true,
				ErrorMessage:       "Special case triggered state inconsistency",
			}
		} else if string(output) == "consensus_failure" {
			return &FuzzResult{
				ID:               fmt.Sprintf("special_%d", time.Now().UnixNano()),
				Failed:           true,
				ConsensusFailure: true,
				ErrorMessage:     "Special case triggered consensus failure",
			}
		}
	}

	return nil
}

// // FuzzResult structure repeated here to make the file self-contained
// type FuzzResult struct {
// 	ID                 string
// 	Input              []byte
// 	ErrorMessage       string
// 	Failed             bool
// 	StateInconsistency bool
// 	ConsensusFailure   bool
// 	Crashed            bool
// }
