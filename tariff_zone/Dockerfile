FROM golang:1.18-alpine


# Install necessary dependencies
RUN apk add --no-cache bash jq

# Create and set the working directory
WORKDIR /app

# If you don't have go.mod and go.sum files, skip copying them
# Copy the source code
# Copy Go module files and download dependencies
COPY go.mod ./
RUN go mod download


# Copy the Go source code and build it
COPY main.go .
RUN go build -o main main.go

# Copy the run.sh file and make it executable
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh

# Expose the port the app runs on
EXPOSE 8080

# Start the application
CMD ["/app/run.sh"]
