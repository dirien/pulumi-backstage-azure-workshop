# Chapter 2 - Containerize Backstage

<img src="docs/static/media/chap3.png">

## Overview

In this chapter, we will containerize our Backstage instance. We will use the Pulumi Docker provider to build and push
the container to the Azure Container Registry.

The container will be used in the next chapter to deploy the Backstage instance to Azure as we going to build on top of
what we created in this chapter.

In a production environment, you would use a CI/CD pipeline to build and push the container to the registry. Check for
example this [article](https://learn.microsoft.com/en-us/azure/devops/pipelines/ecosystems/containers/acr-template?view=azure-devops).

## Instructions

### Step 0 - Prerequisites

Overwrite the content of `backstage/app-config.production.yaml` with the following content:

```yaml
app:
  baseUrl: ${BACKSTAGE_BASE_URL}

backend:
  baseUrl: ${BACKSTAGE_BASE_URL}
  listen:
    port: ${WEBSITES_PORT}
    host: 0.0.0.0
  cors:
    origin: ${BACKSTAGE_BASE_URL}
  database:
    client: pg
    connection:
      host: ${POSTGRES_HOST}
      port: ${POSTGRES_PORT}
      user: ${POSTGRES_USER}
      password: ${POSTGRES_PASSWORD}
```

This will be used to configure the Backstage instance to work with Azure, as we will pass the values as environment
variables to the container.

Create a new directory for all the workshop files and navigate into it.

```shell
mkdir backstage-azure-infrastructure
cd backstage-azure-infrastructure
```

This will be the root directory for the workshop. Keep this in mind for all the following steps and chapters.

### Step 1 - Create a new Pulumi project

Similar to the previous chapter, we will create a new Pulumi project. This time, I will use the TypeScript language, but
you
can use any other language you want.

```shell
pulumi new azure-typescript --force
```

> I use the `--force` flag as I want to create the project in the existing directory we created in the previous step.

As we are going to build a container, we need the Pulumi Docker provider. You can add it to your project by running the
following command:

```shell
npm install @pulumi/docker --save
```

And as we work with the Microsoft Entra ID, we also need the Azure Active Directory provider:

```shell
npm install @pulumi/azuread --save
```

> If you are using a different language, you can find the installation instructions for the Docker provider and the
> Azure Active Directory provider in the [Pulumi registry](https://www.pulumi.com/registry/).

You can delete the content of the `index.ts` file and replace it with the following content:

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";
import * as azuread from "@pulumi/azuread";
import * as docker from "@pulumi/docker";


const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
```

<details>
  <summary>YAML</summary>

{% highlight yaml %}
name: backstage-azure-infrastructure-yaml
runtime: yaml
description: A minimal Azure Native Pulumi YAML program

    resources:
    resourceGroup:
      type: azure-native:resources:ResourceGroup

{% endhighlight %}

</details>

<details>
  <summary>Go</summary>

{% highlight go %}
package main

    import (
        "fmt"
    
        "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
        "github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
        "github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
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

{% endhighlight %}

</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    import pulumi
    import pulumi_azure_native as azure_native
    import pulumi_azuread as azuread
    import pulumi_docker as docker
    
    resource_group = azure_native.resources.ResourceGroup("resourceGroup")
  {% endhighlight %}

</details>

<br/>
As the next step, we will create an Azure Container Registry. This will be the place where we will push our container
to.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";
import * as azuread from "@pulumi/azuread";
import * as docker from "@pulumi/docker";

const roleDefinitionIds = [
    "b24988ac-6180-42a0-ab88-20f7382dd24c",
    "acdd72a7-3385-48ef-bd42-f606fba81ae7",
    "7f951dda-4ed3-4680-a7ca-43fe172d538d",
];
const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
const backstageContainerRegistry = new azure_native.containerregistry.Registry("backstageContainerRegistry", {
    resourceGroupName: resourceGroup.name,
    registryName: "pulumibackstage",
    sku: {
        name: "Basic",
    },
    identity: {
        type: azure_native.containerregistry.ResourceIdentityType.SystemAssigned,
    },
    adminUserEnabled: false,
});
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
      backstageContainerRegistry:
        type: azure-native:containerregistry:Registry
        properties:
          resourceGroupName: ${resourceGroup.name}
          registryName: pulumibackstage
          sku:
            name: Basic
          identity:
            type: SystemAssigned
          adminUserEnabled: false
  {% endhighlight %}

</details>

<details>
  <summary>Go</summary>

  {% highlight go %}
    package main
    
    import (
        "fmt"
    
        "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
        "github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
        "github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    )
    
    func main() {
        pulumi.Run(func(ctx *pulumi.Context) error {
            resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
            if err != nil {
                return err
            }
            backstageContainerRegistry, err := containerregistry.NewRegistry(ctx, "backstageContainerRegistry", &containerregistry.RegistryArgs{
                ResourceGroupName: resourceGroup.Name,
                RegistryName:      pulumi.String("pulumibackstage"),
                Sku: &containerregistry.SkuArgs{
                    Name: pulumi.String("Basic"),
                },
                Identity: &containerregistry.IdentityPropertiesArgs{
                    Type: containerregistry.ResourceIdentityTypeSystemAssigned,
                },
                AdminUserEnabled: pulumi.Bool(false),
            })
            if err != nil {
                return err
            }
            
            return nil
        })
    }
  {% endhighlight %}

</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    import pulumi
    import pulumi_azure_native as azure_native
    import pulumi_azuread as azuread
    import pulumi_docker as docker
    
    role_definition_ids = [
        "b24988ac-6180-42a0-ab88-20f7382dd24c",
        "acdd72a7-3385-48ef-bd42-f606fba81ae7",
        "7f951dda-4ed3-4680-a7ca-43fe172d538d",
    ]
    resource_group = azure_native.resources.ResourceGroup("resourceGroup")
    backstage_container_registry = azure_native.containerregistry.Registry("backstageContainerRegistry",
        resource_group_name=resource_group.name,
        registry_name="pulumibackstage",
        sku=azure_native.containerregistry.SkuArgs(
            name="Basic",
        ),
        identity=azure_native.containerregistry.IdentityPropertiesArgs(
            type=azure_native.containerregistry.ResourceIdentityType.SYSTEM_ASSIGNED,
        ),
        admin_user_enabled=False)
  {% endhighlight %}

</details>

<br/>
Now we need to create a service principal for the Azure Container Registry. This will be used to authenticate with the
registry. The service principal will also have some roles assigned to it.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";
import * as azuread from "@pulumi/azuread";
import * as docker from "@pulumi/docker";

const roleDefinitionIds = [
    "b24988ac-6180-42a0-ab88-20f7382dd24c",
    "acdd72a7-3385-48ef-bd42-f606fba81ae7",
    "7f951dda-4ed3-4680-a7ca-43fe172d538d",
];
const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
const backstageContainerRegistry = new azure_native.containerregistry.Registry("backstageContainerRegistry", {
    resourceGroupName: resourceGroup.name,
    registryName: "pulumibackstage",
    sku: {
        name: "Basic",
    },
    identity: {
        type: azure_native.containerregistry.ResourceIdentityType.SystemAssigned,
    },
    adminUserEnabled: false,
});
const contributorRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[0],
    scope: backstageContainerRegistry.id,
});
const readerRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[1],
    scope: backstageContainerRegistry.id,
});
const acrPullRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[2],
    scope: backstageContainerRegistry.id,
});
const backstageAzureApplication = new azuread.Application("backstageAzureApplication", {displayName: "backstage"});
const backstageAzureServicePrincipal = new azuread.ServicePrincipal("backstageAzureServicePrincipal", {
    applicationId: backstageAzureApplication.applicationId,
    tags: ["backstage"],
});
const backstageAzureApplicationPassword = new azuread.ApplicationPassword("backstageAzureApplicationPassword", {
    applicationObjectId: backstageAzureApplication.objectId,
    endDate: "2099-01-01T00:00:00Z",
});
const roleAssignment0 = new azure_native.authorization.RoleAssignment("roleAssignment0", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: contributorRoleDefinition.apply(contributorRoleDefinition => contributorRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
const roleAssignment1 = new azure_native.authorization.RoleAssignment("roleAssignment1", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: readerRoleDefinition.apply(readerRoleDefinition => readerRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
const roleAssignment2 = new azure_native.authorization.RoleAssignment("roleAssignment2", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: acrPullRoleDefinition.apply(acrPullRoleDefinition => acrPullRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
    variables:
      roleDefinitionIds:
      - b24988ac-6180-42a0-ab88-20f7382dd24c
      - acdd72a7-3385-48ef-bd42-f606fba81ae7
      - 7f951dda-4ed3-4680-a7ca-43fe172d538d
    
      contributorRoleDefinition:
        fn::invoke:
          function: azure-native:authorization:getRoleDefinition
          arguments:
            roleDefinitionId: ${roleDefinitionIds[0]}
            scope: ${backstageContainerRegistry.id}
    
      readerRoleDefinition:
        fn::invoke:
          function: azure-native:authorization:getRoleDefinition
          arguments:
            roleDefinitionId: ${roleDefinitionIds[1]}
            scope: ${backstageContainerRegistry.id}
    
      acrPullRoleDefinition:
        fn::invoke:
          function: azure-native:authorization:getRoleDefinition
          arguments:
            roleDefinitionId: ${roleDefinitionIds[2]}
            scope: ${backstageContainerRegistry.id}
    
    resources:
      resourceGroup:
        type: azure-native:resources:ResourceGroup
    
      backstageAzureApplication:
        type: azuread:Application
        properties:
          displayName: "backstage"
    
      backstageAzureServicePrincipal:
        type: azuread:ServicePrincipal
        properties:
          applicationId: ${backstageAzureApplication.applicationId}
          tags: ["backstage"]
    
      backstageAzureApplicationPassword:
        type: azuread:ApplicationPassword
        properties:
          applicationObjectId: ${backstageAzureApplication.objectId}
          endDate: "2099-01-01T00:00:00Z"
    
    
      backstageContainerRegistry:
        type: azure-native:containerregistry:Registry
        properties:
          resourceGroupName: ${resourceGroup.name}
          registryName: pulumibackstage
          sku:
            name: Basic
          identity:
            type: SystemAssigned
          adminUserEnabled: false
    
      roleAssignment0:
        type: azure-native:authorization:RoleAssignment
        properties:
          principalId: ${backstageAzureServicePrincipal.objectId}
          roleDefinitionId: ${contributorRoleDefinition.id}
          scope: ${backstageContainerRegistry.id}
          principalType: ServicePrincipal
    
      roleAssignment1:
        type: azure-native:authorization:RoleAssignment
        properties:
          principalId: ${backstageAzureServicePrincipal.objectId}
          roleDefinitionId: ${readerRoleDefinition.id}
          scope: ${backstageContainerRegistry.id}
          principalType: ServicePrincipal
    
      roleAssignment2:
        type: azure-native:authorization:RoleAssignment
        properties:
          principalId: ${backstageAzureServicePrincipal.objectId}
          roleDefinitionId: ${acrPullRoleDefinition.id}
          scope: ${backstageContainerRegistry.id}
          principalType: ServicePrincipal
  {% endhighlight %}

</details>

<details>
  <summary>Go</summary>

  {% highlight go %}
    package main

    import (
        "fmt"
    
        "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
        "github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
        "github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    )
    
    func main() {
        pulumi.Run(func(ctx *pulumi.Context) error {
            roleDefinitionIds := []string{
                "b24988ac-6180-42a0-ab88-20f7382dd24c",
                "acdd72a7-3385-48ef-bd42-f606fba81ae7",
                "7f951dda-4ed3-4680-a7ca-43fe172d538d",
            }
            resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
            if err != nil {
                return err
            }
            backstageContainerRegistry, err := containerregistry.NewRegistry(ctx, "backstageContainerRegistry", &containerregistry.RegistryArgs{
                ResourceGroupName: resourceGroup.Name,
                RegistryName:      pulumi.String("pulumibackstage"),
                Sku: &containerregistry.SkuArgs{
                    Name: pulumi.String("Basic"),
                },
                Identity: &containerregistry.IdentityPropertiesArgs{
                    Type: containerregistry.ResourceIdentityTypeSystemAssigned,
                },
                AdminUserEnabled: pulumi.Bool(false),
            })
            if err != nil {
                return err
            }
            contributorRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[0]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            readerRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[1]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            acrPullRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[2]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            backstageAzureApplication, err := azuread.NewApplication(ctx, "backstageAzureApplication", &azuread.ApplicationArgs{
                DisplayName: pulumi.String("backstage"),
            })
            if err != nil {
                return err
            }
            backstageAzureServicePrincipal, err := azuread.NewServicePrincipal(ctx, "backstageAzureServicePrincipal", &azuread.ServicePrincipalArgs{
                ApplicationId: backstageAzureApplication.ApplicationId,
                Tags: pulumi.StringArray{
                    pulumi.String("backstage"),
                },
            })
            if err != nil {
                return err
            }
            backstageAzureApplicationPassword, err := azuread.NewApplicationPassword(ctx, "backstageAzureApplicationPassword", &azuread.ApplicationPasswordArgs{
                ApplicationObjectId: backstageAzureApplication.ObjectId,
                EndDate:             pulumi.String("2099-01-01T00:00:00Z"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment0", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: contributorRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment1", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: readerRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment2", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: acrPullRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            return nil
        })
    }
  {% endhighlight %}

</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    import pulumi
    import pulumi_azure_native as azure_native
    import pulumi_azuread as azuread
    import pulumi_docker as docker
    
    role_definition_ids = [
        "b24988ac-6180-42a0-ab88-20f7382dd24c",
        "acdd72a7-3385-48ef-bd42-f606fba81ae7",
        "7f951dda-4ed3-4680-a7ca-43fe172d538d",
    ]
    resource_group = azure_native.resources.ResourceGroup("resourceGroup")
    backstage_container_registry = azure_native.containerregistry.Registry("backstageContainerRegistry",
                                                                           resource_group_name=resource_group.name,
                                                                           registry_name="pulumibackstage",
                                                                           sku=azure_native.containerregistry.SkuArgs(
                                                                               name="Basic",
                                                                           ),
                                                                           identity=azure_native.containerregistry.IdentityPropertiesArgs(
                                                                               type=azure_native.containerregistry.ResourceIdentityType.SYSTEM_ASSIGNED,
                                                                           ),
                                                                           admin_user_enabled=False)
    contributor_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[0],
                                                                                        scope=backstage_container_registry.id)
    reader_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[1],
                                                                                   scope=backstage_container_registry.id)
    acr_pull_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[2],
                                                                                     scope=backstage_container_registry.id)
    backstage_azure_application = azuread.Application("backstageAzureApplication", display_name="backstage")
    backstage_azure_service_principal = azuread.ServicePrincipal("backstageAzureServicePrincipal",
                                                                 application_id=backstage_azure_application.application_id,
                                                                 tags=["backstage"])
    backstage_azure_application_password = azuread.ApplicationPassword("backstageAzureApplicationPassword",
                                                                       application_object_id=backstage_azure_application.object_id,
                                                                       end_date="2099-01-01T00:00:00Z")
    role_assignment0 = azure_native.authorization.RoleAssignment("roleAssignment0",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=contributor_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
    role_assignment1 = azure_native.authorization.RoleAssignment("roleAssignment1",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=reader_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
    role_assignment2 = azure_native.authorization.RoleAssignment("roleAssignment2",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=acr_pull_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
  {% endhighlight %}

</details>

<br/>
The next step is to build the container and push it to the Azure Container Registry. We will use the `docker.Image`
resource from the Pulumi Docker provider to build the container and push it to the registry.

We also going to output the `repoDigest` from the `docker.Image` resource. We will need this in the next chapter to
deploy the Backstage instance to Azure.

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";
import * as azuread from "@pulumi/azuread";
import * as docker from "@pulumi/docker";

const roleDefinitionIds = [
    "b24988ac-6180-42a0-ab88-20f7382dd24c",
    "acdd72a7-3385-48ef-bd42-f606fba81ae7",
    "7f951dda-4ed3-4680-a7ca-43fe172d538d",
];
const resourceGroup = new azure_native.resources.ResourceGroup("resourceGroup", {});
const backstageContainerRegistry = new azure_native.containerregistry.Registry("backstageContainerRegistry", {
    resourceGroupName: resourceGroup.name,
    registryName: "pulumibackstage",
    sku: {
        name: "Basic",
    },
    identity: {
        type: azure_native.containerregistry.ResourceIdentityType.SystemAssigned,
    },
    adminUserEnabled: false,
});
const contributorRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[0],
    scope: backstageContainerRegistry.id,
});
const readerRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[1],
    scope: backstageContainerRegistry.id,
});
const acrPullRoleDefinition = azure_native.authorization.getRoleDefinitionOutput({
    roleDefinitionId: roleDefinitionIds[2],
    scope: backstageContainerRegistry.id,
});
const backstageAzureApplication = new azuread.Application("backstageAzureApplication", {displayName: "backstage"});
const backstageAzureServicePrincipal = new azuread.ServicePrincipal("backstageAzureServicePrincipal", {
    applicationId: backstageAzureApplication.applicationId,
    tags: ["backstage"],
});
const backstageAzureApplicationPassword = new azuread.ApplicationPassword("backstageAzureApplicationPassword", {
    applicationObjectId: backstageAzureApplication.objectId,
    endDate: "2099-01-01T00:00:00Z",
});
const roleAssignment0 = new azure_native.authorization.RoleAssignment("roleAssignment0", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: contributorRoleDefinition.apply(contributorRoleDefinition => contributorRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
const roleAssignment1 = new azure_native.authorization.RoleAssignment("roleAssignment1", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: readerRoleDefinition.apply(readerRoleDefinition => readerRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
const roleAssignment2 = new azure_native.authorization.RoleAssignment("roleAssignment2", {
    principalId: backstageAzureServicePrincipal.objectId,
    roleDefinitionId: acrPullRoleDefinition.apply(acrPullRoleDefinition => acrPullRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});
const backstageImage = new docker.Image("backstageImage", {
    build: {
        context: "../backstage",
        platform: "linux/amd64",
        builderVersion: docker.BuilderVersion.BuilderBuildKit,
        dockerfile: "../backstage/packages/backend/Dockerfile",
    },
    imageName: pulumi.interpolate`${backstageContainerRegistry.loginServer}/backstage`,
    registry: {
        server: backstageContainerRegistry.loginServer,
        username: backstageAzureServicePrincipal.applicationId,
        password: backstageAzureApplicationPassword.value,
    },
});
export const repoDigest = backstageImage.repoDigest;
```

<details>
  <summary>YAML</summary>

  {% highlight yaml %}
          backstageImage:
          type: docker:Image
          properties:
            build:
              context: "../backstage"
              platform: "linux/amd64"
              builderVersion: BuilderBuildKit
              dockerfile: "../backstage/packages/backend/Dockerfile"
            imageName: ${backstageContainerRegistry.loginServer}/backstage
            registry:
              server: ${backstageContainerRegistry.loginServer}
              username: ${backstageAzureServicePrincipal.applicationId}
              password: ${backstageAzureApplicationPassword.value}

      outputs:
        repoDigest: ${backstageImage.repoDigest}
  {% endhighlight %}

</details>

<details>
  <summary>Go</summary>

  {% highlight go %}
    package main

    import (
        "fmt"
    
        "github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
        "github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
        "github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
        "github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    )
    
    func main() {
        pulumi.Run(func(ctx *pulumi.Context) error {
            roleDefinitionIds := []string{
                "b24988ac-6180-42a0-ab88-20f7382dd24c",
                "acdd72a7-3385-48ef-bd42-f606fba81ae7",
                "7f951dda-4ed3-4680-a7ca-43fe172d538d",
            }
            resourceGroup, err := resources.NewResourceGroup(ctx, "resourceGroup", nil)
            if err != nil {
                return err
            }
            backstageContainerRegistry, err := containerregistry.NewRegistry(ctx, "backstageContainerRegistry", &containerregistry.RegistryArgs{
                ResourceGroupName: resourceGroup.Name,
                RegistryName:      pulumi.String("pulumibackstage"),
                Sku: &containerregistry.SkuArgs{
                    Name: pulumi.String("Basic"),
                },
                Identity: &containerregistry.IdentityPropertiesArgs{
                    Type: containerregistry.ResourceIdentityTypeSystemAssigned,
                },
                AdminUserEnabled: pulumi.Bool(false),
            })
            if err != nil {
                return err
            }
            contributorRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[0]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            readerRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[1]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            acrPullRoleDefinition := authorization.LookupRoleDefinitionOutput(ctx, authorization.LookupRoleDefinitionOutputArgs{
                RoleDefinitionId: pulumi.String(roleDefinitionIds[2]),
                Scope:            backstageContainerRegistry.ID(),
            }, nil)
            backstageAzureApplication, err := azuread.NewApplication(ctx, "backstageAzureApplication", &azuread.ApplicationArgs{
                DisplayName: pulumi.String("backstage"),
            })
            if err != nil {
                return err
            }
            backstageAzureServicePrincipal, err := azuread.NewServicePrincipal(ctx, "backstageAzureServicePrincipal", &azuread.ServicePrincipalArgs{
                ApplicationId: backstageAzureApplication.ApplicationId,
                Tags: pulumi.StringArray{
                    pulumi.String("backstage"),
                },
            })
            if err != nil {
                return err
            }
            backstageAzureApplicationPassword, err := azuread.NewApplicationPassword(ctx, "backstageAzureApplicationPassword", &azuread.ApplicationPasswordArgs{
                ApplicationObjectId: backstageAzureApplication.ObjectId,
                EndDate:             pulumi.String("2099-01-01T00:00:00Z"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment0", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: contributorRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment1", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: readerRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            _, err = authorization.NewRoleAssignment(ctx, "roleAssignment2", &authorization.RoleAssignmentArgs{
                PrincipalId:      backstageAzureServicePrincipal.ObjectId,
                RoleDefinitionId: acrPullRoleDefinition.Id(),
                Scope:            backstageContainerRegistry.ID(),
                PrincipalType:    pulumi.String("ServicePrincipal"),
            })
            if err != nil {
                return err
            }
            backstageImage, err := docker.NewImage(ctx, "backstageImage", &docker.ImageArgs{
                Build: &docker.DockerBuildArgs{
                    Context:        pulumi.String("../../backstage"),
                    Platform:       pulumi.String("linux/amd64"),
                    BuilderVersion: docker.BuilderVersionBuilderBuildKit,
                    Dockerfile:     pulumi.String("../../backstage/packages/backend/Dockerfile"),
                },
                ImageName: backstageContainerRegistry.LoginServer.ApplyT(func(loginServer string) (string, error) {
                    return fmt.Sprintf("%v/backstage", loginServer), nil
                }).(pulumi.StringOutput),
                Registry: &docker.RegistryArgs{
                    Server:   backstageContainerRegistry.LoginServer,
                    Username: backstageAzureServicePrincipal.ApplicationId,
                    Password: backstageAzureApplicationPassword.Value,
                },
            })
            if err != nil {
                return err
            }
            ctx.Export("repoDigest", backstageImage.RepoDigest)
            return nil
        })
    }
  {% endhighlight %}

</details>

<details>
  <summary>Python</summary>

  {% highlight python %}
    import pulumi
    import pulumi_azure_native as azure_native
    import pulumi_azuread as azuread
    import pulumi_docker as docker
    
    role_definition_ids = [
        "b24988ac-6180-42a0-ab88-20f7382dd24c",
        "acdd72a7-3385-48ef-bd42-f606fba81ae7",
        "7f951dda-4ed3-4680-a7ca-43fe172d538d",
    ]
    resource_group = azure_native.resources.ResourceGroup("resourceGroup")
    backstage_container_registry = azure_native.containerregistry.Registry("backstageContainerRegistry",
                                                                           resource_group_name=resource_group.name,
                                                                           registry_name="pulumibackstage",
                                                                           sku=azure_native.containerregistry.SkuArgs(
                                                                               name="Basic",
                                                                           ),
                                                                           identity=azure_native.containerregistry.IdentityPropertiesArgs(
                                                                               type=azure_native.containerregistry.ResourceIdentityType.SYSTEM_ASSIGNED,
                                                                           ),
                                                                           admin_user_enabled=False)
    contributor_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[0],
                                                                                        scope=backstage_container_registry.id)
    reader_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[1],
                                                                                   scope=backstage_container_registry.id)
    acr_pull_role_definition = azure_native.authorization.get_role_definition_output(role_definition_id=role_definition_ids[2],
                                                                                     scope=backstage_container_registry.id)
    backstage_azure_application = azuread.Application("backstageAzureApplication", display_name="backstage")
    backstage_azure_service_principal = azuread.ServicePrincipal("backstageAzureServicePrincipal",
                                                                 application_id=backstage_azure_application.application_id,
                                                                 tags=["backstage"])
    backstage_azure_application_password = azuread.ApplicationPassword("backstageAzureApplicationPassword",
                                                                       application_object_id=backstage_azure_application.object_id,
                                                                       end_date="2099-01-01T00:00:00Z")
    role_assignment0 = azure_native.authorization.RoleAssignment("roleAssignment0",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=contributor_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
    role_assignment1 = azure_native.authorization.RoleAssignment("roleAssignment1",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=reader_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
    role_assignment2 = azure_native.authorization.RoleAssignment("roleAssignment2",
                                                                 principal_id=backstage_azure_service_principal.object_id,
                                                                 role_definition_id=acr_pull_role_definition.id,
                                                                 scope=backstage_container_registry.id,
                                                                 principal_type="ServicePrincipal")
    backstage_image = docker.Image("backstageImage",
                                   build=docker.DockerBuildArgs(
                                       context="../backstage",
                                       platform="linux/amd64",
                                       builder_version=docker.BuilderVersion.BUILDER_BUILD_KIT,
                                       dockerfile="../backstage/packages/backend/Dockerfile",
                                   ),
                                   image_name=backstage_container_registry.login_server.apply(lambda login_server: f"{login_server}/backstage"),
                                   registry=docker.RegistryArgs(
                                       server=backstage_container_registry.login_server,
                                       username=backstage_azure_service_principal.application_id,
                                       password=backstage_azure_application_password.value,
                                   ))
    pulumi.export("repoDigest", backstage_image.repo_digest)
  {% endhighlight %}

</details>

### Step 2 - Run Pulumi Up

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

You should see the following output:

```shell
...
Outputs:
  + repoDigest: "pulumibackstage.azurecr.io/backstage@sha256:38e6095a50f677b87ae95defe00bd0b2deff7c741ea74ec8fcfb46a6d014c56b"
```

This is the `repoDigest` of your container!

## Stretch Goals

- Can you set a sensible `RetentionPolicy` for the Azure Container Registry? (Hint: You have to change the SKU of the
  registry to `Premium` to use the `RetentionPolicy`.)

## Learn more

- [Pulumi](https://www.pulumi.com/)
- [Pulumi Azure Native Provider](https://www.pulumi.com/registry/packages/azure-native/)
- [Pulumi Microsoft Entra ID Provider](https://www.pulumi.com/registry/packages/azuread/)
- [Pulumi Docker Provider](https://www.pulumi.com/registry/packages/docker/)
