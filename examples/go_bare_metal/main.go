package main

import (
	_ "embed"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-aem-native/sdk/go/aem/compose"
)

//go:embed ec2-key.cer
var privateKey string

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		instanceResourceModel, err := compose.NewInstanceResourceModel(ctx, "aem_single", &compose.InstanceResourceModelArgs{
			Client: compose.ClientModelArgs{
				Type: pulumi.String("ssh"),
				Settings: pulumi.StringMap{
					"host":   pulumi.String("x.x.x.x"),
					"port":   pulumi.String("22"),
					"user":   pulumi.String("root"),
					"secure": pulumi.String("false"),
				},
				Credentials: pulumi.StringMap{
					"private_key": pulumi.String(privateKey),
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
			"aemInstances": instanceResourceModel.Instances,
		})
		return nil
	})
}
