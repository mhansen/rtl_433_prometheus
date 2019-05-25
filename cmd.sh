#!/bin/bash

# Important: quit when anything in the pipe fails
set -e
set -o pipefail
rtl_433 -F json -R 19 -R 127 | /rtl_433_prometheus
