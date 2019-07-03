FROM golang:alpine as gobuilder

RUN apk update && apk add git && apk add ca-certificates

WORKDIR /root
RUN mkdir /root/app
COPY go.mod go.sum *.go /root/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -a rtl_433_prometheus.go

FROM debian:latest as cbuilder
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

FROM debian:latest
RUN apt-get update && apt-get install -y librtlsdr0
WORKDIR /
COPY --from=gobuilder /root/rtl_433_prometheus /
COPY --from=cbuilder /usr/local/bin/rtl_433 /
RUN chmod +x /rtl_433
EXPOSE 9001
ENTRYPOINT ["/rtl_433_prometheus"]
CMD ["--subprocess", "/rtl_433 -F json -M newmodel"]

