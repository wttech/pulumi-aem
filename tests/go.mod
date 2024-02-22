module github.com/wttech/pulumi-provider-aem/tests

go 1.21

replace github.com/wttech/pulumi-provider-aem/provider => ../provider

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/pulumi/pulumi-go-provider v0.14.0
	github.com/pulumi/pulumi-go-provider/integration v0.10.0
	github.com/pulumi/pulumi/sdk/v3 v3.107.0
	github.com/stretchr/testify v1.8.4
	github.com/wttech/pulumi-provider-aem/provider v0.0.0-00010101000000-000000000000
)
