FROM golang:1.23-alpine AS builder

WORKDIR /app


COPY . .
RUN go mod download
RUN go build -o main .

# Start a new minimal image
FROM alpine:latest

# Set the working directory in the new image
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./main"]
