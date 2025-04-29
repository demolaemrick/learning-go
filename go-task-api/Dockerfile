# Use the official Go 1.23 image with Alpine Linux as the base for the build stage, named "builder"
# - golang:1.23-alpine is a lightweight image (~300MB) with Go tools pre-installed
# - Alpine Linux keeps the image small and secure
# - "AS builder" names this stage for later reference in multi-stage builds
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container to /app
# - All subsequent commands (COPY, RUN, etc.) operate relative to /app
# - Creates /app if it doesn’t exist
WORKDIR /app

# Copy go.mod and go.sum files to the working directory (/app)
# - go.mod defines the Go module and dependencies
# - go.sum ensures dependency integrity
# - Copying these first allows caching of dependencies for faster rebuilds
COPY go.mod go.sum ./

# Download Go module dependencies specified in go.mod
# - Runs "go mod download" to fetch dependencies (e.g., jwt/v5, gorilla/mux)
# - Stores them in the container’s module cache
# - Cached unless go.mod/go.sum changes, speeding up builds
RUN go mod download

# Copy all files from the current host directory (project root) to /app
# - Includes main.go, middleware.go, handlers/*.go, and any other files
# - .dockerignore (if present) excludes files like .git, *.md
# - Overwrites go.mod, go.sum from earlier COPY (no effect, just simpler)
COPY . .

# Compile the Go application into a static binary named "task-api"
# - CGO_ENABLED=0 disables C bindings for a fully static binary (no external libs)
# - GOOS=linux ensures compatibility with Linux-based containers (e.g., Alpine)
# - go build -o /task-api compiles main.go, middleware.go, handlers/*.go
# - Output binary is placed at /task-api in the container
# - . specifies source files (all .go files in /app)
RUN CGO_ENABLED=0 GOOS=linux go build -o /task-api .

# Start a new stage with the minimal Alpine Linux image
# - alpine:latest is tiny (~5MB), reducing the final image size
# - Discards the Go toolchain from the builder stage
# - Only includes what’s needed to run the binary
FROM alpine:latest

# Set the working directory to /root/ in the final image
# - /root/ is a standard location for the binary
# - Keeps the file structure clean
WORKDIR /root/

# Copy the compiled binary from the builder stage to /root/task-api
# - --from=builder refers to the first stage (golang:1.23-alpine)
# - Copies only the /task-api binary, keeping the image minimal (~12MB total)
COPY --from=builder /task-api .

# Inform Docker that the container listens on port 9000
# - Matches PORT := ":9000" in main.go
# - Doesn’t actually publish the port (use -p flag when running)
# - Helps documentation and tools like Railway/Fly.io
EXPOSE 9000

# Specify the command to run when the container starts
# - ["./task-api"] executes the compiled binary
# - Runs your API server, listening on port 9000
# - Uses exec format (["..."]) for proper signal handling
CMD ["./task-api"]