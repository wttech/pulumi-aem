import pulumi
import pulumi_aws as aws
import wttech_aem as aem
import os

workspace = "aemc"
env = "tf-minimal"
envType = "aem-single"
host = "aem-single"
dataDevice = "/dev/nvme1n1"
dataDir = "/data"
composeDir = f"{dataDir}/aemc"
sshUser = "ec2-user"

with open("aem.yml", "r") as config_file:
    configYML = config_file.read()

with open("ec2-key.cer.pub", "r") as public_key_file:
    publicKey = public_key_file.read()

with open("ec2-key.cer", "r") as private_key_file:
    privateKey = private_key_file.read()

tags = {
    "Workspace": workspace,
    "Env": env,
    "EnvType": envType,
    "Host": host,
    "Name": f"{workspace}_{envType}_{host}",
}

role = aws.iam.Role("aem_ec2",
                    name=f"{workspace}_aem_ec2",
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

aws.iam.RolePolicyAttachment("s3",
                             role=role.name,
                             policy_arn="arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
                             )

instanceProfile = aws.iam.InstanceProfile("aem_ec2",
                                          name=f"{workspace}_aem_ec2",
                                          role=role.name,
                                          tags=tags,
                                          )

keyPair = aws.ec2.KeyPair("aem_single",
                          key_name=f"{workspace}-example-tf",
                          public_key=publicKey,
                          tags=tags,
                          )

instance = aws.ec2.Instance("aem_single",
                            ami="ami-043e06a423cbdca17",  # RHEL 8
                            instance_type="m5.xlarge",
                            associate_public_ip_address=True,
                            iam_instance_profile=instanceProfile.name,
                            key_name=keyPair.key_name,
                            tags=tags,
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
                                       "type": "ssh",
                                       "settings": {
                                           "host": instance.public_ip,
                                           "port": "22",
                                           "user": sshUser,
                                           "secure": "false",
                                       },
                                       "credentials": {
                                           "private_key": privateKey,
                                       },
                                   },
                                   system={
                                       "data_dir": composeDir,
                                       "bootstrap": {
                                           "inline": [
                                               f"sudo mkfs -t ext4 {dataDevice}",
                                               f"sudo mkdir -p {dataDir}",
                                               f"sudo mount {dataDevice} {dataDir}",
                                               f"sudo chown -R {sshUser} {dataDir}",
                                               f"echo '{dataDevice} {dataDir} ext4 defaults 0 0' | sudo tee -a /etc/fstab",
                                               "sudo yum install -y unzip",
                                               "curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'",
                                               "unzip -q awscliv2.zip",
                                               "sudo ./aws/install --update",
                                           ],
                                       },
                                   },
                                   compose={
                                       "config": configYML,
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
