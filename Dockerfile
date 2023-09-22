# Use a golang base image to build the Go project
FROM golang:latest AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go project files into the container
COPY . .

# Build the Go project for various platforms
RUN GOOS=linux GOARCH=amd64 go build -o ./bin/croned-twitcasting-recorder-mp4_linux_amd64

# Use a new image as a runtime container
FROM debian:bullseye-slim

# Install ffmpeg using apt-get
RUN apt-get update \
    && apt-get install -y ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Copy the built Go binaries from the builder stage into the runtime container
COPY --from=builder /app/bin/ /app/

# Set the working directory inside the container
WORKDIR /app

# Define the entry point for your application (modify as needed)
CMD ["./croned-twitcasting-recorder-mp4_linux_amd64"]
