package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/schema"
	dotnetGen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	goGen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsGen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pythonGen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/wttech/pulumi-aem/provider/instance"
)

var Version string

const Name string = "aem"

func Provider() p.Provider {
	return infer.Provider(infer.Options{
		Resources: []infer.InferredResource{
			infer.Resource[Instance, InstanceArgs, InstanceState](),
		},
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "compose",
		},
		Metadata: schema.Metadata{
			LanguageMap: map[string]any{
				"nodejs": nodejsGen.NodePackageInfo{
					PackageName: "@wttech/aem",
					Dependencies: map[string]string{
						"@pulumi/pulumi": "^3.56.0", // Includes support for Github & Gitlab releases
					},
					DevDependencies: map[string]string{
						"@types/node": "^10.0.0", // so we can access strongly typed node definitions.
						"@types/mime": "^2.0.0",
					},
				},
				"go": goGen.GoPackageInfo{
					ImportBasePath:                 "github.com/wttech/pulumi-aem-native/sdk/go/aem",
					GenerateResourceContainerTypes: true,
				},
				"python": pythonGen.PackageInfo{
					PackageName: "wttech_pulumi_aem",
					Requires: map[string]string{
						"pulumi": ">=3.56.0,<4.0.0",
					},
				},
				"csharp": dotnetGen.CSharpPackageInfo{
					RootNamespace: "WTTech",
					PackageReferences: map[string]string{
						"Pulumi": "[3.56.0,4)",
					},
				},
			},
		},
	})
}

type Instance struct{}

type InstanceArgs struct {
	Client  Client            `pulumi:"client"`
	Files   map[string]string `pulumi:"files,optional"`
	System  System            `pulumi:"system,optional"`
	Compose Compose           `pulumi:"compose,optional"`
}

func (m *InstanceArgs) Annotate(a infer.Annotator) {
	a.Describe(&m.Client, "Connection settings used to access the machine on which the AEM instance will be running.")
	a.Describe(&m.Files, "Files or directories to be copied into the machine.")
	a.Describe(&m.System, "Operating system configuration for the machine on which AEM instance will be running.")
	a.Describe(&m.Compose, "AEM Compose CLI configuration. See documentation(https://github.com/wttech/aemc#configuration).")
}

type Client struct {
	Type          string            `pulumi:"type"`
	Settings      map[string]string `pulumi:"settings"`
	Credentials   map[string]string `pulumi:"credentials,optional"`
	ActionTimeout string            `pulumi:"action_timeout,optional"`
	StateTimeout  string            `pulumi:"state_timeout,optional"`
}

func (m *Client) Annotate(a infer.Annotator) {
	a.Describe(&m.Type, "Type of connection to use to connect to the machine on which AEM instance will be running.")
	a.Describe(&m.Settings, "Settings for the connection type")
	a.Describe(&m.Credentials, "Credentials for the connection type")
	a.Describe(&m.ActionTimeout, "Used when trying to connect to the AEM instance machine (often right after creating it). Need to be enough long because various types of connections (like AWS SSM or SSH) may need some time to boot up the agent.")
	a.Describe(&m.StateTimeout, "Used when reading the AEM instance state when determining the plan.")
}

type System struct {
	DataDir       string            `pulumi:"data_dir,optional"`
	WorkDir       string            `pulumi:"work_dir,optional"`
	Env           map[string]string `pulumi:"env,optional"`
	ServiceConfig string            `pulumi:"service_config,optional"`
	User          string            `pulumi:"user,optional"`
	Bootstrap     InstanceScript    `pulumi:"bootstrap,optional"`
}

func (m *System) Annotate(a infer.Annotator) {
	a.Describe(&m.DataDir, "Remote root path in which AEM Compose files and unpacked AEM instances will be stored.")
	a.Describe(&m.WorkDir, "Remote root path where provider-related files will be stored.")
	a.Describe(&m.Env, "Environment variables for AEM instances.")
	a.Describe(&m.ServiceConfig, "Contents of the AEM system service definition file (systemd).")
	a.Describe(&m.User, "System user under which AEM instance will be running. By default, the same as the user used to connect to the machine.")
	a.Describe(&m.Bootstrap, "Script executed once upon instance connection, often for mounting on VM data volumes from attached disks (e.g., AWS EBS, Azure Disk Storage). This script runs only once, even during instance recreation, as changes are typically persistent and system-wide. If re-execution is needed, it is recommended to set up a new machine.")
}

type Compose struct {
	Download  bool           `pulumi:"download,optional"`
	Version   string         `pulumi:"version,optional"`
	Config    string         `pulumi:"config,optional"`
	Create    InstanceScript `pulumi:"create,optional"`
	Configure InstanceScript `pulumi:"configure,optional"`
	Delete    InstanceScript `pulumi:"delete,optional"`
}

func (m *Compose) Annotate(a infer.Annotator) {
	a.Describe(&m.Download, "Toggle automatic AEM Compose CLI wrapper download. If set to false, assume the wrapper is present in the data directory.")
	a.Describe(&m.Version, "Version of AEM Compose tool to use on remote machine.")
	a.Describe(&m.Config, "Contents of the AEM Compose YML configuration file.")
	a.Describe(&m.Create, "Script(s) for creating an instance or restoring it from a backup. Typically customized to provide AEM library files (quickstart.jar, license.properties, service packs) from alternative sources (e.g., AWS S3, Azure Blob Storage). Instance recreation is forced if changed.")
	a.Describe(&m.Configure, "Script(s) for configuring a launched instance. Must be idempotent as it is executed always when changed. Typically used for installing AEM service packs, setting up replication agents, etc.")
	a.Describe(&m.Delete, "Script(s) for deleting a stopped instance.")
}

type InstanceScript struct {
	Inline []string `pulumi:"inline,optional"`
	Script string   `pulumi:"script,optional"`
}

func (m *InstanceScript) Annotate(a infer.Annotator) {
	a.Describe(&m.Inline, "Inline shell commands to be executed")
	a.Describe(&m.Script, "Multiline shell script to be executed")
}

type InstanceModel struct {
	ID         string   `pulumi:"id"`
	URL        string   `pulumi:"url"`
	AemVersion string   `pulumi:"aem_version"`
	Dir        string   `pulumi:"dir"`
	Attributes []string `pulumi:"attributes"`
	RunModes   []string `pulumi:"run_modes"`
}

func (m *InstanceModel) Annotate(a infer.Annotator) {
	a.Describe(&m.ID, "Unique identifier of AEM instance defined in the configuration.")
	a.Describe(&m.URL, "The machine-internal HTTP URL address used for communication with the AEM instance.")
	a.Describe(&m.AemVersion, "Version of the AEM instance. Reflects service pack installations.")
	a.Describe(&m.Dir, "Remote path in which AEM instance is stored.")
	a.Describe(&m.Attributes, "A brief description of the state details for a specific AEM instance. Possible states include 'created', 'uncreated', 'running', 'unreachable', 'up-to-date', and 'out-of-date'.")
	a.Describe(&m.RunModes, "A list of run modes for a specific AEM instance.")
}

type InstanceState struct {
	InstanceArgs
	Instances []InstanceModel `pulumi:"instances"`
}

func (m *InstanceState) Annotate(a infer.Annotator) {
	a.Describe(&m.Instances, "Current state of the configured AEM instances.")
}

func (Instance) Create(ctx p.Context, name string, input InstanceArgs, preview bool) (string, InstanceState, error) {
	state := InstanceState{InstanceArgs: input}
	if preview {
		return name, state, nil
	}

	instanceResource := NewInstanceResource()
	status, err := instanceResource.Create(ctx, input)
	if err != nil {
		return name, state, err
	}

	var instances []InstanceModel
	for _, item := range status.Data.Instances {
		instances = append(instances, InstanceModel{
			ID:         item.ID,
			URL:        item.URL,
			AemVersion: item.AemVersion,
			Dir:        item.Dir,
			Attributes: item.Attributes,
			RunModes:   item.RunModes,
		})
	}
	state.Instances = instances

	return name, state, nil
}

func (Instance) Update(ctx p.Context, id string, oldState InstanceState, input InstanceArgs, preview bool) (InstanceState, error) {
	if preview {
		return oldState, nil
	}

	state := InstanceState{InstanceArgs: input}
	instanceResource := NewInstanceResource()
	status, err := instanceResource.Update(ctx, input)
	if err != nil {
		return state, err
	}

	var instances []InstanceModel
	for _, item := range status.Data.Instances {
		instances = append(instances, InstanceModel{
			ID:         item.ID,
			URL:        item.URL,
			AemVersion: item.AemVersion,
			Dir:        item.Dir,
			Attributes: item.Attributes,
			RunModes:   item.RunModes,
		})
	}
	state.Instances = instances

	return state, nil
}

func (Instance) Delete(ctx p.Context, id string, props InstanceState) error {
	instanceResource := NewInstanceResource()
	if err := instanceResource.Delete(ctx, props.InstanceArgs); err != nil {
		return err
	}

	return nil
}

func (Instance) Check(ctx p.Context, name string, oldInputs, newInputs resource.PropertyMap) (InstanceArgs, []p.CheckFailure, error) {
	inputs := determineInputs(newInputs, "client")
	setDefaultValue(inputs, "credentials", resource.NewObjectProperty(resource.PropertyMap{}))
	setDefaultValue(inputs, "action_timeout", resource.NewStringProperty("10m"))
	setDefaultValue(inputs, "state_timeout", resource.NewStringProperty("30s"))

	_ = determineInputs(newInputs, "files")

	inputs = determineInputs(newInputs, "system")
	setDefaultInlineScripts(inputs, "bootstrap", []string{})
	setDefaultValue(inputs, "data_dir", resource.NewStringProperty("/mnt/aemc"))
	setDefaultValue(inputs, "work_dir", resource.NewStringProperty("/tmp/aemc"))
	setDefaultValue(inputs, "service_config", resource.NewStringProperty(instance.ServiceConf))
	setDefaultValue(inputs, "user", resource.NewStringProperty(""))
	setDefaultValue(inputs, "env", resource.NewObjectProperty(resource.PropertyMap{}))

	inputs = determineInputs(newInputs, "compose")
	setDefaultValue(inputs, "download", resource.NewBoolProperty(true))
	setDefaultValue(inputs, "version", resource.NewStringProperty("1.6.12"))
	setDefaultValue(inputs, "config", resource.NewStringProperty(instance.ConfigYML))
	setDefaultInlineScripts(inputs, "create", instance.CreateScriptInline)
	setDefaultInlineScripts(inputs, "configure", instance.LaunchScriptInline)
	setDefaultInlineScripts(inputs, "delete", instance.DeleteScriptInline)

	return infer.DefaultCheck[InstanceArgs](newInputs)
}

func determineInputs(allInputs resource.PropertyMap, key resource.PropertyKey) resource.PropertyMap {
	if inputs, ok := allInputs[key]; ok {
		return inputs.V.(resource.PropertyMap)
	} else {
		inputs = resource.NewObjectProperty(resource.PropertyMap{})
		allInputs[key] = inputs
		return inputs.V.(resource.PropertyMap)
	}
}

func setDefaultValue(inputs resource.PropertyMap, key resource.PropertyKey, value resource.PropertyValue) {
	if _, ok := inputs[key]; !ok {
		inputs[key] = value
	}
}

func setDefaultInlineScripts(allInputs resource.PropertyMap, key resource.PropertyKey, scripts []string) {
	inputs := determineInputs(allInputs, key)
	if !inputs.HasValue("inline") && !inputs.HasValue("script") {
		var wrappedScripts []resource.PropertyValue
		for _, script := range scripts {
			wrappedScripts = append(wrappedScripts, resource.NewStringProperty(script))
		}
		inputs["inline"] = resource.NewArrayProperty(wrappedScripts)
	}
}
