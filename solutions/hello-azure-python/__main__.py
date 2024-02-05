import pulumi
import pulumi_azure_native as azure_native

resource_group = azure_native.resources.ResourceGroup("resourceGroup")
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
pulumi.export("vmName", virtual_machine.name)
