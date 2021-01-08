package checkrunner

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
)

func getCommandLine() string {
	customCommandLine := "echo 'hello world'"
	if runtime.GOOS == "windows" {
		customCommandLine = `powershell.exe -command "echo 'hello world'"`
	}
	return customCommandLine
}

func TestRun(t *testing.T) {
	cc := &CustomCheckHandler{
		ResultOutput: make(chan *CustomCheckResult),
		Configuration: []*config.CustomCheck{
			{
				Name:     "check_1",
				Interval: 1,
				Enabled:  true,
				Timeout:  30,
				Command:  getCommandLine(),
			},
		},
	}
	cc.Start(context.Background())

	timeout := time.After(time.Second * 3)
	results := []*utils.CommandResult{}

outerfor:
	for {
		select {
		case <-timeout:
			fmt.Println("timeout")
			go func() {
				cc.Shutdown()
				close(cc.ResultOutput)
			}()
		case res, ok := <-cc.ResultOutput:
			if !ok {
				break outerfor
			}
			//fmt.Println(res)
			results = append(results, res.Result)

		}
	}

	if len(results) < 1 && len(results) > 3 {
		t.Fatal("Custom check was executed to often or to less: ", len(results))
	}

}
