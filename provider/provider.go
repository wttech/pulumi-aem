package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/wttech/pulumi-aem/provider/instance"
)

var Version string

const Name string = "aem"

func Provider() p.Provider {
	return infer.Provider(infer.Options{
		Resources: []infer.InferredResource{
			infer.Resource[InstanceResourceModel, InstanceResourceModelArgs, InstanceResourceModelState](),
		},
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "compose",
		},
	})
}

type InstanceResourceModel struct{}

type InstanceResourceModelArgs struct {
	Client  ClientModel       `pulumi:"client"`
	Files   map[string]string `pulumi:"files,optional"`
	System  SystemModel       `pulumi:"system,optional"`
	Compose ComposeModel      `pulumi:"compose,optional"`
}

type ClientModel struct {
	Type          string            `pulumi:"type"`
	Settings      map[string]string `pulumi:"settings"`
	Credentials   map[string]string `pulumi:"credentials,optional"`
	ActionTimeout string            `pulumi:"action_timeout,optional"`
	StateTimeout  string            `pulumi:"state_timeout,optional"`
}

type SystemModel struct {
	DataDir       string            `pulumi:"data_dir,optional"`
	WorkDir       string            `pulumi:"work_dir,optional"`
	Env           map[string]string `pulumi:"env,optional"`
	ServiceConfig string            `pulumi:"service_config,optional"`
	User          string            `pulumi:"user,optional"`
	Bootstrap     InstanceScript    `pulumi:"bootstrap,optional"`
}

type ComposeModel struct {
	Download  bool           `pulumi:"download,optional"`
	Version   string         `pulumi:"version,optional"`
	Config    string         `pulumi:"config,optional"`
	Create    InstanceScript `pulumi:"create,optional"`
	Configure InstanceScript `pulumi:"configure,optional"`
	Delete    InstanceScript `pulumi:"delete,optional"`
}

type InstanceScript struct {
	Inline []string `pulumi:"inline,optional"`
	Script string   `pulumi:"script,optional"`
}

type InstanceModel struct {
	ID         string   `pulumi:"id"`
	URL        string   `pulumi:"url"`
	AemVersion string   `pulumi:"aem_version"`
	Dir        string   `pulumi:"dir"`
	Attributes []string `pulumi:"attributes"`
	RunModes   []string `pulumi:"run_modes"`
}

type InstanceResourceModelState struct {
	InstanceResourceModelArgs
	Instances []InstanceModel `pulumi:"instances"`
}

func (InstanceResourceModel) Create(ctx p.Context, name string, input InstanceResourceModelArgs, preview bool) (string, InstanceResourceModelState, error) {
	state := InstanceResourceModelState{InstanceResourceModelArgs: input}
	if preview {
		return name, state, nil
	}

	instanceResource := NewInstanceResource()
	status, err := instanceResource.CreateOrUpdate(ctx, input)
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

func (InstanceResourceModel) Delete(ctx p.Context, id string, props InstanceResourceModelState) error {
	instanceResource := NewInstanceResource()
	if err := instanceResource.Delete(ctx, props.InstanceResourceModelArgs); err != nil {
		return err
	}

	return nil
}

func (InstanceResourceModel) Check(ctx p.Context, name string, oldInputs, newInputs resource.PropertyMap) (InstanceResourceModelArgs, []p.CheckFailure, error) {
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

	return infer.DefaultCheck[InstanceResourceModelArgs](newInputs)
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
