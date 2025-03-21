# Start from the official Golang image for building
FROM golang:1.23-alpine AS builder

# Use the same Dockerfile for both apps
ARG PROGRAM_NAME

# Set the working directory inside the container
WORKDIR /app

# Copy shared lib first
COPY shared/ shared/

# Download libraries
COPY $PROGRAM_NAME/go.mod $PROGRAM_NAME/go.sum $PROGRAM_NAME/
WORKDIR $PROGRAM_NAME
RUN go mod download

# Compile the program
COPY $PROGRAM_NAME/ .
RUN CGO_ENABLED=0 GOOS=linux go build

# Start from the scratch image for the final stage
FROM scratch

ARG PROGRAM_NAME

COPY --from=builder /app/$PROGRAM_NAME/$PROGRAM_NAME /app

EXPOSE 8080

ENTRYPOINT ["/app"]
