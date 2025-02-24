#!/bin/bash

# Initialize score
score=0
source ./scripts/setOrgPeerContext.sh 1
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${PWD}/configtx

# Query the product
echo "Querying the product..."
QUERY_OUTPUT=$(peer chaincode query -C mychannel -n supplychain -c '{"Args":["GetProduct","prod1"]}' 2>&1)

# Check if the product exists
if [[ $QUERY_OUTPUT == *"Error"* ]]; then
    echo "Product not found. It was not created."
    # Output the final score
    echo "Final Score: $score/50"
    exit 0
else
    echo "Product found."
fi

# Check if the product owner is SupplierA (indicating creation)
if [[ $QUERY_OUTPUT == *"SupplierA"* ]]; then
    echo "Product creation successful. Supplier is SupplierA."
    score=$((score + 25))
fi

# Check if the product owner is DistributorA (indicating transfer)
if [[ $QUERY_OUTPUT == *"DistributorA"* ]]; then
    echo "Product transfer successful. New owner is DistributorA."
    score=$((score + 50))
fi

# Final score output
echo "Final Score: $score/50"

# Exit with success
exit 0
