---
title: AEM
meta_desc: Provides an overview of the AEM Provider for Pulumi.
layout: overview
---

## AEM Provider for Pulumi

This provider allows development teams to easily set up [Adobe Experience Manager (AEM)](https://business.adobe.com/products/experience-manager/adobe-experience-manager.html) instances on virtual machines in the cloud (AWS, Azure, GCP, etc.) or bare metal machines.
It's based on the [AEM Compose](https://github.com/wttech/aemc) tool and aims to simplify the process of creating AEM environments without requiring deep DevOps knowledge.

## Example

{{< chooser language "typescript,go" >}}
{{% choosable language typescript %}}

```typescript
import * as aem from "@wttech/aem";
import * as fs from "fs";

const privateKey = fs.readFileSync("ec2-key.cer", "utf8");

const aemInstance = new aem.compose.Instance("aem_instance", {
    client: {
        type: "ssh",
        settings: {
            host: "x.x.x.x",
            port: "22",
            user: "root",
            secure: "false",
        },
        credentials: {
            private_key: privateKey,
        },
    },
    files: {
        lib: "/data/aemc/aem/home/lib",
    },
});

export const output = {
    aemInstances: aemInstance.instances,
};
```

{{% /choosable %}}
{{% choosable language go %}}

```go
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
		aemInstance, err := compose.NewInstance(ctx, "aem_instance", &compose.InstanceArgs{
			Client: compose.ClientArgs{
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
			"aemInstances": aemInstance.Instances,
		})
		return nil
	})
}
```

{{% /choosable %}}
{{< /chooser >}}