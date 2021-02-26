package testing

import (
	"io"

	"github.com/loft-sh/devspace/pkg/devspace/config/versions/latest"
	"github.com/loft-sh/devspace/pkg/devspace/deploy"
	"github.com/loft-sh/devspace/pkg/util/log"
)

// FakeController is the fake build controller
type FakeController struct{}

// NewFakeController creates a new fake build controller
func NewFakeController(config *latest.Config) deploy.Controller {
	return &FakeController{}
}

// Deploy deploys the deployments
func (f *FakeController) Deploy(options *deploy.Options, log log.Logger) error {
	return nil
}

// Render implements interface
func (f *FakeController) Render(options *deploy.Options, out io.Writer, log log.Logger) error {
	return nil
}

// Purge purges the deployments
func (f *FakeController) Purge(deployments []string, log log.Logger) error {
	return nil
}
