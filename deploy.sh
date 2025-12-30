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

# Add key to ssh-agent
ssh-add k8s-key.pem
echo "Key added to ssh-agent."

# Get Jumpbox Public IP
JUMPBOX_IP=$(pulumi stack output jumpboxPublicIp)

echo ""
echo "To connect to the jumpbox, run:"
echo "ssh admin@$JUMPBOX_IP"
