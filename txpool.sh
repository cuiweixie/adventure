#!/bin/bash

curl --location 'http://127.0.0.1:8545' \
--header 'Content-Type: application/json' \
--data '{
    "jsonrpc": "2.0",
    "method": "txpool_status",
    "params": [],
    "id": 1
}'
