package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/spf13/cast"
	"github.com/wttech/pulumi-aem/provider/client"
	"golang.org/x/exp/maps"
	"time"
)

func NewInstanceResource() *InstanceResource {
	return &InstanceResource{
		clientManager: client.ManagerDefault,
	}
}

type InstanceResource struct {
	clientManager *client.Manager
}

func (r *InstanceResource) Create(ctx p.Context, model InstanceArgs) (*InstanceStatus, error) {
	return r.createOrUpdate(ctx, model, true)
}

func (r *InstanceResource) Update(ctx p.Context, model InstanceArgs) (*InstanceStatus, error) {
	return r.createOrUpdate(ctx, model, false)
}

func (r *InstanceResource) createOrUpdate(ctx p.Context, model InstanceArgs, create bool) (*InstanceStatus, error) {
	ctx.Log(diag.Info, "Started setting up AEM instance resource")

	ic, err := r.client(ctx, model, cast.ToDuration(model.Client.ActionTimeout))
	if err != nil {
		ctx.Logf(diag.Error, "Unable to connect to AEM instance %s", err)
		return nil, err
	}
	defer func(ic *InstanceClient) {
		err := ic.Close()
		if err != nil {
			ctx.Logf(diag.Warning, "Unable to disconnect from AEM instance %s", err)
		}
	}(ic)

	if create {
		if err := ic.bootstrap(); err != nil {
			ctx.Logf(diag.Error, "Unable to bootstrap AEM instance machine %s", err)
			return nil, err
		}
	}
	if err := ic.copyFiles(); err != nil {
		ctx.Logf(diag.Error, "Unable to copy AEM instance files %s", err)
		return nil, err
	}
	if err := ic.prepareWorkDir(); err != nil {
		ctx.Logf(diag.Error, "Unable to prepare AEM work directory %s", err)
		return nil, err
	}
	if err := ic.prepareDataDir(); err != nil {
		ctx.Logf(diag.Error, "Unable to prepare AEM data directory %s", err)
		return nil, err
	}
	if err := ic.installComposeCLI(); err != nil {
		ctx.Logf(diag.Error, "Unable to install AEM Compose CLI %s", err)
		return nil, err
	}
	if err := ic.writeConfigFile(); err != nil {
		ctx.Logf(diag.Error, "Unable to write AEM configuration file %s", err)
		return nil, err
	}
	if create {
		if err := ic.create(); err != nil {
			ctx.Logf(diag.Error, "Unable to create AEM instance %s", err)
			return nil, err
		}
	}
	if err := ic.launch(); err != nil {
		ctx.Logf(diag.Error, "Unable to launch AEM instance %s", err)
		return nil, err
	}

	ctx.Log(diag.Info, "Finished setting up AEM instance resource")

	status, err := ic.ReadStatus()
	if err != nil {
		ctx.Logf(diag.Error, "Unable to read AEM instance status %s", err)
		return nil, err
	}

	return &status, nil
}

func (r *InstanceResource) Delete(ctx p.Context, model InstanceArgs) error {
	ctx.Log(diag.Info, "Started deleting AEM instance resource")

	ic, err := r.client(ctx, model, cast.ToDuration(model.Client.StateTimeout))
	if err != nil {
		ctx.Logf(diag.Error, "Unable to connect to AEM instance %s", err)
		return err
	}
	defer func(ic *InstanceClient) {
		err := ic.Close()
		if err != nil {
			ctx.Logf(diag.Warning, "Unable to disconnect from AEM instance %s", err)
		}
	}(ic)

	if err := ic.terminate(); err != nil {
		ctx.Logf(diag.Error, "Unable to terminate AEM instance %s", err)
		return err
	}

	if err := ic.deleteDataDir(); err != nil {
		ctx.Logf(diag.Error, "Unable to delete AEM data directory %s", err)
		return err
	}

	ctx.Log(diag.Info, "Finished deleting AEM instance resource")
	return nil
}

func (r *InstanceResource) client(ctx p.Context, model InstanceArgs, timeout time.Duration) (*InstanceClient, error) {
	typeName := model.Client.Type
	ctx.Logf(diag.Info, "Connecting to AEM instance machine using %s", typeName)

	cl, err := r.clientManager.Make(typeName, r.clientSettings(model))
	if err != nil {
		return nil, err
	}

	if err := cl.ConnectWithRetry(timeout, func() { ctx.Log(diag.Info, "Awaiting connection to AEM instance machine") }); err != nil {
		return nil, err
	}

	cl.Env["AEM_CLI_VERSION"] = model.Compose.Version
	cl.Env["AEM_OUTPUT_LOG_MODE"] = "both"
	cl.WorkDir = model.System.WorkDir

	if err := cl.SetupEnv(); err != nil {
		return nil, err
	}

	ctx.Logf(diag.Info, "Connected to AEM instance machine using %s", cl.Connection().Info())
	return &InstanceClient{cl, ctx, model}, nil
}

func (r *InstanceResource) clientSettings(model InstanceArgs) map[string]string {
	settings := model.Client.Settings
	credentials := model.Client.Credentials

	combined := map[string]string{}
	maps.Copy(combined, credentials)
	maps.Copy(combined, settings)
	return combined
}
