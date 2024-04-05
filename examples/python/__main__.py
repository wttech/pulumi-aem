import pulumi
import wttech_aem as aem

aem_instance = aem.compose.Instance("aemInstance",
    client=aem.compose.ClientArgs(
        type="ssh",
        settings={
            "host": "x.x.x.x",
            "port": "22",
            "user": "root",
            "secure": "false",
        },
        credentials={
            "private_key": "[[private_key]]",
        },
    ),
    files={
        "lib": "/data/aemc/aem/home/lib",
    })
pulumi.export("output", {
    "aemInstances": aem_instance.instances,
})
