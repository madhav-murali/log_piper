#FROM golang:1.21-alpine

#WORKDIR /app

#COPY go.mod go.sum ./
#RUN go mod download

#COPY . .
#RUN go build -o ingestor ./cmd/ingestor

#CMD ["./ingestor"]

# --- Build Stage ---
# Use the Go image to build the binary
FROM golang:1.21-alpine AS builder
WORKDIR /src

# Copy module files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app/ingestor .

# --- Final Stage ---
# Use a minimal image for the final container
FROM alpine:latest
WORKDIR /app

# Copy *only* the compiled binary from the builder stage
COPY --from=builder /app/ingestor .

# (Optional) If your app needs SSL certificates
RUN apk --no-cache add ca-certificates

# Command to run the app
CMD ["./ingestor"]
