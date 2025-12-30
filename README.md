# Kubernetes the Hard Way - Infrastructure

This Pulumi project provisions the infrastructure required for "Kubernetes the Hard Way".

## Infrastructure

It creates:
- A VPC with a public subnet (`10.240.0.0/24`).
- An Internet Gateway and Route Table.
- A Security Group allowing SSH, internal communication, and Kubernetes API access.
- 4 EC2 instances running Debian 12 (ARM64):
  - `jumpbox`: t4g.small (2 vCPU, 2 GB RAM) - _Upgraded for Free Tier eligibility_
  - `server`: t4g.small (2 vCPU, 2 GB RAM)
  - `node-0`: t4g.small (2 vCPU, 2 GB RAM)
  - `node-1`: t4g.small (2 vCPU, 2 GB RAM)

## Prerequisites

- Pulumi CLI
- AWS Credentials configured
- Go 1.20+

## Usage

1. Initialize the stack (if not already done):
   ```bash
   pulumi stack init dev
   ```

2. Set the AWS region:
   ```bash
   pulumi config set aws:region us-west-2
   ```

3. **Deploy using the helper script**:
   We recommend using the provided script, which handles deployment and automatically exports the generated SSH key.
   ```bash
   ./deploy.sh
   ```
   
   Alternatively, you can run `pulumi up` manually, but you will need to extract the private key yourself:
   ```bash
   pulumi up
   pulumi stack output privateKey --show-secrets > k8s-key.pem
   chmod 600 k8s-key.pem
   ```

## SSH Access

The project automatically generates an SSH key pair for the cluster. The private key is saved to `k8s-key.pem` by the `deploy.sh` script.

To access the jumpbox:
```bash
ssh -i k8s-key.pem admin@<JUMPBOX_PUBLIC_IP>
```

From the jumpbox, you can reach internal nodes using their internal IPs (`10.240.0.11`, etc.).