package wmiexecutor

import (
	"context"
	"encoding/json"

	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckWinService gathers information about Windows Services services
type CheckWinService struct {
	verbose bool
	debug   bool
}

func (c *CheckWinService) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

func (c *CheckWinService) RunCheck() (string, error) {
	check := &checks.CheckWinService{}

	result, err := check.Run(context.Background())

	js, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
