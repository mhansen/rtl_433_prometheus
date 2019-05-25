#!/bin/bash

# Important: quit when anything in the pipe fails
set -e
set -o pipefail
rtl_433 -F json | /rtl_433_prometheus
