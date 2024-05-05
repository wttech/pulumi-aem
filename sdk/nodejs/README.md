[![AEM Compose Logo](https://raw.githubusercontent.com/wttech/pulumi-aem/main/docs/logo-with-text.png)](https://github.com/wttech/aemc)
[![WTT Logo](https://raw.githubusercontent.com/wttech/pulumi-aem/main/docs/wtt-logo.png)](https://www.wundermanthompson.com/service/technology)

[![Apache License, Version 2.0, January 2004](https://raw.githubusercontent.com/wttech/pulumi-aem/main/docs/apache-license-badge.svg)](http://www.apache.org/licenses/)

# AEM Compose - Pulumi Native Provider

This provider allows development teams to easily set up [Adobe Experience Manager (AEM)](https://business.adobe.com/products/experience-manager/adobe-experience-manager.html) instances on virtual machines in the cloud (AWS, Azure, GCP, etc.) or bare metal machines.
It's based on the [AEM Compose](https://github.com/wttech/aemc) tool and aims to simplify the process of creating AEM environments without requiring deep DevOps knowledge.

Published in [Pulumi Registry](https://www.pulumi.com/registry/packages/aem/).

## Purpose

The main purpose of this provider is to enable users to:

- Set up as many AEM environments as needed with minimal effort
- Eliminate the need for deep DevOps knowledge
- Allow for seamless integration with popular cloud platforms such as AWS and Azure
- Provide a simple and efficient way to manage AEM instances

## Features

- Easy configuration and management of AEM instances
- Support for multiple cloud platforms and bare metal machines
- Seamless integration with Pulumi for infrastructure provisioning
- Based on the powerful [AEM Compose](https://github.com/wttech/aemc) tool

## Quickstart

The easiest way to get started is to review, copy and adapt provided examples:

1. AWS EC2 instance with private IP
   * [Go](https://github.com/wttech/pulumi-aem/tree/main/examples/go_aws_ssm)
   * [NodeJS](https://github.com/wttech/pulumi-aem/tree/main/examples/nodejs_aws_ssm)
   * [Python](https://github.com/wttech/pulumi-aem/tree/main/examples/python_aws_ssm)
   * [.NET](https://github.com/wttech/pulumi-aem/tree/main/examples/dotnet_aws_ssm)
2. AWS EC2 instance with public IP
   * [Go](https://github.com/wttech/pulumi-aem/tree/main/examples/go_aws_ssh)
   * [NodeJS](https://github.com/wttech/pulumi-aem/tree/main/examples/nodejs_aws_ssh)
   * [Python](https://github.com/wttech/pulumi-aem/tree/main/examples/python_aws_ssh)
   * [.NET](https://github.com/wttech/pulumi-aem/tree/main/examples/dotnet_aws_ssh)
3. Bare metal machine
   * [Go](https://github.com/wttech/pulumi-aem/tree/main/examples/go_bare_metal)
   * [NodeJS](https://github.com/wttech/pulumi-aem/tree/main/examples/nodejs_bare_metal)
   * [Python](https://github.com/wttech/pulumi-aem/tree/main/examples/python_bare_metal)
   * [.NET](https://github.com/wttech/pulumi-aem/tree/main/examples/dotnet_bare_metal)

- - -

## Development

This repository is showing how to create and locally test a native Pulumi provider.

### Authoring a Pulumi Native Provider

This creates a working Pulumi-owned provider named `aem`.
It implements a random number generator that you can [build and test out for yourself](#test-against-the-example) and then replace the Random code with code specific to your provider.


#### Prerequisites

Prerequisites for this repository are already satisfied by the [Pulumi Devcontainer](https://github.com/pulumi/devcontainer) if you are using Github Codespaces, or VSCode.

If you are not using VSCode, you will need to ensure the following tools are installed and present in your `$PATH`:

* [`pulumictl`](https://github.com/pulumi/pulumictl#installation)
* [Go](https://golang.org/dl/) or 1.latest
* [NodeJS](https://nodejs.org/en/) 14.x.  We recommend using [nvm](https://github.com/nvm-sh/nvm) to manage NodeJS installations.
* [Yarn](https://yarnpkg.com/)
* [TypeScript](https://www.typescriptlang.org/)
* [Python](https://www.python.org/downloads/) (called as `python3`).  For recent versions of MacOS, the system-installed version is fine.
* [.NET](https://dotnet.microsoft.com/download)


#### Build & test the AEM provider

1. Create a new Github CodeSpaces environment using this repository.
1. Open a terminal in the CodeSpaces environment.
1. Run `make build install` to build and install the provider.
1. Run `make gen_examples` to generate the example programs in `examples/` off of the source `examples/yaml` example program.
1. Run `make up` to run the example program in `examples/yaml`.
1. Run `make down` to tear down the example program.

##### Build the provider and install the plugin

   ```bash
   $ make build install
   ```

This will:

1. Create the SDK codegen binary and place it in a `./bin` folder (gitignored)
2. Create the provider binary and place it in the `./bin` folder (gitignored)
3. Generate the dotnet, Go, Node, and Python SDKs and place them in the `./sdk` folder
4. Install the provider on your machine.

##### Test against the example

```bash
$ cd examples/simple
$ yarn link @wttech/aem
$ yarn install
$ pulumi stack init test
$ pulumi up
```

Now that you have completed all of the above steps, you have a working provider that generates a random string for you.

##### A brief repository overview

You now have:

1. A `provider/` folder containing the building and implementation logic
    1. `cmd/pulumi-resource-aem/main.go` - holds the provider's sample implementation logic.
2. `deployment-templates` - a set of files to help you around deployment and publication
3. `sdk` - holds the generated code libraries created by `pulumi-gen-aem/main.go`
4. `examples` a folder of Pulumi programs to try locally and/or use in CI.
5. A `Makefile` and this `README`.

##### Additional Details

This repository depends on the pulumi-go-provider library. For more details on building providers, please check
the [Pulumi Go Provider docs](https://github.com/pulumi/pulumi-go-provider).

#### Build Examples

Create an example program using the resources defined in your provider, and place it in the `examples/` folder.

You can now repeat the steps for [build, install, and test](#test-against-the-example).

### Configuring CI and releases

1. Follow the instructions laid out in the [deployment templates](./deployment-templates/README-DEPLOYMENT.md).

### References

Other resources/examples for implementing providers:
* [Pulumi Command provider](https://github.com/pulumi/pulumi-command/blob/master/provider/pkg/provider/provider.go)
* [Pulumi Go Provider repository](https://github.com/pulumi/pulumi-go-provider)
