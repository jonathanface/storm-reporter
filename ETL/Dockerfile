# Use the official Golang image
FROM golang:1.22 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN go build -o etl-service .

# Expose port for debugging/logging (optional)
EXPOSE 3000

# Command to run the ETL service
CMD ["./etl-service"]
