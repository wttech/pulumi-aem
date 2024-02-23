package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	workspace := "aemc"
	env := "tf-minimal"
	envType := "aem-single"
	host := "aem-single"

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

		_, err = ec2.NewVolumeAttachment(ctx, "aem_single_data", &ec2.VolumeAttachmentArgs{
			DeviceName: pulumi.String("/dev/xvdf"),
			VolumeId:   volume.ID(),
			InstanceId: instance.ID(),
		})
		if err != nil {
			return err
		}

		ctx.Export("instanceIp", instance.PublicIp)
		return nil
	})
}
