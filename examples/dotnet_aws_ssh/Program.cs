using System.Collections.Generic;
using System.Linq;
using Pulumi;
using System.IO;
using Pulumi.Aws.Ec2;
using Pulumi.Aws.Ebs;
using Pulumi.Aws.Iam;
using Aem = WTTech.Aem;

return await Deployment.RunAsync(() =>
{
    var configYML = File.ReadAllText("aem.yml");
    var publicKey = File.ReadAllText("ec2-key.cer.pub");
    var privateKey = File.ReadAllText("ec2-key.cer");

    var workspace = "aemc";
    var env = Deployment.Instance.StackName;
    var envType = "tf-minimal";
    var host = "aem-single";
    var dataDevice = "/dev/nvme1n1";
    var dataDir = "/data";
    var composeDir = $"{dataDir}/aemc";
    var sshUser = "ec2-user";

    var tags = new InputMap<string>
    {
        { "Workspace", workspace },
        { "Env", env },
        { "EnvType", envType },
        { "Host", host },
        { "Name", $"{workspace}_{env}_{host}" },
    };

    var role = new Role("aem_ec2", new RoleArgs
    {
        Name = $"{workspace}_{env}_aem_ec2",
        AssumeRolePolicy = @"{
    ""Version"": ""2012-10-17"",
    ""Statement"": {
        ""Effect"": ""Allow"",
        ""Principal"": {""Service"": ""ec2.amazonaws.com""},
        ""Action"": ""sts:AssumeRole""
    }
}",
        Tags = tags,
    });

    var s3PolicyAttachment = new RolePolicyAttachment("s3", new RolePolicyAttachmentArgs
    {
        Role = role.Name,
        PolicyArn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
    });

    var instanceProfile = new InstanceProfile("aem_ec2", new InstanceProfileArgs
    {
        Name = $"{workspace}_{env}_aem_ec2",
        Role = role.Name,
        Tags = tags,
    });

    var keyPair = new KeyPair("aem_single", new KeyPairArgs
    {
        KeyName = $"{workspace}-{env}-example-tf",
        PublicKey = publicKey,
        Tags = tags,
    });

    var instance = new Instance("aem_single", new InstanceArgs
    {
        Ami = "ami-043e06a423cbdca17", // RHEL 8
        InstanceType = "m5.xlarge",
        AssociatePublicIpAddress = true,
        IamInstanceProfile = instanceProfile.Name,
        KeyName = keyPair.KeyName,
        Tags = tags,
    });

    var volume = new Volume("aem_single_data", new VolumeArgs
    {
        AvailabilityZone = instance.AvailabilityZone,
        Size = 128,
        Type = "gp2",
        Tags = tags,
    });

    var volumeAttachment = new VolumeAttachment("aem_single_data", new VolumeAttachmentArgs
    {
        DeviceName = "/dev/xvdf",
        VolumeId = volume.Id,
        InstanceId = instance.Id,
    });

    var aemInstance = new Aem.Compose.Instance("aem_instance", new()
    {
        Client = new Aem.Compose.Inputs.ClientArgs
        {
            Type = "ssh",
            Settings = new InputMap<string>
            {
                { "host", instance.PublicIp },
                { "port", "22" },
                { "user", sshUser },
                { "secure", "false" },
            },
            Credentials = new InputMap<string>
            {
                { "private_key", privateKey },
            },
        },
        System = new Aem.Compose.Inputs.SystemArgs
        {
            Data_dir = composeDir,
            Bootstrap = new InstanceScriptArgs
            {
                Inline = new InputList<string>
                {
                    $"sudo mkfs -t ext4 {dataDevice}",
                    $"sudo mkdir -p {dataDir}",
                    $"sudo mount {dataDevice} {dataDir}",
                    $"echo '{dataDevice} {dataDir} ext4 defaults 0 0' | sudo tee -a /etc/fstab",
                    "sudo yum install -y unzip",
                    "curl 'https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip' -o 'awscliv2.zip'",
                    "unzip -q awscliv2.zip",
                    "sudo ./aws/install --update",
                },
            },
        },
        Compose = new Aem.Compose.Inputs.ComposeArgs
        {
            Config = configYML,
            Create = new InstanceScriptArgs
            {
                Inline = new InputList<string>
                {
                    $"mkdir -p '{composeDir}/aem/home/lib'",
                    $"aws s3 cp --recursive --no-progress 's3://aemc/instance/classic/' '{composeDir}/aem/home/lib'",
                    "sh aemw instance init",
                    "sh aemw instance create",
                },
            },
            Configure = new InstanceScriptArgs
            {
                Inline = new InputList<string>
                {
                    "sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'",
                    "sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'",
                    "sh aemw package deploy --file 'aem/home/lib/aem-service-pkg-6.5.*.0.zip'",
                },
            },
        },
    }, new CustomResourceOptions
    {
        DependsOn = { instance, volumeAttachment },
    });

    return new Dictionary<string, object?>
    {
        ["output"] =
        {
            { "instanceIp", instance.PublicIp },
            { "aemInstances", aemInstance.Instances },
        },
    };
});
