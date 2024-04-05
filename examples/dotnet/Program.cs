using System.Collections.Generic;
using System.Linq;
using Pulumi;
using Aem = WTTech.Aem;

return await Deployment.RunAsync(() => 
{
    var aemInstance = new Aem.Compose.Instance("aemInstance", new()
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
                { "private_key", "[[private_key]]" },
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

