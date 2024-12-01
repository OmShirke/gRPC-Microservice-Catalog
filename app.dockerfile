# Build Stage
FROM golang:1.23-alpine AS build

# Install necessary build dependencies
RUN apk --no-cache add gcc g++ make ca-certificates

# Set the working directory
WORKDIR /go/src/github.com/OmShirke/gRPC-Microservice-Catalog

# Copy dependency management files
COPY go.mod go.sum ./

# Ensure dependencies are downloaded
RUN go mod download

# Copy application-specific files
COPY vendor vendor
COPY catalog catalog

# Build the application binary
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./catalog/cmd/catalog

# Runtime Stage
FROM alpine:3.11

# Set the working directory
WORKDIR /usr/bin

# Copy the compiled binary from the build stage
COPY --from=build /go/bin/ .

# Expose the required port
EXPOSE 8080

# Run the service
CMD ["app"]
