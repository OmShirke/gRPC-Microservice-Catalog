# Use a Golang image for building the application
FROM golang:1.23-alpine AS build

# Install necessary dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY gRPC-Catalog-service/go.mod gRPC-Catalog-service/go.sum ./

# Copy the source code
COPY gRPC-Catalog-service/ ./

# Download the Go modules dependencies (will use the vendor directory)
RUN GO111MODULE=on go mod tidy
RUN GO111MODULE=on go mod vendor

# Build the application binary
RUN go build -o /app/bin/catalog ./cmd/catalog

# Final stage: create a smaller image to run the application
FROM alpine:3.11

# Set the working directory inside the container
WORKDIR /usr/bin

# Copy the compiled binary from the build stage
COPY --from=build /app/bin/catalog .

# Expose the port your application runs on
EXPOSE 8080

# Command to run the application
CMD ["./catalog"]

