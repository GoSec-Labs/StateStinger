#!/bin/bash

# StateStinger - Cosmos SDK State Machine Fuzzer
# Usage: ./bensh.sh [options] <target_module_path>

# Default values
FUZZ_COUNT=5000
OUTPUT_DIR="./results"
VERBOSE=false
SPECIAL=true
MODULE=""

# Show help function
show_help() {
    echo "StateStinger - Fuzzing Engine for Cosmos SDK State Machines"
    echo ""
    echo "Usage: $0 [options] <target_module_path>"
    echo ""
    echo "Options:"
    echo "  -c, --count NUM       Number of fuzzing iterations (default: 5000)"
    echo "  -o, --output DIR      Output directory for results (default: ./results)"
    echo "  -m, --module NAME     Module name (default: inferred from path)"
    echo "  -v, --verbose         Enable verbose output"
    echo "  -s, --seed NUM        Random seed (default: time-based)"
    echo "  --no-special          Disable special case testing"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 --count 10000 --verbose ../cosmos-sdk/x/bank"
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--count)
            FUZZ_COUNT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -m|--module)
            MODULE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -s|--seed)
            SEED="$2"
            shift 2
            ;;
        --no-special)
            SPECIAL=false
            shift
            ;;
        -h|--help)
            show_help
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Check for required arguments
if [ -z "$TARGET" ]; then
    echo "Error: Target module path is required"
    show_help
fi

# Build command arguments
ARGS="--target $TARGET --count $FUZZ_COUNT --output $OUTPUT_DIR"

if [ -n "$MODULE" ]; then
    ARGS="$ARGS --module $MODULE"
fi

if [ "$VERBOSE" = true ]; then
    ARGS="$ARGS --verbose"
fi

if [ -n "$SEED" ]; then
    ARGS="$ARGS --seed $SEED"
fi

if [ "$SPECIAL" = false ]; then
    ARGS="$ARGS --no-special"
fi

# Execute the fuzzing command
echo "Running StateStinger with the following arguments:"
echo "$ARGS"

# Assuming the fuzzing engine is a binary named `state_stinger`
./state_stinger $ARGS

# Check the exit status of the fuzzing engine
if [ $? -eq 0 ]; then
    echo "Fuzzing completed successfully."
else
    echo "Fuzzing encountered an error."
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Display run info
echo "=== StateStinger Fuzzing Run ==="
echo "Target: $TARGET"
echo "Module: ${MODULE:-"(auto-detect)"}"
echo "Iterations: $FUZZ_COUNT"
echo "Output: $OUTPUT_DIR"
echo "Verbose: $VERBOSE"
echo "Special Cases: $SPECIAL"
echo "=========================="

# Run the StateStinger command
echo "Starting fuzzing run..."
cd "$(dirname "$0")/.." || exit 1
go run ./cmd/statestinger $ARGS

# Exit with the same code as the command
exit $?