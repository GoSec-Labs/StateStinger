FROM golang:1.24-alpine

LABEL maintainer="StateStinger Team"
LABEL description="Container for StateStinger - Cosmos SDK State Machine Fuzzer"

# Install build dependencies
RUN apk add --no-cache \
    bash \
    git \
    make \
    grep \
    gcc \
    libc-dev

# Set up working directory
WORKDIR /app

# Copy go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the statestinger binary
RUN go build -o /usr/local/bin/statestinger ./cmd/statestinger

# Create a directory for results
RUN mkdir -p /data

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/statestinger"]

# Default command (can be overridden)
CMD ["--help"]

# Usage instructions
# To run with Docker:
# docker build -t statestinger .
# docker run -v $(pwd)/results:/data statestinger --target /path/to/module --output /data