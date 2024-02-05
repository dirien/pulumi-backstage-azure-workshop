import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";

const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
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
export const vmName = virtualMachine.name;
