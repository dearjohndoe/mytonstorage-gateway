#!/bin/bash

cd /opt/storage

env $(cat config.env | xargs) ./mtpo-gateway >> /var/log/mytonstorage_gateway.app/mytonstorage_gateway.app.log 2>&1 &

sleep 5

if pgrep -f "./mtpo-gateway" > /dev/null; then
    echo "✅ Gateway application started successfully."
else
    echo "❌ Failed to start gateway application."
    exit 1
fi
