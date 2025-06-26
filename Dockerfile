# Stage 1: Build the application
FROM docker.io/library/golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum go.work* ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Install make
RUN apk add --no-cache make

# Build the statically linked binary
RUN make build-linux

# Stage 2: Create the final, minimal image
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /app/gitversion-go /gitversion-go

# Set the entrypoint
ENTRYPOINT ["/gitversion-go"]
