import * as pulumi from "@pulumi/pulumi";
import * as azure_native from "@pulumi/azure-native";
import * as azuread from "@pulumi/azuread";
import * as docker from "@pulumi/docker";


const config = new pulumi.Config();
const azureDevOpsToken = config.require("azureDevOpsToken");
const azureDevOpsOrganization = config.get("azureDevOpsOrganization") || "dirien";
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
        context: "../../backstage",
        platform: "linux/amd64",
        builderVersion: docker.BuilderVersion.BuilderBuildKit,
        dockerfile: "../../backstage/packages/backend/Dockerfile",
    },
    imageName: pulumi.interpolate`${backstageContainerRegistry.loginServer}/backstage`,
    registry: {
        server: backstageContainerRegistry.loginServer,
        username: backstageAzureServicePrincipal.applicationId,
        password: backstageAzureApplicationPassword.value,
    },
});
const backstageOperationalInsightsWorkspace = new azure_native.operationalinsights.Workspace("backstageOperationalInsightsWorkspace", {
    resourceGroupName: resourceGroup.name,
    retentionInDays: 30,
    sku: {
        name: "PerGB2018",
    },
});
const backstageAppInsights = new azure_native.insights.Component("backstageAppInsights", {
    resourceGroupName: resourceGroup.name,
    applicationType: "other",
    kind: "other",
    workspaceResourceId: backstageOperationalInsightsWorkspace.id,
});
const backstagePostgresqlServer = new azure_native.dbforpostgresql.v20230301preview.Server("backstagePostgresqlServer", {
    resourceGroupName: resourceGroup.name,
    sku: {
        name: "Standard_D2ds_v4",
        tier: "GeneralPurpose",
    },
    storage: {
        storageSizeGB: 32,
    },
    backup: {
        geoRedundantBackup: "Disabled",
    },
    version: "11",
    administratorLogin: "backstage",
    administratorLoginPassword: "1Backstage1!",
    serverName: `backstage-postgresql-${pulumi.getProject()}`,
});
const backstagePostgresqlFirewallRule = new azure_native.dbforpostgresql.FirewallRule("backstagePostgresqlFirewallRule", {
    resourceGroupName: resourceGroup.name,
    serverName: backstagePostgresqlServer.name,
    startIpAddress: "0.0.0.0",
    endIpAddress: "255.255.255.255",
});
const backstagePostgresqlDatabase = new azure_native.dbforpostgresql.Database("backstagePostgresqlDatabase", {
    resourceGroupName: resourceGroup.name,
    serverName: backstagePostgresqlServer.name,
    collation: "en_US.utf8",
    charset: "UTF8",
});
const backstageAppServicePlan = new azure_native.web.AppServicePlan("backstageAppServicePlan", {
    resourceGroupName: resourceGroup.name,
    kind: "Linux",
    sku: {
        name: "S1",
        tier: "Standard",
    },
    reserved: true,
    isSpot: false,
});
const backstageWebApp = new azure_native.web.WebApp("backstageWebApp", {
    name: "my-backstage-app",
    resourceGroupName: resourceGroup.name,
    serverFarmId: backstageAppServicePlan.id,
    kind: "app,linux,container",
    identity: {
        type: azure_native.web.ManagedServiceIdentityType.SystemAssigned,
    },
    siteConfig: {
        cors: {
            supportCredentials: true,
            allowedOrigins: ["https://my-backstage-app.azurewebsites.net"],
        },
        httpLoggingEnabled: true,
        appSettings: [
            {
                name: "POSTGRES_HOST",
                value: backstagePostgresqlServer.fullyQualifiedDomainName,
            },
            {
                name: "POSTGRES_PORT",
                value: "5432",
            },
            {
                name: "POSTGRES_USER",
                value: backstagePostgresqlServer.administratorLogin.apply(administratorLogin => administratorLogin || "xxx"),
            },
            {
                name: "POSTGRES_PASSWORD",
                value: "1Backstage1!",
            },
            {
                name: "AZURE_PAT",
                value: azureDevOpsToken,
            },
            {
                name: "AZURE_ORG",
                value: azureDevOpsOrganization,
            },
            {
                name: "APPINSIGHTS_INSTRUMENTATIONKEY",
                value: backstageAppInsights.instrumentationKey,
            },
            {
                name: "DOCKER_ENABLE_CI",
                value: "true",
            },
            {
                name: "WEBSITES_PORT",
                value: "7007",
            },
            {
                name: "PORT",
                value: "8080",
            },
            {
                name: "BACKSTAGE_BASE_URL",
                value: "https://my-backstage-app.azurewebsites.net",
            },
            {
                name: "APP_CONFIG_backend_database_connection_ssl_required",
                value: "true",
            },
            {
                name: "APP_CONFIG_backend_database_connection_ssl_rejectUnauthorized",
                value: "true",
            },
        ],
        acrUseManagedIdentityCreds: true,
        linuxFxVersion: pulumi.interpolate`DOCKER|${backstageImage.repoDigest}`,
    },
});
const roleAssignment3 = new azure_native.authorization.RoleAssignment("roleAssignment3", {
    principalId: backstageWebApp.identity.apply(identity => identity?.principalId || ""),
    roleDefinitionId: acrPullRoleDefinition.apply(acrPullRoleDefinition => acrPullRoleDefinition.id),
    scope: backstageContainerRegistry.id,
    principalType: "ServicePrincipal",
});

export const backstageWebAppUrl = pulumi.interpolate`https://${backstageWebApp.defaultHostName}`;
export const repoDigest = backstageImage.repoDigest;
