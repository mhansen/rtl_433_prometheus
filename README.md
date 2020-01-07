A Prometheus exporter for radio messages received from rtl_433.

Hosted on Docker Hub: https://hub.docker.com/r/markhnsn/rtl_433_prometheus

You can configure locations using the name+channel like this:

     ./rtl_433_prometheus --channel_matcher=Acurite-Tower,1,Bedroom --channel_matcher=Acurite-Tower,2,Downstairs

And using name+ID like this:

     ./rtl_433_prometheus --id_matcher=Acurite-Tower,12345,Bedroom --id_matcher=Acurite-Tower,23456,Downstairs

You can also combine.
