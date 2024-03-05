package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-aem-native/sdk/go/aem/compose"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		aemInstance, err := compose.NewInstance(ctx, "aem_instance", &compose.InstanceArgs{
			Client: &compose.ClientArgs{
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
		ctx.Export("output", pulumi.Map{
			"aemInstances": aemInstance.Instances,
		})
		return nil
	})
}
