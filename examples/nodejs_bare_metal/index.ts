import * as aem from "@wttech/aem";
import * as fs from "fs";

const privateKey = fs.readFileSync("ec2-key.cer", "utf8");

const instanceResourceModel = new aem.compose.InstanceResourceModel("aem_single", {
    client: {
        type: "ssh",
        settings: {
            host: "x.x.x.x",
            port: "22",
            user: "root",
            secure: "false",
        },
        credentials: {
            private_key: privateKey,
        },
    },
    files: {
        lib: "/data/aemc/aem/home/lib",
    },
});

export const output = {
    aemInstances: instanceResourceModel.instances,
};
