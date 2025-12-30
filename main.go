package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a VPC for Kubernetes the Hard Way
		vpc, err := ec2.NewVpc(ctx, "kubernetes-vpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.240.0.0/16"),
			EnableDnsSupport:   pulumi.Bool(true),
			EnableDnsHostnames: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes-the-hard-way"),
			},
		})
		if err != nil {
			return err
		}

		// Create a Subnet
		subnet, err := ec2.NewSubnet(ctx, "kubernetes-subnet", &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String("10.240.0.0/24"),
			AvailabilityZone:    pulumi.String("us-east-1a"),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes"),
			},
		})
		if err != nil {
			return err
		}

		// Create an Internet Gateway
		igw, err := ec2.NewInternetGateway(ctx, "kubernetes-igw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes"),
			},
		})
		if err != nil {
			return err
		}

		// Create a Route Table
		rt, err := ec2.NewRouteTable(ctx, "kubernetes-rt", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: igw.ID(),
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes"),
			},
		})
		if err != nil {
			return err
		}

		// Associate Route Table with Subnet
		_, err = ec2.NewRouteTableAssociation(ctx, "kubernetes-rta", &ec2.RouteTableAssociationArgs{
			SubnetId:     subnet.ID(),
			RouteTableId: rt.ID(),
		})
		if err != nil {
			return err
		}

		// Create a Security Group
		sg, err := ec2.NewSecurityGroup(ctx, "kubernetes-sg", &ec2.SecurityGroupArgs{
			VpcId:       vpc.ID(),
			Description: pulumi.String("Kubernetes security group"),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Protocol:    pulumi.String("-1"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					CidrBlocks:  pulumi.StringArray{pulumi.String("10.240.0.0/24")},
					Description: pulumi.String("Allow internal communication"),
				},
				&ec2.SecurityGroupIngressArgs{
					Protocol:    pulumi.String("tcp"),
					FromPort:    pulumi.Int(22),
					ToPort:      pulumi.Int(22),
					CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					Description: pulumi.String("Allow SSH"),
				},
				&ec2.SecurityGroupIngressArgs{
					Protocol:    pulumi.String("tcp"),
					FromPort:    pulumi.Int(6443),
					ToPort:      pulumi.Int(6443),
					CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					Description: pulumi.String("Allow Kubernetes API"),
				},
				&ec2.SecurityGroupIngressArgs{
					Protocol:    pulumi.String("icmp"),
					FromPort:    pulumi.Int(-1),
					ToPort:      pulumi.Int(-1),
					CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
					Description: pulumi.String("Allow ICMP"),
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes"),
			},
		})
		if err != nil {
			return err
		}

		// Create SSH Key Pair
		sshKey, err := tls.NewPrivateKey(ctx, "k8s-ssh-key", &tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.Int(4096),
		})
		if err != nil {
			return err
		}

		keyPair, err := ec2.NewKeyPair(ctx, "k8s-keypair", &ec2.KeyPairArgs{
			PublicKey: sshKey.PublicKeyOpenssh,
			Tags: pulumi.StringMap{
				"Name": pulumi.String("kubernetes"),
			},
		})
		if err != nil {
			return err
		}

		// Export the private key
		ctx.Export("privateKey", sshKey.PrivateKeyOpenssh)
		ctx.Export("publicKey", sshKey.PublicKeyOpenssh)

		// Lookup Debian 12 AMI (ARM64)
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			MostRecent: pulumi.BoolRef(true),
			Filters: []ec2.GetAmiFilter{
				{
					Name:   "name",
					Values: []string{"debian-12-arm64-*"},
				},
				{
					Name:   "architecture",
					Values: []string{"arm64"},
				},
				{
					Name:   "virtualization-type",
					Values: []string{"hvm"},
				},
				{
					Name:   "owner-id",
					Values: []string{"136693071363"},
				},
			},
		})
		if err != nil {
			return err
		}

		// Define instances
		instances := []struct {
			Name      string
			Type      string
			DiskSize  int
			PrivateIP string
		}{
			{"jumpbox", "t4g.small", 10, "10.240.0.10"},
			{"server", "t4g.small", 20, "10.240.0.11"},
			{"node-0", "t4g.small", 20, "10.240.0.20"},
			{"node-1", "t4g.small", 20, "10.240.0.21"},
		}

		for _, instance := range instances {
			inst, err := ec2.NewInstance(ctx, instance.Name, &ec2.InstanceArgs{
				Ami:                      pulumi.String(ami.Id),
				InstanceType:             pulumi.String(instance.Type),
				SubnetId:                 subnet.ID(),
				VpcSecurityGroupIds:      pulumi.StringArray{sg.ID()},
				PrivateIp:                pulumi.String(instance.PrivateIP),
				AssociatePublicIpAddress: pulumi.Bool(true),
				KeyName:                  keyPair.KeyName,
				RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
					VolumeSize: pulumi.Int(instance.DiskSize),
					VolumeType: pulumi.String("gp3"),
				},
				Tags: pulumi.StringMap{
					"Name": pulumi.String(instance.Name),
				},
			})
			if err != nil {
				return err
			}

			ctx.Export(instance.Name+"PublicIp", inst.PublicIp)
			ctx.Export(instance.Name+"PrivateIp", inst.PrivateIp)
		}

		return nil
	})
}
