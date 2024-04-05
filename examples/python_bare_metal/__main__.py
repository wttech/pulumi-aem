import pulumi
import wttech_aem as aem
import os

with open("ec2-key.cer", "r") as private_key_file:
    privateKey = private_key_file.read()

aemInstance = aem.compose.Instance("aem_instance",
                                   client=aem.compose.ClientArgs(
                                       type="ssh",
                                       settings={
                                           "host": "x.x.x.x",
                                           "port": "22",
                                           "user": "root",
                                           "secure": "false",
                                       },
                                       credentials={
                                           "private_key": privateKey,
                                       },
                                   ),
                                   files={
                                       "lib": "/data/aemc/aem/home/lib",
                                   })

pulumi.export("output", {
    "aemInstances": aemInstance.instances,
})
