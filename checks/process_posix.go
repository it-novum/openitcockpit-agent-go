// +build linux darwin

package checks

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
)

type ps struct {
	Pid     int32
	Ppid    int32
	Cpup    float64
	Memp    float32
	User    string
	Stat    []string
	Nice    int32
	Rss     uint64
	VSZ     uint64
	Pagein  uint64
	Command string
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	/*PID  PPID  %CPU %MEM USER             STAT NI    RSS      VSZ PAGEIN COMMAND
	 *  1     0   0,1  0,1 root             Ss    0  21340  4313476      0 /sbin/launchd
	 *109     1   0,0  0,0 root             Ss    0   1136  4306048      0 /usr/sbin/syslogd
	 *110     1   0,0  0,1 root             Ss    0  11456  4336292      0 /usr/libexec/UserEventAgent (System)
	 *112     1   0,0  0,0 root             Ss    0   2252  4296900      0 /System/Library/PrivateFrameworks/Uninstall.framework/Resources/uninstalld
	 *113     1   0,0  0,1 root             Ss    0  20908  4857808      0 /usr/libexec/kextd
	 *114     1   0,0  0,0 root             Ss    0   8288  4324792      0 /System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/FSEvents.framework/Versions/A/Support/fseventsd
	 *116     1   0,0  0,1 root             Ss    0  12508  4336452      0 /System/Library/PrivateFrameworks/MediaRemote.framework/Support/mediaremoted
	 *119     1   0,0  0,1 root             Ss    0  13780  4344220      0 /usr/sbin/systemstats --daemon
	 *120     1   0,0  0,1 root             Ss    0   8992  4339428      0 /usr/libexec/configd
	 */

	var err error

	timeout := 10 * time.Second
	command := "ps -ax -o pid,ppid,%cpu,%mem,user,state,nice,rss,vsize,pagein,command"
	if runtime.GOOS == "linux" {
		command = command + " --columns 10000"
	}

	result, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: command,
		Timeout: timeout,
	})
	if err != nil || result.RC > 0 {
		return nil, fmt.Errorf("Error while executing '%v'", command)
	}

	lines := strings.Split(result.Stdout, "\n")

	if len(lines) > 0 {
		//Remove first line (ps header)
		lines = lines[1:]
	}

	var processes []*ps
	fields := []string{"PID", "PPID", "%CPU", "%MEM", "USER", "STAT", "NI", "RSS", "VSZ", "PAGEIN", "COMMAND"}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		index := 0

		ps := &ps{}

		var piece string
		delimiter := " "
		for i, c := range line {
			char := string(c)

			if fields[index] == "COMMAND" {
				// Command could have spaces so we use the complete string...
				ps.Command = line[i:]
				processes = append(processes, ps)

				// Go to the next line of the ps output
				index = 0
				break
			}

			if char != delimiter {
				piece = piece + char
			} else {
				// We hit a space
				if piece != "" {
					//fmt.Println(fields[index] + ": " + piece)

					if fields[index] == "PID" {
						pid, _ := strconv.ParseInt(piece, 10, 32)
						ps.Pid = int32(pid)
					}

					if fields[index] == "PPID" {
						ppid, _ := strconv.ParseInt(piece, 10, 32)
						ps.Ppid = int32(ppid)
					}

					if fields[index] == "%CPU" {
						cpu, _ := strconv.ParseFloat(piece, 64)
						ps.Cpup = cpu
					}

					if fields[index] == "%MEM" {
						mem, _ := strconv.ParseFloat(piece, 64)
						ps.Memp = float32(mem)
					}

					if fields[index] == "USER" {
						ps.User = piece
					}

					if fields[index] == "STAT" {
						s := getStatusName(piece[0:1])
						var status []string
						status = append(status, s)

						ps.Stat = status
					}

					if fields[index] == "NI" {
						nice, _ := strconv.Atoi(piece)
						ps.Nice = int32(nice)
					}

					if fields[index] == "RSS" {
						rss, _ := strconv.ParseInt(piece, 10, 64)
						ps.Rss = uint64(rss)
					}

					if fields[index] == "VSZ" {
						vsz, _ := strconv.ParseInt(piece, 10, 64)
						ps.VSZ = uint64(vsz)
					}

					if fields[index] == "PAGEIN" {
						pagein, _ := strconv.ParseInt(piece, 10, 64)
						ps.Pagein = uint64(pagein)
					}

					index++
				}
				piece = ""
			}
		}
	}

	processResults := make([]*resultProcess, 0, len(processes))
	for _, process := range processes {
		cmd := strings.Split(process.Command, " ")
		binaryWithPath := cmd[0]
		binary := path.Base(cmd[0])

		result := &resultProcess{
			Pid:           process.Pid,
			Ppid:          process.Ppid,
			Username:      process.User,
			Name:          binary,
			CPUPercent:    process.Cpup,
			MemoryPercent: process.Memp,
			Cmdline:       process.Command,
			Status:        process.Stat,
			Exe:           binaryWithPath,
			Nice:          process.Nice,
			NumFds:        0,
		}
		result.Memory.RSS = process.Rss
		result.Memory.VMS = process.VSZ
		result.Memory.Swap = process.Pagein
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
