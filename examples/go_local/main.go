package main

import (
	_ "embed"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-aem/sdk/go/aem/compose"
)

//go:embed aem.yml
var configYML string

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		env := ctx.Stack()
		dataDir := fmt.Sprintf("~/data/%s", env)
		composeDir := fmt.Sprintf("%s/aemc", dataDir)
		workDir := fmt.Sprintf("~/tmp/%s/aemc", env)
		libraryDir := "~/lib"

		_, err := compose.NewInstance(ctx, "aem_instance", &compose.InstanceArgs{
			Client: compose.ClientArgs{
				Type:     pulumi.String("local"),
				Settings: pulumi.StringMap{},
			},
			System: compose.SystemArgs{
				Data_dir: pulumi.String(composeDir),
				Work_dir: pulumi.String(workDir),
				Bootstrap: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{},
				},
			},
			Compose: compose.ComposeArgs{
				Config: pulumi.String(configYML),
				Create: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.Sprintf("mkdir -p %s/aem/home/lib", composeDir),
						pulumi.Sprintf("cp %s/* %s/aem/home/lib", libraryDir, composeDir),
						pulumi.String("sh aemw instance init"),
						pulumi.String("sh aemw instance create"),
					},
				},
				Configure: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.String("sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'"),
						pulumi.String("sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("output", pulumi.Map{
			"instanceIp": pulumi.String("127:0:0:1"),
		})
		return nil
	})
}
