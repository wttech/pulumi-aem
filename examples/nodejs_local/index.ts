import * as aem from "@wttech/aem";
import * as pulumi from "@pulumi/pulumi";
import * as fs from "fs";

const configYML = fs.readFileSync("aem.yml", "utf8");

const env = pulumi.getStack()
const dataDir = `/data/${env}`
const composeDir = `${dataDir}/aemc`
const workDir = `~/tmp/${env}/aemc`
const libraryDir = "~/lib"

new aem.compose.Instance("aem_instance", {
    client: {
        type: "local",
        settings: {},
    },
    system: {
        data_dir: composeDir,
        work_dir: workDir,
        bootstrap: {
            inline: [],
        },
    },
    compose: {
        config: configYML,
        create: {
            inline: [
                `mkdir -p ${composeDir}/aem/home/lib`,
                `cp ${libraryDir}/* ${composeDir}/aem/home/lib`,
                "sh aemw instance init",
                "sh aemw instance create",
            ],
        },
        configure: {
            inline: [
                "sh aemw osgi config save --pid 'org.apache.sling.jcr.davex.impl.servlets.SlingDavExServlet' --input-string 'alias: /crx/server'",
                "sh aemw repl agent setup -A --location 'author' --name 'publish' --input-string '{enabled: true, transportUri: \"http://localhost:4503/bin/receive?sling:authRequestLogin=1\", transportUser: admin, transportPassword: admin, userId: admin}'",
            ],
        },
    }
});

export const output = {
    instanceIp: "127.0.0.1",
};
