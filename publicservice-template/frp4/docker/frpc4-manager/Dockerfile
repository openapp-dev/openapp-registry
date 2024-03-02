# Use a Go base image
FROM golang:1.20-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the source code to the working directory
COPY . .

# Download the Go module dependencies
RUN go mod tidy

# Build the Go binary
RUN go build -o frpc4-manager cmd/main.go

# Use a lightweight Alpine base image for the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary from the build stage to the final image
COPY --from=build /app/frpc4-manager .

# Expose any necessary ports
EXPOSE 8080

# Set the command to run the Go binary
CMD ["./frpc4-manager"]
