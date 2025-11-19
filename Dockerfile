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


# Set the working directory for building
WORKDIR /build

# Copy the Go modules and source code
COPY src ./

# Build the application
# Use build arguments for architecture
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o delivery main.go

# Set the working directory for runtime
WORKDIR /app

# Copy the built executable to a location that won't be overridden by volume mount
RUN cp /build/delivery /usr/local/bin/delivery && chmod +x /usr/local/bin/delivery

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["delivery"]
