# Use an official Golang image as a base
FROM golang:1.21-alpine

# Set working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o api-service

# Expose the API port
EXPOSE 8080

# Run the application
CMD ["./api-service"]
