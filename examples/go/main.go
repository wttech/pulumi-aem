package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-provider-aem/sdk/go/aem/compose"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		myModel, err := compose.NewInstanceResourceModel(ctx, "myModel", &compose.InstanceResourceModelArgs{
			Length: pulumi.Int(24),
		})
		if err != nil {
			return err
		}
		ctx.Export("output", pulumi.StringMap{
			"value": myModel.Result,
		})
		return nil
	})
}
