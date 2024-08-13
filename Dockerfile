# May be built from x86_64, using cross-build-start magic.

# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.23 as gobuilder

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

FROM gcr.io/rtl433/rtl_433:latest as rtl_433
FROM balenalib/raspberrypi3:run

# https://www.balena.io/docs/reference/base-images/base-images/#building-arm-containers-on-x86-machines
RUN [ "cross-build-start" ]

RUN apt-get update && apt-get install -y librtlsdr0

WORKDIR /
COPY --from=gobuilder /app/rtl_433_prometheus /
COPY --from=rtl_433 /usr/local/bin/rtl_433 /
RUN chmod +x /rtl_433

# https://www.balena.io/docs/reference/base-images/base-images/#building-arm-containers-on-x86-machines
RUN [ "cross-build-end" ]

EXPOSE 9550
ENTRYPOINT ["/rtl_433_prometheus"]
CMD ["--subprocess", "/rtl_433 -F json -M newmodel"]

