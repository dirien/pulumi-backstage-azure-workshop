import pulumi
import pulumi_azure_native as azure_native
import pulumi_azuread as azuread
import pulumi_docker as docker

config = pulumi.Config()
azure_dev_ops_token = config.require("azureDevOpsToken")
azure_dev_ops_organization = config.get("azureDevOpsOrganization")
if azure_dev_ops_organization is None:
    azure_dev_ops_organization = "dirien"
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
        context="../../backstage",
        platform="linux/amd64",
        builder_version=docker.BuilderVersion.BUILDER_BUILD_KIT,
        dockerfile="../../backstage/packages/backend/Dockerfile",
    ),
    image_name=backstage_container_registry.login_server.apply(lambda login_server: f"{login_server}/backstage"),
    registry=docker.RegistryArgs(
        server=backstage_container_registry.login_server,
        username=backstage_azure_service_principal.application_id,
        password=backstage_azure_application_password.value,
    ))
backstage_operational_insights_workspace = azure_native.operationalinsights.Workspace("backstageOperationalInsightsWorkspace",
    resource_group_name=resource_group.name,
    retention_in_days=30,
    sku=azure_native.operationalinsights.WorkspaceSkuArgs(
        name="PerGB2018",
    ))
backstage_app_insights = azure_native.insights.Component("backstageAppInsights",
    resource_group_name=resource_group.name,
    application_type="other",
    kind="other",
    workspace_resource_id=backstage_operational_insights_workspace.id)
backstage_postgresql_server = azure_native.dbforpostgresql.v20230301preview.Server("backstagePostgresqlServer",
    resource_group_name=resource_group.name,
    sku=azure_native.dbforpostgresql.v20230301preview.SkuArgs(
        name="Standard_D2ds_v4",
        tier="GeneralPurpose",
    ),
    storage=azure_native.dbforpostgresql.v20230301preview.StorageArgs(
        storage_size_gb=32,
    ),
    backup=azure_native.dbforpostgresql.v20230301preview.BackupArgs(
        geo_redundant_backup="Disabled",
    ),
    version="11",
    administrator_login="backstage",
    administrator_login_password="1Backstage1!",
    server_name=f"backstage-postgresql-{pulumi.get_project()}")
backstage_postgresql_firewall_rule = azure_native.dbforpostgresql.FirewallRule("backstagePostgresqlFirewallRule",
    resource_group_name=resource_group.name,
    server_name=backstage_postgresql_server.name,
    start_ip_address="0.0.0.0",
    end_ip_address="255.255.255.255")
backstage_postgresql_database = azure_native.dbforpostgresql.Database("backstagePostgresqlDatabase",
    resource_group_name=resource_group.name,
    server_name=backstage_postgresql_server.name,
    collation="en_US.utf8",
    charset="UTF8")
backstage_app_service_plan = azure_native.web.AppServicePlan("backstageAppServicePlan",
    resource_group_name=resource_group.name,
    kind="Linux",
    sku=azure_native.web.SkuDescriptionArgs(
        name="S1",
        tier="Standard",
    ),
    reserved=True,
    is_spot=False)
backstage_web_app = azure_native.web.WebApp("backstageWebApp",
    name="my-backstage-app",
    resource_group_name=resource_group.name,
    server_farm_id=backstage_app_service_plan.id,
    kind="app,linux,container",
    identity=azure_native.web.ManagedServiceIdentityArgs(
        type=azure_native.web.ManagedServiceIdentityType.SYSTEM_ASSIGNED,
    ),
    site_config=azure_native.web.SiteConfigArgs(
        cors=azure_native.web.CorsSettingsArgs(
            support_credentials=True,
            allowed_origins=["https://my-backstage-app.azurewebsites.net"],
        ),
        http_logging_enabled=True,
        app_settings=[
            azure_native.web.NameValuePairArgs(
                name="POSTGRES_HOST",
                value=backstage_postgresql_server.fully_qualified_domain_name,
            ),
            azure_native.web.NameValuePairArgs(
                name="POSTGRES_PORT",
                value="5432",
            ),
            azure_native.web.NameValuePairArgs(
                name="POSTGRES_USER",
                value=backstage_postgresql_server.administrator_login,
            ),
            azure_native.web.NameValuePairArgs(
                name="POSTGRES_PASSWORD",
                value="1Backstage1!",
            ),
            azure_native.web.NameValuePairArgs(
                name="AZURE_PAT",
                value=azure_dev_ops_token,
            ),
            azure_native.web.NameValuePairArgs(
                name="AZURE_ORG",
                value=azure_dev_ops_organization,
            ),
            azure_native.web.NameValuePairArgs(
                name="APPINSIGHTS_INSTRUMENTATIONKEY",
                value=backstage_app_insights.instrumentation_key,
            ),
            azure_native.web.NameValuePairArgs(
                name="DOCKER_ENABLE_CI",
                value="true",
            ),
            azure_native.web.NameValuePairArgs(
                name="WEBSITES_PORT",
                value="7007",
            ),
            azure_native.web.NameValuePairArgs(
                name="PORT",
                value="8080",
            ),
            azure_native.web.NameValuePairArgs(
                name="BACKSTAGE_BASE_URL",
                value="https://my-backstage-app.azurewebsites.net",
            ),
            azure_native.web.NameValuePairArgs(
                name="APP_CONFIG_backend_database_connection_ssl_required",
                value="true",
            ),
            azure_native.web.NameValuePairArgs(
                name="APP_CONFIG_backend_database_connection_ssl_rejectUnauthorized",
                value="true",
            ),
        ],
        acr_use_managed_identity_creds=True,
        linux_fx_version=backstage_image.repo_digest.apply(lambda repo_digest: f"DOCKER|{repo_digest}"),
    ))
role_assignment3 = azure_native.authorization.RoleAssignment("roleAssignment3",
    principal_id=backstage_web_app.identity.principal_id,
    role_definition_id=acr_pull_role_definition.id,
    scope=backstage_container_registry.id,
    principal_type="ServicePrincipal")
pulumi.export("backstageWebAppUrl", pulumi.Output.concat("https://", backstage_web_app.default_host_name))
pulumi.export("repoDigest", backstage_image.repo_digest)
