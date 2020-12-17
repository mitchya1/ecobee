#!/usr/bin/env bash

docker run -p 8086:8086 -v ${HOME}/tmp/influx:/var/lib/influxdb quay.io/influxdb/influxdb:v2.0.3