import pulumi
import pulumi_aws as aws
import wttech_aem as aem

workspace = "aemc"
env = pulumi.get_stack()
envType = "tf-minimal"
host = "aem-single"
dataDevice = "/dev/nvme1n1"
dataDir = "/data"
composeDir = f"{dataDir}/aemc"

tags = {
    "Workspace": workspace,
    "Env": env,
    "EnvType": envType,
    "Host": host,
    "Name": f"{workspace}_{env}_{host}",
}

role = aws.iam.Role("aem_ec2",
                    name=f"{workspace}_{env}_aem_ec2",
                    assume_role_policy="""{
    "Version": "2012-10-17",
    "Statement": {
        "Effect": "Allow",
        "Principal": {"Service": "ec2.amazonaws.com"},
        "Action": "sts:AssumeRole"
    }
}""",
                    tags=tags,
                    )

aws.iam.RolePolicyAttachment("ssm",
                             role=role.name,
                             policy_arn="arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
                             )

aws.iam.RolePolicyAttachment("s3",
                             role=role.name,
                             policy_arn="arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
                             )

instanceProfile = aws.iam.InstanceProfile("aem_ec2",
                                          name=f"{workspace}_{env}_aem_ec2",
                                          role=role.name,
                                          tags=tags,
                                          )

instance = aws.ec2.Instance("aem_single",
                            ami="ami-043e06a423cbdca17",  # RHEL 8
                            instance_type="m5.xlarge",
                            iam_instance_profile=instanceProfile.name,
                            tags=tags,
                            user_data=pulumi.String("""#!/bin/bash
sudo dnf install -y https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/linux_amd64/amazon-ssm-agent.rpm"""),
                            )

volume = aws.ebs.Volume("aem_single_data",
                        availability_zone=instance.availability_zone,
                        size=128,
                        type="gp2",
                        tags=tags,
                        )

volumeAttachment = aws.ec2.VolumeAttachment("aem_single_data",
                                            device_name="/dev/xvdf",
                                            volume_id=volume.id,
                                            instance_id=instance.id,
                                            )

aemInstance = aem.compose.Instance("aem_instance",
                                   client={
                                       "type": "aws-ssm",
                                       "settings": {
                                           "instance_id": instance.id,
                                       },
                                   },
                                   system={
                                       "data_dir": composeDir,
                                       "bootstrap": {
                                           "inline": [
                                               f"sudo mkfs -t ext4 {dataDevice}",
                                               f"sudo mkdir -p {dataDir}",
                                               f"sudo mount {dataDevice} {dataDir}",
                                               f"echo '{dataDevice} {dataDir} ext4 defaults 0 0' | sudo tee -a /etc/fstab",
                                               "sudo yum install -y unzip",
                                               "curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'",
                                               "unzip -q awscliv2.zip",
                                               "sudo ./aws/install --update",
                                           ],
                                       },
                                   },
                                   compose={
                                       "create": {
                                           "inline": [
                                               f"mkdir -p '{composeDir}/aem/home/lib'",
                                               f"aws s3 cp --recursive --no-progress 's3://aemc/instance/classic/' '{composeDir}/aem/home/lib'",
                                               "sh aemw instance init",
                                               "sh aemw instance create",
                                           ],
                                       },
                                       "configure": {
                                           "inline": [
                                               "sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'",
                                               "sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'",
                                               "sh aemw package deploy --file 'aem/home/lib/aem-service-pkg-6.5.*.0.zip'",
                                           ],
                                       },
                                   },
                                   depends_on=[instance, volumeAttachment],
                                   )

pulumi.export("output", {
    "instanceIp": instance.public_ip,
    "aemInstances": aemInstance.instances,
})
