package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/spf13/cast"
	"github.com/wttech/pulumi-aem/provider/client"
	"golang.org/x/exp/maps"
	"time"
)

func NewInstanceResource() *InstanceResource {
	return &InstanceResource{
		clientManager: client.ClientManagerDefault,
	}
}

type InstanceResource struct {
	clientManager *client.ClientManager
}

func (r *InstanceResource) Create(log p.Logger, model InstanceArgs) (*InstanceStatus, error) {
	return r.createOrUpdate(log, model, true)
}

func (r *InstanceResource) Update(log p.Logger, model InstanceArgs) (*InstanceStatus, error) {
	return r.createOrUpdate(log, model, false)
}

func (r *InstanceResource) createOrUpdate(log p.Logger, model InstanceArgs, create bool) (*InstanceStatus, error) {
	log.Info("Started setting up AEM instance resource")

	ic, err := r.client(log, model, cast.ToDuration(model.Client.ActionTimeout))
	if err != nil {
		log.Errorf("Unable to connect to AEM instance %s", err)
		return nil, err
	}
	defer func(ic *InstanceClient) {
		err := ic.Close()
		if err != nil {
			log.Warningf("Unable to disconnect from AEM instance %s", err)
		}
	}(ic)

	if create {
		if err := ic.bootstrap(); err != nil {
			log.Errorf("Unable to bootstrap AEM instance machine %s", err)
			return nil, err
		}
	}
	if err := ic.copyFiles(); err != nil {
		log.Errorf("Unable to copy AEM instance files %s", err)
		return nil, err
	}
	if err := ic.prepareWorkDir(); err != nil {
		log.Errorf("Unable to prepare AEM work directory %s", err)
		return nil, err
	}
	if err := ic.prepareDataDir(); err != nil {
		log.Errorf("Unable to prepare AEM data directory %s", err)
		return nil, err
	}
	if err := ic.installComposeCLI(); err != nil {
		log.Errorf("Unable to install AEM Compose CLI %s", err)
		return nil, err
	}
	if err := ic.writeConfigFile(); err != nil {
		log.Errorf("Unable to write AEM configuration file %s", err)
		return nil, err
	}
	if create {
		if err := ic.create(); err != nil {
			log.Errorf("Unable to create AEM instance %s", err)
			return nil, err
		}
	}
	if err := ic.launch(); err != nil {
		log.Errorf("Unable to launch AEM instance %s", err)
		return nil, err
	}

	log.Info("Finished setting up AEM instance resource")

	status, err := ic.ReadStatus()
	if err != nil {
		log.Errorf("Unable to read AEM instance status %s", err)
		return nil, err
	}

	return &status, nil
}

func (r *InstanceResource) Delete(log p.Logger, model InstanceArgs) error {
	log.Info("Started deleting AEM instance resource")

	ic, err := r.client(log, model, cast.ToDuration(model.Client.StateTimeout))
	if err != nil {
		log.Errorf("Unable to connect to AEM instance %s", err)
		return err
	}
	defer func(ic *InstanceClient) {
		err := ic.Close()
		if err != nil {
			log.Warningf("Unable to disconnect from AEM instance %s", err)
		}
	}(ic)

	if err := ic.terminate(); err != nil {
		log.Errorf("Unable to terminate AEM instance %s", err)
		return err
	}

	if err := ic.deleteDataDir(); err != nil {
		log.Errorf("Unable to delete AEM data directory %s", err)
		return err
	}

	log.Info("Finished deleting AEM instance resource")
	return nil
}

func (r *InstanceResource) client(log p.Logger, model InstanceArgs, timeout time.Duration) (*InstanceClient, error) {
	typeName := model.Client.Type
	log.Infof("Connecting to AEM instance machine using %s", typeName)

	cl, err := r.clientManager.Make(typeName, r.clientSettings(model))
	if err != nil {
		return nil, err
	}

	if err := cl.ConnectWithRetry(timeout, func() { log.Info("Awaiting connection to AEM instance machine") }); err != nil {
		return nil, err
	}

	cl.Env["AEM_CLI_VERSION"] = model.Compose.Version
	cl.Env["AEM_OUTPUT_LOG_MODE"] = "both"
	cl.WorkDir = model.System.WorkDir

	if err := cl.SetupEnv(); err != nil {
		return nil, err
	}

	log.Infof("Connected to AEM instance machine using %s", cl.Connection().Info())
	return &InstanceClient{cl, log, model}, nil
}

func (r *InstanceResource) clientSettings(model InstanceArgs) map[string]string {
	settings := model.Client.Settings
	credentials := model.Client.Credentials

	combined := map[string]string{}
	maps.Copy(combined, credentials)
	maps.Copy(combined, settings)
	return combined
}
