package provider

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/wttech/pulumi-aem/provider/utils"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"time"
)

const (
	ServiceName = "aem"
)

type InstanceClient ClientContext[InstanceArgs]

func (ic *InstanceClient) Close() error {
	return ic.cl.Disconnect()
}

func (ic *InstanceClient) dataDir() string {
	return ic.data.System.DataDir
}

func (ic *InstanceClient) prepareWorkDir() error {
	return ic.cl.DirEnsure(ic.cl.WorkDir)
}

func (ic *InstanceClient) prepareDataDir() error {
	return ic.cl.DirEnsure(ic.dataDir())
}

func (ic *InstanceClient) installComposeCLI() error {
	if !ic.data.Compose.Download {
		ic.ctx.Log(diag.Info, "Skipping AEM Compose CLI wrapper download. It is expected to be alternatively installed under the data directory.")
		return nil
	}
	exists, err := ic.cl.FileExists(fmt.Sprintf("%s/aemw", ic.dataDir()))
	if err != nil {
		return fmt.Errorf("cannot check if AEM Compose CLI wrapper is installed: %w", err)
	}
	if !exists {
		ic.ctx.Log(diag.Info, "Downloading AEM Compose CLI wrapper")
		out, err := ic.cl.RunShellCommand("curl -s 'https://raw.githubusercontent.com/wttech/aemc/main/pkg/project/common/aemw' -o 'aemw'", ic.dataDir())
		ic.ctx.Log(diag.Info, string(out))
		if err != nil {
			return fmt.Errorf("cannot download AEM Compose CLI wrapper: %w", err)
		}
		ic.ctx.Log(diag.Info, "Downloaded AEM Compose CLI wrapper")
	}
	return nil
}

func (ic *InstanceClient) writeConfigFile() error {
	configYAML := ic.data.Compose.Config
	if err := ic.cl.FileWrite(fmt.Sprintf("%s/aem/default/etc/aem.yml", ic.dataDir()), configYAML); err != nil {
		return fmt.Errorf("unable to copy AEM configuration file: %w", err)
	}
	return nil
}

func (ic *InstanceClient) copyFiles() error {
	filesMap := ic.data.Files
	for localPath, remotePath := range filesMap {
		if err := ic.cl.PathCopy(localPath, remotePath, true); err != nil {
			return fmt.Errorf("unable to copy path '%s' to '%s': %w", localPath, remotePath, err)
		}
	}
	return nil
}

func (ic *InstanceClient) create() error {
	ic.ctx.Log(diag.Info, "Creating AEM instance(s)")
	if err := ic.configureService(); err != nil {
		return err
	}
	if err := ic.saveProfileScript(); err != nil {
		return err
	}
	if err := ic.runScript("create", ic.data.Compose.Create, ic.dataDir()); err != nil {
		return err
	}
	ic.ctx.Log(diag.Info, "Created AEM instance(s)")
	return nil
}

func (ic *InstanceClient) saveProfileScript() error {
	envFile := fmt.Sprintf("/etc/profile.d/%s.sh", ServiceName)

	systemEnvMap := ic.data.System.Env

	envMap := map[string]string{}
	maps.Copy(envMap, ic.cl.Env)
	maps.Copy(envMap, systemEnvMap)

	ic.cl.Sudo = true
	defer func() { ic.cl.Sudo = false }()

	if err := ic.cl.FileWrite(envFile, utils.EnvToScript(envMap)); err != nil {
		return fmt.Errorf("unable to write AEM environment variables file '%s': %w", envFile, err)
	}
	return nil
}

func (ic *InstanceClient) configureService() error {
	user := ic.data.System.User
	if user == "" {
		user = ic.cl.Connection().User()
	}
	vars := map[string]string{
		"DATA_DIR": ic.dataDir(),
		"USER":     user,
	}

	ic.cl.Sudo = true
	defer func() { ic.cl.Sudo = false }()

	serviceTemplated, err := utils.TemplateString(ic.data.System.ServiceConfig, vars)
	if err != nil {
		return fmt.Errorf("unable to template AEM system service definition: %w", err)
	}
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", ServiceName)
	if err := ic.cl.FileWrite(serviceFile, serviceTemplated); err != nil {
		return fmt.Errorf("unable to write AEM system service definition '%s': %w", serviceFile, err)
	}

	if err := ic.runServiceAction("enable"); err != nil {
		return err
	}
	return nil
}

func (ic *InstanceClient) runServiceAction(action string) error {
	ic.cl.Sudo = true
	defer func() { ic.cl.Sudo = false }()

	outBytes, err := ic.cl.RunShellCommand(fmt.Sprintf("systemctl %s %s.service", action, ServiceName), ".")
	if err != nil {
		return fmt.Errorf("unable to perform AEM system service action '%s': %w", action, err)
	}
	outText := string(outBytes)
	ic.ctx.Log(diag.Info, outText)
	return nil
}

func (ic *InstanceClient) launch() error {
	ic.ctx.Log(diag.Info, "Launching AEM instance(s)")
	if err := ic.runServiceAction("start"); err != nil {
		return err
	}
	if err := ic.applyConfig(); err != nil {
		return err
	}
	if err := ic.runScript("configure", ic.data.Compose.Configure, ic.dataDir()); err != nil {
		return err
	}
	ic.ctx.Log(diag.Info, "Launched AEM instance(s)")
	return nil
}

func (ic *InstanceClient) applyConfig() error {
	ic.ctx.Log(diag.Info, "Applying AEM instance configuration")
	outBytes, err := ic.cl.RunShellCommand("sh aemw instance launch", ic.dataDir())
	if err != nil {
		return fmt.Errorf("unable to apply AEM instance configuration: %w", err)
	}
	outText := string(outBytes)
	ic.ctx.Log(diag.Info, outText)
	ic.ctx.Log(diag.Info, "Applied AEM instance configuration")
	return nil
}

func (ic *InstanceClient) terminate() error {
	ic.ctx.Log(diag.Info, "Terminating AEM instance(s)")
	if err := ic.runServiceAction("stop"); err != nil {
		return err
	}
	if err := ic.runScript("delete", ic.data.Compose.Delete, ic.dataDir()); err != nil {
		return err
	}
	ic.ctx.Log(diag.Info, "Terminated AEM instance(s)")
	return nil
}

func (ic *InstanceClient) deleteDataDir() error {
	if err := ic.cl.PathDelete(ic.dataDir()); err != nil {
		return fmt.Errorf("cannot delete AEM data directory: %w", err)
	}
	return nil
}

type InstanceStatus struct {
	Data struct {
		Instances []struct {
			ID           string   `yaml:"id"`
			URL          string   `yaml:"url"`
			AemVersion   string   `yaml:"aem_version"`
			Attributes   []string `yaml:"attributes"`
			RunModes     []string `yaml:"run_modes"`
			HealthChecks []string `yaml:"health_checks"`
			Dir          string   `yaml:"dir"`
		} `yaml:"instances"`
	}
}

func (ic *InstanceClient) ReadStatus() (InstanceStatus, error) {
	var status InstanceStatus
	yamlBytes, err := ic.cl.RunShellCommand("sh aemw instance status --output-format yaml", ic.dataDir())
	if err != nil {
		return status, err
	}
	if err := yaml.Unmarshal(yamlBytes, &status); err != nil {
		return status, fmt.Errorf("unable to parse AEM instance status: %w", err)
	}
	return status, nil
}

func (ic *InstanceClient) bootstrap() error {
	return ic.doActionOnce("bootstrap", ic.cl.WorkDir, func() error {
		return ic.runScript("bootstrap", ic.data.System.Bootstrap, ".")
	})
}

func (ic *InstanceClient) runScript(name string, script InstanceScript, dir string) error {
	scriptCmd := script.Script
	inlineCmds := script.Inline

	if scriptCmd != "" {
		if err := ic.runScriptMultiline(name, scriptCmd, dir); err != nil {
			return err
		}
	}
	if len(inlineCmds) > 0 {
		if err := ic.runScriptInline(name, inlineCmds, dir); err != nil {
			return err
		}
	}

	return nil
}

func (ic *InstanceClient) runScriptInline(name string, inlineCmds []string, dir string) error {
	for i, cmd := range inlineCmds {
		ic.ctx.Logf(diag.Info, "Executing command '%s' of script '%s' (%d/%d)", cmd, name, i+1, len(inlineCmds))
		textOut, err := ic.cl.RunShellScript(name, cmd, dir)
		if err != nil {
			return fmt.Errorf("unable to execute command '%s' of script '%s' properly: %w", cmd, name, err)
		}
		textStr := string(textOut)
		ic.ctx.Logf(diag.Info, "Executed command '%s' of script '%s' (%d/%d)", cmd, name, i+1, len(inlineCmds))
		ic.ctx.Log(diag.Info, textStr)
	}
	return nil
}

func (ic *InstanceClient) runScriptMultiline(name string, scriptCmd string, dir string) error {
	ic.ctx.Logf(diag.Info, "Executing instance script '%s'", name)
	textOut, err := ic.cl.RunShellScript(name, scriptCmd, dir)
	if err != nil {
		return fmt.Errorf("unable to execute script '%s' properly: %w", name, err)
	}
	textStr := string(textOut)
	ic.ctx.Logf(diag.Info, "Executed instance script '%s'", name)
	ic.ctx.Log(diag.Info, textStr)
	return nil
}

func (ic *InstanceClient) doActionOnce(name string, lockDir string, action func() error) error {
	lock := fmt.Sprintf("%s/provider/%s.lock", lockDir, name)
	exists, err := ic.cl.FileExists(lock)
	if err != nil {
		return fmt.Errorf("cannot read lock file '%s': %w", lock, err)
	}
	if exists {
		ic.ctx.Logf(diag.Info, "Skipping AEM instance action '%s' (lock file already exists '%s')", name, lock)
		return nil
	}
	if err := action(); err != nil {
		return err
	}
	if err := ic.cl.FileWrite(lock, time.Now().String()); err != nil {
		return fmt.Errorf("cannot save lock file '%s': %w", lock, err)
	}
	return nil
}
