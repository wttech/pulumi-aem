package main

import (
	_ "embed"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/wttech/pulumi-aem/sdk/go/aem/compose"
)

//go:embed aem.yml
var configYML string

//go:embed ec2-key.cer.pub
var publicKey string

//go:embed ec2-key.cer
var privateKey string

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		workspace := "aemc"
		env := ctx.Stack()
		envType := "tf-minimal"
		host := "aem-single"
		dataDevice := "/dev/nvme1n1"
		dataDir := "/data"
		composeDir := fmt.Sprintf("%s/aemc", dataDir)
		sshUser := "ec2-user"

		tags := pulumi.StringMap{
			"Workspace": pulumi.String(workspace),
			"Env":       pulumi.String(env),
			"EnvType":   pulumi.String(envType),
			"Host":      pulumi.String(host),
			"Name":      pulumi.Sprintf("%s_%s_%s", workspace, env, host),
		}

		role, err := iam.NewRole(ctx, "aem_ec2", &iam.RoleArgs{
			Name: pulumi.Sprintf("%s_%s_aem_ec2", workspace, env),
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

		_, err = iam.NewRolePolicyAttachment(ctx, "s3", &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"),
		})
		if err != nil {
			return err
		}

		instanceProfile, err := iam.NewInstanceProfile(ctx, "aem_ec2", &iam.InstanceProfileArgs{
			Name: pulumi.Sprintf("%s_%s_aem_ec2", workspace, env),
			Role: role.Name,
			Tags: tags,
		})
		if err != nil {
			return err
		}

		keyPair, err := ec2.NewKeyPair(ctx, "aem_single", &ec2.KeyPairArgs{
			KeyName:   pulumi.Sprintf("%s-%s-example-tf", workspace, env),
			PublicKey: pulumi.Sprintf(publicKey),
			Tags:      tags,
		})
		if err != nil {
			return err
		}

		instance, err := ec2.NewInstance(ctx, "aem_single", &ec2.InstanceArgs{
			Ami:                      pulumi.String("ami-043e06a423cbdca17"), // RHEL 8
			InstanceType:             pulumi.String("m5.xlarge"),
			AssociatePublicIpAddress: pulumi.Bool(true),
			IamInstanceProfile:       instanceProfile.Name,
			KeyName:                  keyPair.KeyName,
			Tags:                     tags,
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
				Type: pulumi.String("ssh"),
				Settings: pulumi.StringMap{
					"host":   instance.PublicIp,
					"port":   pulumi.String("22"),
					"user":   pulumi.String(sshUser),
					"secure": pulumi.String("false"),
				},
				Credentials: pulumi.StringMap{
					"private_key": pulumi.String(privateKey),
				},
			},
			System: compose.SystemArgs{
				Data_dir: pulumi.String(composeDir),
				Bootstrap: compose.InstanceScriptArgs{
					Inline: pulumi.StringArray{
						pulumi.Sprintf("sudo mkfs -t ext4 %s", dataDevice),
						pulumi.Sprintf("sudo mkdir -p %s", dataDir),
						pulumi.Sprintf("sudo mount %s %s", dataDevice, dataDir),
						pulumi.Sprintf("sudo chown -R %s %s", sshUser, dataDir),
						pulumi.Sprintf("echo '%s %s ext4 defaults 0 0' | sudo tee -a /etc/fstab", dataDevice, dataDir),
						pulumi.String("sudo yum install -y unzip"),
						pulumi.String("curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'"),
						pulumi.String("unzip -q awscliv2.zip"),
						pulumi.String("sudo ./aws/install --update"),
					},
				},
			},
			Compose: compose.ComposeArgs{
				Config: pulumi.String(configYML),
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
