import * as aem from "@wttech/aem";
import * as aws from "@pulumi/aws";
import * as fs from "fs";

const configYML = fs.readFileSync("aem.yml","utf8");
const publicKey = fs.readFileSync("ec2-key.cer.pub","utf8");
const privateKey = fs.readFileSync("ec2-key.cer","utf8");

const workspace = "aemc"
const env = "tf-minimal"
const envType = "aem-single"
const host = "aem-single"
const dataDevice = "/dev/nvme1n1"
const dataDir = "/data"
const composeDir = `${dataDir}/aemc`
const sshUser = "ec2-user"

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

new aws.iam.RolePolicyAttachment("s3", {
    role: role.name,
    policyArn: "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
});

const instanceProfile = new aws.iam.InstanceProfile("aem_ec2", {
    name: `${workspace}_aem_ec2`,
    role: role.name,
    tags: tags,
});

const keyPair = new aws.ec2.KeyPair("aem_single", {
    keyName: `${workspace}-example-tf`,
    publicKey: publicKey,
    tags: tags,
});

const instance = new aws.ec2.Instance("aem_single", {
    ami: "ami-043e06a423cbdca17", // RHEL 8
    instanceType: "m5.xlarge",
    associatePublicIpAddress: true,
    iamInstanceProfile: instanceProfile.name,
    keyName: keyPair.keyName,
    tags: tags,
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

const instanceResourceModel = new aem.compose.InstanceResourceModel("aem_single", {
    client: {
        type: "ssh",
        settings: {
            host: instance.publicIp,
            port: "22",
            user: sshUser,
            secure: "false",
        },
        credentials: {
            private_key: privateKey,
        },
    },
    system: {
        data_dir: composeDir,
        bootstrap: {
            inline: [
                `sudo mkfs -t ext4 ${dataDevice}`,
                `sudo mkdir -p ${dataDir}`,
                `sudo mount ${dataDevice} ${dataDir}`,
                `sudo chown -R ${sshUser} ${dataDir}`,
                `echo '${dataDevice} ${dataDir} ext4 defaults 0 0' | sudo tee -a /etc/fstab`,
                "sudo yum install -y unzip",
                "curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'",
                "unzip -q awscliv2.zip",
                "sudo ./aws/install --update",
            ],
        },
    },
    compose: {
        config: configYML,
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
    aemInstances: instanceResourceModel.instances,
};
