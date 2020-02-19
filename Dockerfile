# Must be built on an ARM device, else you'll get invalid architecture.

# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.13 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -mod=readonly -a -v rtl_433_prometheus.go

# I'd like to use arm32v6 for RPi Zero W but that doesn't exist on Docker Hub. v5 is good enough.
FROM arm32v5/debian:latest as cbuilder
RUN apt-get update && apt-get install -y git libusb-1.0.0-dev librtlsdr-dev rtl-sdr cmake automake
WORKDIR /tmp/
RUN git clone https://github.com/mhansen/rtl_433.git && \
    cd rtl_433 && \
    mkdir build && \
    cd build && \
    cmake ../ && \
    make && \
    make install && \
    cd / && \
    rm -rf /tmp

# I'd like to use arm32v6 for RPi Zero W but that doesn't exist on Docker Hub. v5 is good enough.
FROM arm32v5/debian:latest
RUN apt-get update && apt-get install -y librtlsdr0
WORKDIR /
COPY --from=gobuilder /root/rtl_433_prometheus /
COPY --from=cbuilder /usr/local/bin/rtl_433 /
RUN chmod +x /rtl_433
EXPOSE 9550
ENTRYPOINT ["/rtl_433_prometheus"]
CMD ["--subprocess", "/rtl_433 -F json -M newmodel"]

