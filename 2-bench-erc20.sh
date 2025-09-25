#!/bin/bash

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "Error: .env file does not exist, please run 1-setup.sh first"
    exit 1
fi

# Read contract address from .env file
source .env

# Check if contract address was successfully read
if [ -z "$CONTRACT_ADDRESS" ]; then
    echo "Error: Failed to read CONTRACT_ADDRESS from .env file"
    echo ".env file content:"
    cat .env
    exit 1
fi

echo "Using contract address: $CONTRACT_ADDRESS"

# Execute ERC20 benchmark test
adventure evm bench erc20 --f ./config/poly_test/fork6_erc20.json --contract $CONTRACT_ADDRESS
