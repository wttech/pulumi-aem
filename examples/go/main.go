package main

import (
	"github.com/pulumi/pulumi-aem/sdk/go/aem/compose"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		instanceResourceModel, err := compose.NewInstanceResourceModel(ctx, "instanceResourceModel", &compose.InstanceResourceModelArgs{
			Client: &compose.ClientModelArgs{
				Type: pulumi.String("ssh"),
				Settings: pulumi.StringMap{
					"host":   pulumi.String("x.x.x.x"),
					"port":   pulumi.String("22"),
					"user":   pulumi.String("root"),
					"secure": pulumi.String("false"),
				},
				Credentials: pulumi.StringMap{
					"private_key": pulumi.String("[[private_key]]"),
				},
			},
			Files: pulumi.StringMap{
				"lib": pulumi.String("/data/aemc/aem/home/lib"),
			},
		})
		if err != nil {
			return err
		}
		ctx.Export("output", map[string]interface{}{
			"aemInstances": instanceResourceModel.Instances,
		})
		return nil
	})
}
