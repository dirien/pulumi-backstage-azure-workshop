package main

import (
	"fmt"
	"github.com/pulumi/pulumi-azure-native-sdk/authorization/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/dbforpostgresql/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/insights/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/operationalinsights/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/web/v2"
	"github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		azureDevOpsToken := cfg.Require("azureDevOpsToken")
		azureDevOpsOrganization := "dirien"
		if param := cfg.Get("azureDevOpsOrganization"); param != "" {
			azureDevOpsOrganization = param
		}
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
		backstageOperationalInsightsWorkspace, err := operationalinsights.NewWorkspace(ctx, "backstageOperationalInsightsWorkspace", &operationalinsights.WorkspaceArgs{
			ResourceGroupName: resourceGroup.Name,
			RetentionInDays:   pulumi.Int(30),
			Sku: &operationalinsights.WorkspaceSkuArgs{
				Name: pulumi.String("PerGB2018"),
			},
		})
		if err != nil {
			return err
		}
		backstageAppInsights, err := insights.NewComponent(ctx, "backstageAppInsights", &insights.ComponentArgs{
			ResourceGroupName:   resourceGroup.Name,
			ApplicationType:     pulumi.String("other"),
			Kind:                pulumi.String("other"),
			WorkspaceResourceId: backstageOperationalInsightsWorkspace.ID(),
		})
		if err != nil {
			return err
		}
		backstagePostgresqlServer, err := dbforpostgresql.NewServer(ctx, "backstagePostgresqlServer", &dbforpostgresql.ServerArgs{
			ResourceGroupName: resourceGroup.Name,
			Sku: &dbforpostgresql.SkuArgs{
				Name: pulumi.String("Standard_D2ds_v4"),
				Tier: pulumi.String("GeneralPurpose"),
			},
			Storage: &dbforpostgresql.StorageArgs{
				StorageSizeGB: pulumi.Int(32),
			},
			Backup: &dbforpostgresql.BackupArgs{
				GeoRedundantBackup: pulumi.String("Disabled"),
			},
			Version:                    pulumi.String("11"),
			AdministratorLogin:         pulumi.String("backstage"),
			AdministratorLoginPassword: pulumi.String("1Backstage1!"),
			ServerName:                 pulumi.String(fmt.Sprintf("backstage-postgresql-%v", ctx.Project())),
		})
		if err != nil {
			return err
		}
		_, err = dbforpostgresql.NewFirewallRule(ctx, "backstagePostgresqlFirewallRule", &dbforpostgresql.FirewallRuleArgs{
			ResourceGroupName: resourceGroup.Name,
			ServerName:        backstagePostgresqlServer.Name,
			StartIpAddress:    pulumi.String("0.0.0.0"),
			EndIpAddress:      pulumi.String("255.255.255.255"),
		})
		if err != nil {
			return err
		}
		_, err = dbforpostgresql.NewDatabase(ctx, "backstagePostgresqlDatabase", &dbforpostgresql.DatabaseArgs{
			ResourceGroupName: resourceGroup.Name,
			ServerName:        backstagePostgresqlServer.Name,
			Collation:         pulumi.String("en_US.utf8"),
			Charset:           pulumi.String("UTF8"),
		})
		if err != nil {
			return err
		}
		backstageAppServicePlan, err := web.NewAppServicePlan(ctx, "backstageAppServicePlan", &web.AppServicePlanArgs{
			ResourceGroupName: resourceGroup.Name,
			Kind:              pulumi.String("Linux"),
			Sku: &web.SkuDescriptionArgs{
				Name: pulumi.String("S1"),
				Tier: pulumi.String("Standard"),
			},
			Reserved: pulumi.Bool(true),
			IsSpot:   pulumi.Bool(false),
		})
		if err != nil {
			return err
		}
		backstageWebApp, err := web.NewWebApp(ctx, "backstageWebApp", &web.WebAppArgs{
			Name:              pulumi.String("my-backstage-app"),
			ResourceGroupName: resourceGroup.Name,
			ServerFarmId:      backstageAppServicePlan.ID(),
			Kind:              pulumi.String("app,linux,container"),
			Identity: &web.ManagedServiceIdentityArgs{
				Type: web.ManagedServiceIdentityTypeSystemAssigned,
			},
			SiteConfig: &web.SiteConfigArgs{
				Cors: &web.CorsSettingsArgs{
					SupportCredentials: pulumi.Bool(true),
					AllowedOrigins: pulumi.StringArray{
						pulumi.String("https://my-backstage-app.azurewebsites.net"),
					},
				},
				HttpLoggingEnabled: pulumi.Bool(true),
				AppSettings: web.NameValuePairArray{
					&web.NameValuePairArgs{
						Name:  pulumi.String("POSTGRES_HOST"),
						Value: backstagePostgresqlServer.FullyQualifiedDomainName,
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("POSTGRES_PORT"),
						Value: pulumi.String("5432"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("POSTGRES_USER"),
						Value: backstagePostgresqlServer.AdministratorLogin,
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("POSTGRES_PASSWORD"),
						Value: pulumi.String("1Backstage1!"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("AZURE_PAT"),
						Value: pulumi.String(azureDevOpsToken),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("AZURE_ORG"),
						Value: pulumi.String(azureDevOpsOrganization),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("APPINSIGHTS_INSTRUMENTATIONKEY"),
						Value: backstageAppInsights.InstrumentationKey,
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("DOCKER_ENABLE_CI"),
						Value: pulumi.String("true"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("WEBSITES_PORT"),
						Value: pulumi.String("7007"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("PORT"),
						Value: pulumi.String("8080"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("BACKSTAGE_BASE_URL"),
						Value: pulumi.String("https://my-backstage-app.azurewebsites.net"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("APP_CONFIG_backend_database_connection_ssl_required"),
						Value: pulumi.String("true"),
					},
					&web.NameValuePairArgs{
						Name:  pulumi.String("APP_CONFIG_backend_database_connection_ssl_rejectUnauthorized"),
						Value: pulumi.String("true"),
					},
				},
				AcrUseManagedIdentityCreds: pulumi.Bool(true),
				LinuxFxVersion: backstageImage.RepoDigest.ApplyT(func(repoDigest string) (string, error) {
					return fmt.Sprintf("DOCKER|%v", repoDigest), nil
				}).(pulumi.StringOutput),
			},
		})
		if err != nil {
			return err
		}
		_, err = authorization.NewRoleAssignment(ctx, "roleAssignment3", &authorization.RoleAssignmentArgs{
			PrincipalId:      backstageWebApp.Identity.PrincipalId().Elem(),
			RoleDefinitionId: acrPullRoleDefinition.Id(),
			Scope:            backstageContainerRegistry.ID(),
			PrincipalType:    pulumi.String("ServicePrincipal"),
		})
		if err != nil {
			return err
		}
		ctx.Export("backstageWebAppUrl", pulumi.Sprintf("https://%v", backstageWebApp.DefaultHostName))
		ctx.Export("repoDigest", backstageImage.RepoDigest)
		return nil
	})
}
