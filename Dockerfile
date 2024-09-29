# Start from the official Go image
FROM --platform=linux/arm64 golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Install wget and unzip
RUN apt-get update && apt-get install -y wget unzip

# Download and install ONNX Runtime
RUN wget https://github.com/microsoft/onnxruntime/releases/download/v1.19.2/onnxruntime-linux-aarch64-1.19.2.tgz && \
    tar -xzf onnxruntime-linux-aarch64-1.19.2.tgz && \
    mv onnxruntime-linux-aarch64-1.19.2 /opt/onnxruntime && \
    rm onnxruntime-linux-aarch64-1.19.2.tgz

# Set the ONNX_PATH environment variable
ENV ONNX_PATH=/opt/onnxruntime/lib/libonnxruntime.so

# Copy the Go module files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]