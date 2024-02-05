package main

import (
	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
		if err != nil {
			return err
		}
		virtualNetwork, err := network.NewVirtualNetwork(ctx, "virtualNetwork", &network.VirtualNetworkArgs{
			ResourceGroupName: resourceGroup.Name,
			AddressSpace: &network.AddressSpaceArgs{
				AddressPrefixes: pulumi.StringArray{
					pulumi.String("10.0.0.0/16"),
				},
			},
		})
		if err != nil {
			return err
		}
		subnet, err := network.NewSubnet(ctx, "subnet", &network.SubnetArgs{
			ResourceGroupName:  resourceGroup.Name,
			VirtualNetworkName: virtualNetwork.Name,
			AddressPrefix:      pulumi.String("10.0.1.0/24"),
		})
		if err != nil {
			return err
		}
		networkInterface, err := network.NewNetworkInterface(ctx, "networkInterface", &network.NetworkInterfaceArgs{
			ResourceGroupName: resourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
				&network.NetworkInterfaceIPConfigurationArgs{
					Name:    pulumi.String("test-ip-config"),
					Primary: pulumi.Bool(true),
					Subnet: &network.SubnetTypeArgs{
						Id: subnet.ID(),
					},
				},
			},
		})
		if err != nil {
			return err
		}
		virtualMachine, err := compute.NewVirtualMachine(ctx, "virtualMachine", &compute.VirtualMachineArgs{
			NetworkProfile: &compute.NetworkProfileArgs{
				NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
					&compute.NetworkInterfaceReferenceArgs{
						Id: networkInterface.ID(),
					},
				},
			},
			HardwareProfile: &compute.HardwareProfileArgs{
				VmSize: pulumi.String("Standard_B2s"),
			},
			OsProfile: &compute.OSProfileArgs{
				ComputerName:  pulumi.String("HelloVM"),
				AdminUsername: pulumi.String("azureuser"),
				AdminPassword: pulumi.String("Password1234!"),
			},
			ResourceGroupName: resourceGroup.Name,
			StorageProfile: &compute.StorageProfileArgs{
				ImageReference: &compute.ImageReferenceArgs{
					Offer:     pulumi.String("0001-com-ubuntu-server-lunar"),
					Publisher: pulumi.String("Canonical"),
					Sku:       pulumi.String("23_04-gen2"),
					Version:   pulumi.String("latest"),
				},
			},
			VmName: pulumi.String("HelloVM"),
		})
		if err != nil {
			return err
		}
		ctx.Export("vmName", virtualMachine.Name)
		return nil
	})
}
