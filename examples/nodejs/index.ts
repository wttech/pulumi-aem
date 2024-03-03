import * as pulumi from "@pulumi/pulumi";
import * as aem from "@pulumi/aem";

const instanceResourceModel = new aem.compose.InstanceResourceModel("instanceResourceModel", {
    client: {
        type: "ssh",
        settings: {
            host: "x.x.x.x",
            port: "22",
            user: "root",
            secure: "false",
        },
        credentials: {
            private_key: "[[private_key]]",
        },
    },
    files: {
        lib: "/data/aemc/aem/home/lib",
    },
});
export const output = {
    aemInstances: instanceResourceModel.instances,
};
