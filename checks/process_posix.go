// +build linux darwin

package checks

import (
	"context"
	"fmt"
	"regexp"
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
	headerLine := lines[0]

	if len(lines) > 0 {
		//Remove first line
		lines = lines[1:]
	}

	var positions = make(map[string]map[string]int)

	re := regexp.MustCompile(`\S+`)
	submatchall := re.FindAllString(headerLine, -1)
	start := 0
	end := 0
	for i := 0; i < len(submatchall); i++ {
		if i != len(submatchall)-1 {
			//fmt.Println("Ends with : ",i, strings.Index(headerLine, submatchall[i+1])-1 )
			end = strings.Index(headerLine, submatchall[i+1]) - 1
		} else {
			end = 0
		}
		if i > 0 {
			start = strings.Index(headerLine, submatchall[i])
		}

		positions[submatchall[i]] = map[string]int{
			"start": start,
			"end":   end,
		}
		//fmt.Println(submatchall[i], "Started with : ", start, " and ends with ", end) // Position 0
		//if end > 0 {
		//	fmt.Println("part = ", i, string(headerLine[start:end]))
		//} else {
		//	fmt.Println("part = ", i, string(headerLine[start:]))
		//}
	}

	//fmt.Println(positions)

	processResults := make([]*resultProcess, 0, len(lines))

	fields := []string{"PID", "PPID", "%CPU", "%MEM", "USER", "STAT", "NI", "RSS", "VSZ", "PAGEIN", "COMM", "COMMAND"}

	for _, l := range lines {
		if l == "" {
			continue
		}

		result := &resultProcess{}
		for _, field := range fields {
			var value string
			if field != "COMMAND" {
				start := positions[field]["start"]
				end := positions[field]["end"]
				value = strings.TrimSpace(l[start:end])
			} else {
				// Comand ist the last element in the string, so get all remaining characters
				start := positions[field]["start"]
				value = strings.TrimSpace(l[start:])
			}

			if field == "PID" {
				pid, _ := strconv.ParseInt(value, 10, 32)
				result.Pid = int32(pid)
			}

			if field == "PPID" {
				ppid, _ := strconv.ParseInt(value, 10, 32)
				result.Ppid = int32(ppid)
			}

			if field == "%CPU" {
				cpu, _ := strconv.ParseFloat(value, 64)
				result.CPUPercent = cpu
			}

			if field == "%MEM" {
				mem, _ := strconv.ParseFloat(value, 64)
				result.MemoryPercent = float32(mem)
			}

			if field == "USER" {
				result.Username = value
			}

			if field == "STAT" {
				s := getStatusName(value[0:1])
				var status []string
				status = append(status, s)

				result.Status = status
			}

			if field == "NI" {
				nice, _ := strconv.Atoi(value)
				result.Nice = int32(nice)
			}

			if field == "RSS" {
				rss, _ := strconv.ParseInt(value, 10, 64)
				result.Memory.RSS = uint64(rss)
			}

			if field == "VSZ" {
				vsz, _ := strconv.ParseInt(value, 10, 64)
				result.Memory.VMS = uint64(vsz)
			}

			if field == "PAGEIN" {
				pagein, _ := strconv.ParseInt(value, 10, 64)
				result.Memory.Swap = uint64(pagein)
			}

			if field == "COMM" {
				result.Exe = value
			}

			if field == "COMMAND" {
				result.Cmdline = value
			}

		}

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
