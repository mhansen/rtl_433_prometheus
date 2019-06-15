FROM golang:alpine as builder

RUN apk update && apk add git && apk add ca-certificates

WORKDIR /root
RUN mkdir /root/app
COPY go.mod go.sum *.go /root/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -a rtl_433_prometheus.go

FROM debian:latest
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
WORKDIR /

COPY --from=builder /root/rtl_433_prometheus /
EXPOSE 9001
ENTRYPOINT ["/rtl_433_prometheus", "--subprocess"]
CMD ["rtl_433 -F json"]

