package tests

import (
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	aem "github.com/wttech/pulumi-aem/provider"
)

func TestInstanceModelCheck(t *testing.T) {
	prov := provider()

	response, err := prov.Check(p.CheckRequest{
		Urn: urn("Instance"),
		News: resource.PropertyMap{
			"client": resource.NewObjectProperty(resource.PropertyMap{
				"type": resource.NewStringProperty("mock"),
				"settings": resource.NewObjectProperty(resource.PropertyMap{
					"setting1": resource.NewStringProperty("value1"),
					"setting2": resource.NewStringProperty("value2"),
					"setting3": resource.NewStringProperty("value3"),
				}),
			}),
		},
	})

	require.NoError(t, err)
	inputs := response.Inputs["system"].V.(resource.PropertyMap)
	result := inputs["data_dir"].StringValue()
	assert.Equal(t, result, "/mnt/aemc")
}

func urn(typ string) resource.URN {
	return resource.NewURN("stack", "proj", "",
		tokens.Type("aem:compose:"+typ), "name")
}

func provider() integration.Server {
	return integration.NewServer(aem.Name, semver.MustParse("1.0.0"), aem.Provider())
}
