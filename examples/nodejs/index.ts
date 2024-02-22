import * as pulumi from "@pulumi/pulumi";
import * as aem from "@pulumi/aem";

const myModel = new aem.InstanceResourceModel("myModel", {length: 24});
export const output = {
    value: myModel.result,
};
