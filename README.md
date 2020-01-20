A Prometheus exporter for radio messages received from rtl_433.

Hosted on Docker Hub: https://hub.docker.com/r/markhnsn/rtl_433_prometheus

You can configure locations using the name+channel like this:

     ./rtl_433_prometheus --channel_matcher=Acurite-Tower,1,Bedroom --channel_matcher=Acurite-Tower,2,Downstairs

And using name+ID like this:

     ./rtl_433_prometheus --id_matcher=Acurite-Tower,12345,Bedroom --id_matcher=Acurite-Tower,23456,Downstairs

You can also combine.


Example docker-compose.yml config:

```
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
