# 1. Use official Go image
FROM golang:1.23.6

# 2. Set working directory inside container
WORKDIR /app

# 3. Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# 4. Copy the rest of the application source code
COPY . .

# 5. Build the Go application
RUN go build -o main .

# 6. Run the application
CMD ["./main"]
