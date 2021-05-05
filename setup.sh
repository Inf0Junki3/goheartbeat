#!/bin/bash

# Build the heartbeat service as a binary
go build heartbeat.go

# Copy heartbeat to /bin and heartbeat.config to /etc
if test -f /bin/heartbeat; then
    echo "/bin/heartbeat exists! Please remove this file before proceeding."
    exit 255
else
    mv heartbeat /bin/
fi

if test -f /etc/heartbeat.config; then
    echo "/etc/heartbeat.config exists! Please remove this file before proceeding."
    exit 255
else
    cp heartbeat.config /etc/
fi

# Create, enable and start the service
if test -f /etc/systemd/system/heartbeat.service; then
    echo "/etc/systemd/system/heartbeat.service exists! Please remove this file before proceeding."
    exit 255
else
    cp heartbeat.service /etc/systemd/system/
fi

systemctl enable heartbeat
systemctl start heartbeat
exit 0