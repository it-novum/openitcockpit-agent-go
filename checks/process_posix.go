// +build linux darwin

package checks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	/*PID  PPID  %CPU %MEM USER             STAT NI    RSS      VSZ PAGEIN COMM             COMMAND
	 *  1     0   0,1  0,1 root             Ss    0  21340  4313476      0 /sbin/launchd    /sbin/launchd
	 *109     1   0,0  0,0 root             Ss    0   1136  4306048      0 /usr/sbin/syslog /usr/sbin/syslogd
	 *110     1   0,0  0,1 root             Ss    0  11456  4336292      0 /usr/libexec/Use /usr/libexec/UserEventAgent (System)
	 *112     1   0,0  0,0 root             Ss    0   2252  4296900      0 /System/Library/ /System/Library/PrivateFrameworks/Uninstall.framework/Resources/uninstalld
	 *113     1   0,0  0,1 root             Ss    0  20908  4857808      0 /usr/libexec/kex /usr/libexec/kextd
	 *114     1   0,0  0,0 root             Ss    0   8288  4324792      0 /System/Library/ /System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/FSEvents.framework/Versions/A/Support/fseventsd
	 *116     1   0,0  0,1 root             Ss    0  12508  4336452      0 /System/Library/ /System/Library/PrivateFrameworks/MediaRemote.framework/Support/mediaremoted
	 *119     1   0,0  0,1 root             Ss    0  13780  4344220      0 /usr/sbin/system /usr/sbin/systemstats --daemon
	 *120     1   0,0  0,1 root             Ss    0   8992  4339428      0 /usr/libexec/con /usr/libexec/configd
	 */

	var err error

	timeout := 10 * time.Second
	command := "ps -ax -o pid,ppid,%cpu,%mem,user,state,nice,rss,vsize,pagein,comm,command"
	result, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: command,
		Timeout: timeout,
	})
	if err != nil || result.RC > 0 {
		return nil, fmt.Errorf("Error while executing '%v'", command)
	}

	lines := strings.Split(result.Stdout, "\n")
	if len(lines) > 0 {
		//Remove first line
		lines = lines[1:]
	}

	var processes [][]string
	for _, l := range lines {
		var lr []string
		for _, r := range strings.Split(l, " ") {
			if r == "" {
				continue
			}
			lr = append(lr, strings.TrimSpace(r))
		}
		if len(lr) != 0 {
			processes = append(processes, lr)
		}
	}

	processResults := make([]*resultProcess, 0, len(lines))

	for _, p := range processes {
		//   0     1     2    3    4                5  6      7        8      9   10                 11
		// PID  PPID  %CPU %MEM USER             STAT NI    RSS      VSZ PAGEIN COMM             COMMAND
		pid, _ := strconv.ParseInt(string(p[0]), 10, 32)
		ppid, _ := strconv.ParseInt(string(p[1]), 10, 32)
		cpu, _ := strconv.ParseFloat(string(p[2]), 64)
		mem, _ := strconv.ParseFloat(string(p[3]), 64)
		user := string(p[4])
		state := string(p[5])
		nice, _ := strconv.Atoi(string(p[6]))
		rss, _ := strconv.ParseInt(string(p[7]), 10, 64)
		vsz, _ := strconv.ParseInt(string(p[8]), 10, 64)
		pagein, _ := strconv.ParseInt(string(p[1]), 10, 64)
		bin := string(p[10])
		cmdline := string(p[11])

		s := getStatusName(state[0:1])
		var status []string
		status = append(status, s)

		result := &resultProcess{
			Pid:           int32(pid),
			Ppid:          int32(ppid),
			Username:      user,
			Name:          bin,
			CPUPercent:    cpu,
			MemoryPercent: float32(mem),
			Cmdline:       cmdline,
			Status:        status,
			Exe:           bin,
			Nice:          int32(nice),
			NumFds:        0,
		}
		result.Memory.RSS = uint64(rss)
		result.Memory.VMS = uint64(vsz)
		result.Memory.Swap = uint64(pagein)

		processResults = append(processResults, result)
	}

	return processResults, nil
}

func getStatusName(s string) string {
	switch s {
	case "R":
		return "Running"
	case "S":
		return "Sleep"
	case "T":
		return "Stop"
	case "I":
		return "Idle"
	case "Z":
		return "Zombie"
	case "W":
		return "Wait"
	case "L":
		return "Lock"
	default:
		return ""
	}
}
