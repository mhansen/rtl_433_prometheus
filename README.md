A Prometheus exporter for radio messages received from [rtl_433](https://github.com/merbanan/rtl_433).

Hosted on Docker Hub: https://hub.docker.com/r/markhnsn/rtl_433_prometheus

You can configure locations using the name+channel like this:

```shell
$ ./rtl_433_prometheus --channel_matcher=Acurite-Tower,1,Bedroom --channel_matcher=Acurite-Tower,2,Downstairs
```

And using name+ID like this:

```shell
$ ./rtl_433_prometheus --id_matcher=Acurite-Tower,12345,Bedroom --id_matcher=Acurite-Tower,23456,Downstairs
```

You can also combine `--id_matcher` with `--channel_matcher`.

Example end-to-end usage, showing prometheus plaintext format:

```shell
$ go build
$ ./rtl_433_prometheus \
   --id_matcher "Acurite-Tower,1234,Study Balcony" \
   --id_matcher "Acurite-Tower,4567,Bedroom Balcony North" \
   --id_matcher "Acurite-Tower,7890,Bedroom Balcony South"


$ curl http://localhost:9550/metrics
...
# HELP rtl_433_battery Battery high (1) or low (0).
# TYPE rtl_433_battery gauge
rtl_433_battery{channel="A",id="1234",location="Study Balcony",model="Acurite tower sensor"} 1
rtl_433_battery{channel="B",id="4567",location="Bedroom Balcony North",model="Acurite tower sensor"} 1
rtl_433_battery{channel="C",id="7890",location="Bedroom Balcony South",model="Acurite tower sensor"} 1
# HELP rtl_433_humidity Relative Humidity (0-1.0)
# TYPE rtl_433_humidity gauge
rtl_433_humidity{channel="A",id="1234",location="Study Balcony",model="Acurite tower sensor"} 0.68
rtl_433_humidity{channel="B",id="4567",location="Bedroom Balcony North",model="Acurite tower sensor"} 0.6
rtl_433_humidity{channel="C",id="7890",location="Bedroom Balcony South",model="Acurite tower sensor"} 0.68
# HELP rtl_433_packets_received Packets (temperature messages) received.
# TYPE rtl_433_packets_received counter
rtl_433_packets_received{channel="A",id="1234",location="Study Balcony",model="Acurite tower sensor"} 65200
rtl_433_packets_received{channel="B",id="4567",location="Bedroom Balcony North",model="Acurite tower sensor"} 73440
rtl_433_packets_received{channel="C",id="7890",location="Bedroom Balcony South",model="Acurite tower sensor"} 70727
# HELP rtl_433_temperature_celsius Temperature in Celsius
# TYPE rtl_433_temperature_celsius gauge
rtl_433_temperature_celsius{channel="A",id="1234",location="Study Balcony",model="Acurite tower sensor"} 24.5
rtl_433_temperature_celsius{channel="B",id="4567",location="Bedroom Balcony North",model="Acurite tower sensor"} 25.6
rtl_433_temperature_celsius{channel="C",id="7890",location="Bedroom Balcony South",model="Acurite tower sensor"} 24.5
# HELP rtl_433_timestamp_seconds Timestamp we received the message (Unix seconds)
# TYPE rtl_433_timestamp_seconds gauge
rtl_433_timestamp_seconds{channel="A",id="1234",location="Study Balcony",model="Acurite tower sensor"} 1.5825398643534584e+09
rtl_433_timestamp_seconds{channel="B",id="4567",location="Bedroom Balcony North",model="Acurite tower sensor"} 1.5825398669721768e+09
rtl_433_timestamp_seconds{channel="C",id="7890",location="Bedroom Balcony South",model="Acurite tower sensor"} 1.5825398706426628e+09
```

Example `docker-compose.yml` config:

```yml
version: '3.4'
services:
  rtl_433_prometheus:
    image: markhnsn/rtl_433_prometheus
    restart: always
    ports:
    - "9550:9550"
    devices:
    - "/dev/bus/usb"
    command: [
      "--subprocess", "/rtl_433 -F json -R 19 -R 127 -R 40",
      "--channel_matcher", "Nexus Temperature/Humidity,1,Study",
      "--channel_matcher", "Nexus Temperature/Humidity,2,Bedroom",
      "--channel_matcher", "Nexus Temperature/Humidity,3,Balcony",
      "--id_matcher", "Acurite tower sensor,6543,Dining Room",
      "--id_matcher", "Acurite tower sensor,5432,Kitchen",
      "--id_matcher", "Acurite tower sensor,4321,Balcony",
    ]
```

Example `prometheus.yml`:

```yml
scrape_configs:
  - job_name: 'rtl_433_prometheus'
      static_configs:
            - targets: ['hostname:9550']
```
