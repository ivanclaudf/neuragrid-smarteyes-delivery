# Use Amazon Linux 2023 as the base image
FROM amazonlinux:2023

# Define build arguments with defaults
ARG GOARCH=amd64
ARG GOOS=linux

# Install required dependencies
RUN yum update -y && \
    yum install -y \
    gcc \
    git \
    tar \
    wget && \
    yum clean all

# Download Go based on target architecture
RUN curl -O https://dl.google.com/go/go1.25.0.linux-${GOARCH}.tar.gz && \
    tar -C /usr/local -xzf go1.25.0.linux-${GOARCH}.tar.gz

ENV GOROOT /usr/local/go
ENV GOPATH /var/www/go
ENV PATH ${PATH}:/var/www/go/bin:/usr/local/go/bin


# Set the working directory
WORKDIR /app

# Copy the Go modules and source code
COPY src ./

# Build the application
# Use build arguments for architecture
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}

# Add error checking to build process
RUN go mod tidy && \
    echo "Building application..." && \
    go build -v -o main . && \
    echo "Build completed." && \
    ls -la && \
    test -f main || (echo "ERROR: main binary was not created. Build failed." && exit 1)

# Expose the application port
EXPOSE 8080

# Command to keep the container running for debugging
CMD ["sleep", "infinity"]
