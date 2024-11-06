FROM golang:1.20-alpine

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create and set the working directory
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o main .

# Make the run.sh script executable
RUN chmod +x /app/run.sh

# Expose the port the app runs on
EXPOSE 8080

# Start the application
CMD ["/app/run.sh"]
