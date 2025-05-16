# Use an official Go runtime as a parent image.
FROM golang:1.24.3 AS build

# Install necessary tools and dependencies.
RUN apt-get update && \
    apt-get install -y \
        git \
        build-essential \
        make

# Create a directory for cloning the repository.
RUN mkdir /app

# Clone the go-continuous-fuzz repo into the /app directory.
RUN git clone https://github.com/NishantBansal2003/LND-Fuzz.git /app

# Change current working directory.
WORKDIR /app

# Install Go modules.
RUN go mod download

# Build the go-continuous-fuzz project.
RUN make build

# By default, run the fuzzing target with `make run`
ENTRYPOINT ["make", "run"]