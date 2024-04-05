import * as pulumi from "@pulumi/pulumi";
import * as aem from "@wttech/aem";

const aemInstance = new aem.compose.Instance("aemInstance", {
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
    aemInstances: aemInstance.instances,
};
