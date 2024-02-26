package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-provider-aem/sdk/go/aem/compose"
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

		instanceResourceModel, err := compose.NewInstanceResourceModel(ctx, "aem_single", &compose.InstanceResourceModelArgs{
			Client: compose.ClientModelArgs{
				Type: pulumi.String("aws-ssm"),
				Settings: pulumi.StringMap{
					"instance_id": instance.ID(),
				},
			},
			System: compose.SystemModelArgs{
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
			Compose: compose.ComposeModelArgs{
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

		ctx.Export("instanceIp", instance.PublicIp)
		ctx.Export("aemInstances", instanceResourceModel.Instances)
		return nil
	})
}
