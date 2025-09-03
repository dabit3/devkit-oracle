#!/bin/bash

# Price Oracle AVS - Get Token Price Script
# Usage: ./get-price.sh <token_id>
# Example: ./get-price.sh bitcoin
#          ./get-price.sh ethereum

# Check if token ID is provided
if [ $# -eq 0 ]; then
    echo "Error: Token ID is required"
    echo "Usage: $0 <token_id>"
    echo ""
    echo "Examples:"
    echo "  $0 bitcoin"
    echo "  $0 ethereum"
    echo "  $0 chainlink"
    echo ""
    echo "Popular token IDs:"
    echo "  - bitcoin"
    echo "  - ethereum"
    echo "  - binancecoin"
    echo "  - ripple"
    echo "  - cardano"
    echo "  - solana"
    echo "  - polkadot"
    echo "  - dogecoin"
    echo "  - avalanche-2"
    echo "  - chainlink"
    echo "  - polygon"
    echo "  - arbitrum"
    echo "  - optimism"
    exit 1
fi

TOKEN_ID="$1"

echo "Fetching price for token: $TOKEN_ID"
echo "----------------------------------------"

# Call the AVS with the token ID as a string parameter
# The signature indicates we're passing a string
# The args contain the token ID
devkit avs call signature="(string)" args="(\"$TOKEN_ID\")"

echo "----------------------------------------"
echo "Price fetch complete!"