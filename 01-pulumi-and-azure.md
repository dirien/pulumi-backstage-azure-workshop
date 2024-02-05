# Chapter 1 - Pulumi & Azure!

<img src="docs/static/media/chap2.png">

## Introduction

In this chapter, we will develop a simple Pulumi program that creates a single virtual machine with a basic web server
running on it. Our goal is to become acquainted with the Pulumi CLI, understand the structure of a Pulumi program, and
learn how to create multiple stacks and override default values.

### Modern Infrastructure As Code with Pulumi

Pulumi is an open-source infrastructure-as-code tool for creating, deploying and managing cloud
infrastructure. Pulumi works with traditional infrastructures like VMs, networks, and databases and modern
architectures, including containers, Kubernetes clusters, and serverless functions. Pulumi supports dozens of public,
private, and hybrid cloud service providers.

Pulumi is a multi-language infrastructure as Code tool using imperative languages to create a declarative
infrastructure description.

You have a wide range of programming languages available, and you can use the one you and your team are the most
comfortable with. Currently, (6/2023) Pulumi supports the following languages:

* Node.js (JavaScript / TypeScript)

* Python

* Go

* Java

* .NET (C#, VB, F#)

* YAML

## Instructions

### Step 1 - Configure the Azure CLI

The CLI instructions assume you’re using the Azure CLI (az).

Log in to the Azure CLI and Pulumi will automatically use your credentials:

```shell
az login
A web browser has been opened at https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize. Please continue the login in the web browser. If no web browser is available or if the web browser fails to open, use device code flow with `az login --use-device-code`.
```

Do as instructed to log in. After completed, az login will return and you are ready to go.

```shell
az account list
```

Pick out the <id> from the list and run:

```shell
az account set --subscription=<id>
```

### Step 2 - Configure the Pulumi CLI

> If you run Pulumi for the first time, you will be asked to log in. Follow the instructions on the screen to
> login. You may need to create an account first, don't worry it is free.

To initialize a new Pulumi project, run `pulumi new` and select from all the available templates the `azure-<language>`.
The language is the programming language you want to use. The example below uses Go.

```shell
pulumi new azure-go --dir hello-azure-go
```

You will be guided through a wizard to create a new Pulumi project. You can use the following values:

```shell
project name (hello-azure-go):  
project description (A minimal Azure Native Go Pulumi program):  
Created project 'hello-azure-go'

Please enter your desired stack name.
To create a stack in an organization, use the format <org-name>/<stack-name> (e.g. `acmecorp/dev`).
stack name (dev):  
Created stack 'dev'

azure-native:location: The Azure location to use (WestUS2): WestEurope 
```

The template `azure-go` will create a new Pulumi project with
the [Pulumi Azure Native provider](https://www.pulumi.com/registry/packages/azure-native/) already installed. For
detailed instructions, refer to the Pulumi Azure Native Provider documentation.

Remove all code from the `main.go` file and replace it with the following code, we will add more resources later on.


### Step 3 - Add a Virtual Machine to the Pulumi program

Now, let's begin adding resources to our Pulumi program, starting with a basic virtual machine.

For a comprehensive list of available options, consult
the [Pulumi Azure Native provider](https://www.pulumi.com/registry/packages/azure-native/) documentation or
utilize your Intellisense for code completion.

To gather more information about the available images and instance types, execute the following 'az' commands.

```shell
az vm list-sizes --location "westeurope" --output table
az vm image list -p canonical -f 0001-com-ubuntu-server-lunar --all  -o table --location "westeurope" -s 23_04-gen2 
```

Please use and `0001-com-ubuntu-server-lunar/Canonical/23_04-gen2` image and a `Standard_B2s` instance type.

We start with the resource group. The resource group is a logical container for resources deployed on Azure.

```go
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
		return nil
	})
}
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
    name: hello-azure-yaml
    runtime: yaml
    description: A minimal Azure Native Pulumi YAML program

    resources:
      resourceGroup:
        type: azure-native:resources:ResourceGroup
  {% endhighlight %}
</details>

<details>
  <summary>TypeScript</summary>

  {% highlight typescript %}
    import * as pulumi from "@pulumi/pulumi";
    import * as azure_native from "@pulumi/azure-native";
    
    const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
  {% endhighlight %}
</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    import pulumi
    import pulumi_azure_native as azure_native
    
    resource_group = azure_native.resources.ResourceGroup("resourceGroup")
  {% endhighlight %}
</details>

<br/>
Next, we add all the networking resources to the program. We need a virtual network, a subnet, and a network interface.

```go
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
		return nil
	})
}
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
        virtualNetwork:
        type: azure-native:network:VirtualNetwork
        properties:
          resourceGroupName: ${resourceGroup.name}
          addressSpace:
            addressPrefixes:
            - "10.0.0.0/16"
        subnet:
        type: azure-native:network:Subnet
        properties:
          resourceGroupName: ${resourceGroup.name}
          virtualNetworkName: ${virtualNetwork.name}
          addressPrefix: "10.0.1.0/24"
        
        # Create an Azure Network Interface
        networkInterface:
        type: azure-native:network:NetworkInterface
        properties:
          resourceGroupName: ${resourceGroup.name}
          ipConfigurations:
          - name: test-ip-config
            primary: true
            subnet:
              id: ${subnet.id}
  {% endhighlight %}
</details>

<details>
  <summary>TypeScript</summary>

  {% highlight typescript %}
    const virtualNetwork = new azure_native.network.VirtualNetwork("virtualNetwork", {
    resourceGroupName: resourceGroup.name,
    addressSpace: {
        addressPrefixes: ["10.0.0.0/16"],
    },
    });
    const subnet = new azure_native.network.Subnet("subnet", {
        resourceGroupName: resourceGroup.name,
        virtualNetworkName: virtualNetwork.name,
        addressPrefix: "10.0.1.0/24",
    });
    const networkInterface = new azure_native.network.NetworkInterface("networkInterface", {
        resourceGroupName: resourceGroup.name,
        ipConfigurations: [{
            name: "test-ip-config",
            primary: true,
            subnet: {
                id: subnet.id,
            },
        }],
    });
  {% endhighlight %}
</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    virtual_network = azure_native.network.VirtualNetwork("virtualNetwork",
      resource_group_name=resource_group.name,
      address_space=azure_native.network.AddressSpaceArgs(
          address_prefixes=["10.0.0.0/16"],
      ))
    subnet = azure_native.network.Subnet("subnet",
     resource_group_name=resource_group.name,
     virtual_network_name=virtual_network.name,
     address_prefix="10.0.1.0/24")
    network_interface = azure_native.network.NetworkInterface("networkInterface",
      resource_group_name=resource_group.name,
      ip_configurations=[azure_native.network.NetworkInterfaceIPConfigurationArgs(
          name="test-ip-config",
          primary=True,
          subnet=azure_native.network.SubnetArgs(
              id=subnet.id,
          ),
      )])
  {% endhighlight %}
</details>

<br/>
Finally, we add the virtual machine to the program.

```go
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
		return nil
	})
}
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
      virtualMachine:
        type: azure-native:compute:VirtualMachine
        properties:
          networkProfile:
            networkInterfaces:
            - id: ${networkInterface.id}
          hardwareProfile:
            vmSize: Standard_B2s
          osProfile:
            computerName: HelloVM
            adminUsername: azureuser
            adminPassword: "Password1234!"
          resourceGroupName: ${resourceGroup.name}
          storageProfile:
            imageReference:
              offer: 0001-com-ubuntu-server-lunar
              publisher: Canonical
              sku: 23_04-gen2
              version: latest
          vmName: HelloVM
  {% endhighlight %}
</details>

<details>
  <summary>TypeScript</summary>

  {% highlight typescript %}
    const virtualMachine = new azure_native.compute.VirtualMachine("virtualMachine", {
        networkProfile: {
            networkInterfaces: [{
                id: networkInterface.id,
            }],
        },
        hardwareProfile: {
            vmSize: "Standard_B2s",
        },
        osProfile: {
            computerName: "HelloVM",
            adminUsername: "azureuser",
            adminPassword: "Password1234!",
        },
        resourceGroupName: resourceGroup.name,
        storageProfile: {
            imageReference: {
                offer: "0001-com-ubuntu-server-lunar",
                publisher: "Canonical",
                sku: "23_04-gen2",
                version: "latest",
            },
        },
        vmName: "HelloVM",
    });
  {% endhighlight %}
</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    virtual_machine = azure_native.compute.VirtualMachine("virtualMachine",
      network_profile=azure_native.compute.NetworkProfileArgs(
          network_interfaces=[azure_native.compute.NetworkInterfaceReferenceArgs(
              id=network_interface.id,
          )],
      ),
      hardware_profile=azure_native.compute.HardwareProfileArgs(
          vm_size="Standard_B2s",
      ),
      os_profile=azure_native.compute.OSProfileArgs(
          computer_name="HelloVM",
          admin_username="azureuser",
          admin_password="Password1234!",
      ),
      resource_group_name=resource_group.name,
      storage_profile=azure_native.compute.StorageProfileArgs(
          image_reference=azure_native.compute.ImageReferenceArgs(
              offer="0001-com-ubuntu-server-lunar",
              publisher="Canonical",
              sku="23_04-gen2",
              version="latest",
          ),
      ),
      vm_name="HelloVM")    
    
  {% endhighlight %}
</details>

<br/>
At the end we output the name of the virtual machine.

```go
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
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
    outputs:
      vmName: ${virtualMachine.name}
  {% endhighlight %}
</details>

<details>
  <summary>TypeScript</summary>

  {% highlight typescript %}
    export const vmName = virtualMachine.name;
  {% endhighlight %}
</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    pulumi.export("vmName", virtual_machine.name)
  {% endhighlight %}
</details>

### Step 4 - Run Pulumi Up

> Before you can run `pulumi up`, you need to be sure that your Azure credentials are set.

```shell
pulumi up
```

This command will show you a preview of all the resources and asks you if you want to deploy them. You can run dedicated
commands to see the preview or to deploy the resources.

```shell
pulumi preview
# or
pulumi up
```

### Step 5 - Destroy the stack

To destroy the stack, run the following command.

```shell
pulumi destroy
pulumi stack rm <stack-name>
```

And confirm the destruction with `yes`.

To switch between stacks, you can use the following command.

```shell
pulumi stack select <stack-name>
```

## Stretch Goals

- Can you create a new stack with a different virtual machine size?
- Can you enable public IP access to the virtual machine?
- Can you enable ssh access to the virtual machine with your local ssh key?

## Learn more

- [Pulumi](https://www.pulumi.com/)
- [Pulumi Azure Native Provider](https://www.pulumi.com/registry/packages/azure-native/)
- [Pulumi ESC](https://www.pulumi.com/docs/pulumi-cloud/esc/)
