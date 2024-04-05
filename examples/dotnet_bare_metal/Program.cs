using System.Collections.Generic;
using System.Linq;
using Pulumi;
using System.IO;
using Aem = WTTech.Aem;

return await Deployment.RunAsync(() =>
{
    var privateKey = File.ReadAllText("ec2-key.cer");

    var aemInstance = new Aem.Compose.Instance("aem_instance", new()
    {
        Client = new Aem.Compose.Inputs.ClientArgs
        {
            Type = "ssh",
            Settings =
            {
                { "host", "x.x.x.x" },
                { "port", "22" },
                { "user", "root" },
                { "secure", "false" },
            },
            Credentials =
            {
                { "private_key", privateKey },
            },
        },
        Files =
        {
            { "lib", "/data/aemc/aem/home/lib" },
        },
    });

    return new Dictionary<string, object?>
    {
        ["output"] = 
        {
            { "aemInstances", aemInstance.Instances },
        },
    };
});

