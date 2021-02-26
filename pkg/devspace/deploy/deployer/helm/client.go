package helm

import (
	"fmt"
	"github.com/loft-sh/devspace/assets"
	"github.com/loft-sh/devspace/pkg/devspace/config/constants"
	"github.com/loft-sh/devspace/pkg/devspace/config/generated"
	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/deploy/deployer"
	"github.com/loft-sh/devspace/pkg/devspace/helm"
	helmtypes "github.com/loft-sh/devspace/pkg/devspace/helm/types"
	helmv2 "github.com/loft-sh/devspace/pkg/devspace/helm/v2"
	"github.com/loft-sh/devspace/pkg/devspace/kubectl"
	"github.com/loft-sh/devspace/pkg/util/log"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ComponentChartFolder holds the component charts
const ComponentChartFolder = "component-chart"

// DevSpaceChartConfig is the config that holds the devspace chart information
var DevSpaceChartConfig = &latest.ChartConfig{
	Name:    "component-chart",
	Version: "0.7.1",
	RepoURL: "https://charts.devspace.sh",
}

// DeployConfig holds the information necessary to deploy via helm
type DeployConfig struct {
	// Public because we can switch them to fake clients for testing
	Kube kubectl.Client
	Helm helmtypes.Client

	TillerNamespace  string
	DeploymentConfig *latest.DeploymentConfig
	Log              log.Logger

	config *latest.Config
}

// New creates a new helm deployment client
func New(config *latest.Config, helmClient helmtypes.Client, kubeClient kubectl.Client, deployConfig *latest.DeploymentConfig, log log.Logger) (deployer.Interface, error) {
	tillerNamespace := ""
	if kubeClient != nil {
		tillerNamespace = kubeClient.Namespace()
		if deployConfig.Helm.TillerNamespace != "" {
			tillerNamespace = deployConfig.Helm.TillerNamespace
		}
	}

	// Exchange chart
	if deployConfig.Helm.ComponentChart != nil && *deployConfig.Helm.ComponentChart == true {
		// extract component chart if possible
		filename := "component-chart-" + DevSpaceChartConfig.Version + ".tgz"
		componentChartBytes, err := assets.Asset(filename)
		if err == nil {
			homedir, _ := homedir.Dir()
			completePath := filepath.Join(homedir, constants.DefaultHomeDevSpaceFolder, ComponentChartFolder, filename)
			_, err := os.Stat(completePath)
			if err != nil {
				// make folder
				err = os.MkdirAll(filepath.Dir(completePath), 0755)
				if err != nil {
					return nil, err
				}

				// write file
				err = ioutil.WriteFile(completePath, componentChartBytes, 0666)
				if err != nil {
					return nil, fmt.Errorf("error writing component chart to file: %v", err)
				}
			}

			deployConfig.Helm.Chart = &latest.ChartConfig{
				Name: completePath,
			}
		} else {
			deployConfig.Helm.Chart = DevSpaceChartConfig
		}
	}

	return &DeployConfig{
		Kube:             kubeClient,
		Helm:             helmClient,
		TillerNamespace:  tillerNamespace,
		DeploymentConfig: deployConfig,
		Log:              log,
		config:           config,
	}, nil
}

// Delete deletes the release
func (d *DeployConfig) Delete(cache *generated.CacheConfig) error {
	// Delete with helm engine
	if d.DeploymentConfig.Helm.V2 == true {
		isDeployed := helmv2.IsTillerDeployed(d.Kube, d.TillerNamespace)
		if isDeployed == false {
			return nil
		}
	}

	if d.Helm == nil {
		var err error

		// Get HelmClient
		d.Helm, err = helm.NewClient(d.config, d.DeploymentConfig, d.Kube, d.TillerNamespace, false, false, d.Log)
		if err != nil {
			return errors.Wrap(err, "new helm client")
		}
	}

	err := d.Helm.DeleteRelease(d.DeploymentConfig.Name, d.DeploymentConfig.Namespace, d.DeploymentConfig.Helm)
	if err != nil {
		return err
	}

	// Delete from cache
	delete(cache.Deployments, d.DeploymentConfig.Helm.Chart.Name)
	return nil
}
