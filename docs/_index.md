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
import * as aws from "@pulumi/aws";

const workspace = "aemc"
const env = "tf-minimal"
const envType = "aem-single"
const host = "aem-single"
const dataDevice = "/dev/nvme1n1"
const dataDir = "/data"
const composeDir = `${dataDir}/aemc`

const tags = {
    "Workspace": workspace,
    "Env": env,
    "EnvType": envType,
    "Host": host,
    "Name": `${workspace}_${envType}_${host}`,
}

const role = new aws.iam.Role("aem_ec2", {
    name: `${workspace}_aem_ec2`,
    assumeRolePolicy: JSON.stringify({
        "Version": "2012-10-17",
        "Statement": {
            "Effect": "Allow",
            "Principal": {"Service": "ec2.amazonaws.com"},
            "Action": "sts:AssumeRole"
        }
    }),
    tags: tags,
});

new aws.iam.RolePolicyAttachment("ssm", {
    role: role.name,
    policyArn: "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
});

new aws.iam.RolePolicyAttachment("s3", {
    role: role.name,
    policyArn: "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
});

const instanceProfile = new aws.iam.InstanceProfile("aem_ec2", {
    name: `${workspace}_aem_ec2`,
    role: role.name,
    tags: tags,
});

const instance = new aws.ec2.Instance("aem_single", {
    ami: "ami-043e06a423cbdca17", // RHEL 8
    instanceType: "m5.xlarge",
    iamInstanceProfile: instanceProfile.name,
    tags: tags,
    userData: `#!/bin/bash
sudo dnf install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm`,
});

const volume = new aws.ebs.Volume("aem_single_data", {
    availabilityZone: instance.availabilityZone,
    size: 128,
    type: "gp2",
    tags: tags,
});

const volumeAttachment = new aws.ec2.VolumeAttachment("aem_single_data", {
    deviceName: "/dev/xvdf",
    volumeId: volume.id,
    instanceId: instance.id,
});

const aemInstance = new aem.compose.Instance("aem_instance", {
    client: {
        type: "aws-ssm",
        settings: {
            instance_id: instance.id,
        },
    },
    system: {
        data_dir: composeDir,
        bootstrap: {
            inline: [
                `sudo mkfs -t ext4 ${dataDevice}`,
                `sudo mkdir -p ${dataDir}`,
                `sudo mount ${dataDevice} ${dataDir}`,
                `echo '${dataDevice} ${dataDir} ext4 defaults 0 0' | sudo tee -a /etc/fstab`,
                "sudo yum install -y unzip",
                "curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'",
                "unzip -q awscliv2.zip",
                "sudo ./aws/install --update",
            ],
        },
    },
    compose: {
        create: {
            inline: [
                `mkdir -p '${composeDir}/aem/home/lib'`,
                `aws s3 cp --recursive --no-progress 's3://aemc/instance/classic/' '${composeDir}/aem/home/lib'`,
                "sh aemw instance init",
                "sh aemw instance create",
            ],
        },
        configure: {
            inline: [
                "sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'",
                "sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'",
                "sh aemw package deploy --file 'aem/home/lib/aem-service-pkg-6.5.*.0.zip'",
            ],
        },
    }
}, {dependsOn: [instance, volumeAttachment]});

export const output = {
    instanceIp: instance.publicIp,
    aemInstances: aemInstance.instances,
};
```

{{% /choosable %}}
{{% choosable language go %}}

```go
package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-aem/sdk/go/aem/compose"
)

func main() {
	workspace := "aemc"
	env := "tf-minimal"
	envType := "aem-single"
	host := "aem-single"
	dataDevice := "/dev/nvme1n1"
	dataDir := "/data"
	composeDir := fmt.Sprintf("%s/aemc", dataDir)

	tags := pulumi.StringMap{
		"Workspace": pulumi.String(workspace),
		"Env":       pulumi.String(env),
		"EnvType":   pulumi.String(envType),
		"Host":      pulumi.String(host),
		"Name":      pulumi.Sprintf("%s_%s_%s", workspace, envType, host),
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		role, err := iam.NewRole(ctx, "aem_ec2", &iam.RoleArgs{
			Name: pulumi.Sprintf("%s_aem_ec2", workspace),
			AssumeRolePolicy: pulumi.String(`{
	"Version": "2012-10-17",
	"Statement": {
		"Effect": "Allow",
		"Principal": {"Service": "ec2.amazonaws.com"},
		"Action": "sts:AssumeRole"
	}
}`),
			Tags: tags,
		})
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicyAttachment(ctx, "ssm", &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
		})
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicyAttachment(ctx, "s3", &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"),
		})
		if err != nil {
			return err
		}

		instanceProfile, err := iam.NewInstanceProfile(ctx, "aem_ec2", &iam.InstanceProfileArgs{
			Name: pulumi.Sprintf("%s_aem_ec2", workspace),
			Role: role.Name,
			Tags: tags,
		})
		if err != nil {
			return err
		}

		instance, err := ec2.NewInstance(ctx, "aem_single", &ec2.InstanceArgs{
			Ami:                pulumi.String("ami-043e06a423cbdca17"), // RHEL 8
			InstanceType:       pulumi.String("m5.xlarge"),
			IamInstanceProfile: instanceProfile.Name,
			Tags:               tags,
			UserData: pulumi.String(`#!/bin/bash
sudo dnf install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm`),
		})
		if err != nil {
			return err
		}

		volume, err := ebs.NewVolume(ctx, "aem_single_data", &ebs.VolumeArgs{
			AvailabilityZone: instance.AvailabilityZone,
			Size:             pulumi.Int(128),
			Type:             pulumi.String("gp2"),
			Tags:             tags,
		})
		if err != nil {
			return err
		}

		volumeAttachment, err := ec2.NewVolumeAttachment(ctx, "aem_single_data", &ec2.VolumeAttachmentArgs{
			DeviceName: pulumi.String("/dev/xvdf"),
			VolumeId:   volume.ID(),
			InstanceId: instance.ID(),
		})
		if err != nil {
			return err
		}

		aemInstance, err := compose.NewInstance(ctx, "aem_instance", &compose.InstanceArgs{
			Client: compose.ClientArgs{
				Type: pulumi.String("aws-ssm"),
				Settings: pulumi.StringMap{
					"instance_id": instance.ID(),
				},
			},
			System: compose.SystemArgs{
				Data_dir: pulumi.String(composeDir),
				Bootstrap: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.Sprintf("sudo mkfs -t ext4 %s", dataDevice),
						pulumi.Sprintf("sudo mkdir -p %s", dataDir),
						pulumi.Sprintf("sudo mount %s %s", dataDevice, dataDir),
						pulumi.Sprintf("echo '%s %s ext4 defaults 0 0' | sudo tee -a /etc/fstab", dataDevice, dataDir),
						pulumi.String("sudo yum install -y unzip"),
						pulumi.String("curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'"),
						pulumi.String("unzip -q awscliv2.zip"),
						pulumi.String("sudo ./aws/install --update"),
					},
				},
			},
			Compose: compose.ComposeArgs{
				Create: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.Sprintf("mkdir -p '%s/aem/home/lib'", composeDir),
						pulumi.Sprintf("aws s3 cp --recursive --no-progress 's3://aemc/instance/classic/' '%s/aem/home/lib'", composeDir),
						pulumi.String("sh aemw instance init"),
						pulumi.String("sh aemw instance create"),
					},
				},
				Configure: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.String("sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'"),
						pulumi.String("sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'"),
						pulumi.String("sh aemw package deploy --file 'aem/home/lib/aem-service-pkg-6.5.*.0.zip'"),
					},
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{instance, volumeAttachment}))
		if err != nil {
			return err
		}

		ctx.Export("output", pulumi.Map{
			"instanceIp":   instance.PublicIp,
			"aemInstances": aemInstance.Instances,
		})
		return nil
	})
}
```

{{% /choosable %}}
{{< /chooser >}}