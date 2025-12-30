#!/bin/bash
set -e

# Deploy infrastructure
echo "Running pulumi up..."
pulumi up --yes

# Export private key
echo "Exporting SSH private key..."
pulumi stack output privateKey --show-secrets > k8s-key.pem

# Set permissions
chmod 600 k8s-key.pem
echo "Key saved to k8s-key.pem with 600 permissions."
