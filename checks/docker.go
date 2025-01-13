package checks

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckDocker gathers information about Docker containers
type CheckDocker struct {
}

// Name will be used in the response as check name
func (c *CheckDocker) Name() string {
	return "docker"
}

type resultDocker struct {
	Id               string  `json:"id"`                // First 10 chars of the container id
	Name             string  `json:"name"`              // First name of the docker container
	Image            string  `json:"image"`             // Name of the used Docker image
	SizeRw           int64   `json:"size_rw"`           // Modified or created files (delta/diff to the docker image) in bytes
	SizeRootFs       int64   `json:"size_root_fs"`      // Total size of the containers file system in bytes
	State            string  `json:"state"`             // created, restarting, running, removing, paused, exited, dead // https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	Status           string  `json:"status"`            // Up 41 minutes
	NetworkRx        float64 `json:"network_rx"`        // received data in bytes
	NetworkTx        float64 `json:"network_tx"`        // sent data in bytes
	CpuPercentage    float64 `json:"cpu_percentage"`    // CPU usage of the container as percentage
	MemoryPercentage float64 `json:"memory_percentage"` // Memory usage of the container as percentage
	DiskRead         uint64  `json:"disk_read"`         // Linux and macOS only - Size of read bytes
	DiskWrite        uint64  `json:"disk_write"`        // Linux and macOS only - Size of written bytes
	MemoryUsed       float64 `json:"memory_used"`       // Used memory of the container in bytes (Windows, Linux and macOs)
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDocker) Run(ctx context.Context) (interface{}, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: false, //Only show running containers
	})
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	resultChan := make(chan *resultDocker, len(containers))
	errorChan := make(chan error, len(containers))

	dockerResults := make([]*resultDocker, 0)
	for _, container := range containers {

		wg.Add(1)
		go func(container types.Container) {
			// This is a new "thread" / go func for each docker container
			defer wg.Done()

			response, err := cli.ContainerStats(ctx, container.ID, false)
			if err != nil {
				errorChan <- err
				return
			}

			defer response.Body.Close()
			responseJson := json.NewDecoder(response.Body)

			var stats *types.StatsJSON
			if err := responseJson.Decode(&stats); err != nil {
				errorChan <- err
				return
			}

			var bytesRead, bytesWrite uint64
			var memory, memoryPercentage, cpuPercentage, networkRx, networkTx float64

			// Memory and CPU usage needs to be calculated manually
			// https://github.com/docker/cli/blob/a4a07c643042f4e2a75bf872f38b134502848214/cli/command/container/stats_helpers.go#L79-L128
			if response.OSType != "windows" {
				// Docker daemon is running in Linux or macOS
				previousCPU := float64(stats.PreCPUStats.CPUUsage.TotalUsage)
				previousSystem := float64(stats.PreCPUStats.SystemUsage)
				cpuPercentage = c.calcCpuPercentageUnix(previousCPU, previousSystem, stats)
				bytesRead, bytesWrite = c.calcDiskIoUnix(stats.BlkioStats)
				networkRx, networkTx = c.calcNetworkIo(stats.Networks)
				memory = c.calcMemoryUsageUnix(stats.MemoryStats)
				memoryLimit := float64(stats.MemoryStats.Limit)
				memoryPercentage = c.calcMemoryPercentageUnix(memoryLimit, memory)
			} else {
				// Docker daemon is running in Windows
				cpuPercentage = c.calcCpuPercentageWindows(stats)
				bytesRead = stats.StorageStats.ReadSizeBytes
				bytesWrite = stats.StorageStats.WriteSizeBytes
				networkRx, networkTx = c.calcNetworkIo(stats.Networks)
				memory = float64(stats.MemoryStats.PrivateWorkingSet)
			}

			containerResult := &resultDocker{
				Id:               container.ID[:10],
				Name:             container.Names[0],
				Image:            container.Image,
				SizeRw:           container.SizeRw,
				SizeRootFs:       container.SizeRootFs,
				State:            container.State,  // running
				Status:           container.Status, // Up 6 minutes
				CpuPercentage:    cpuPercentage,
				MemoryPercentage: memoryPercentage,
				NetworkRx:        networkRx,
				NetworkTx:        networkTx,
				DiskRead:         bytesRead,
				DiskWrite:        bytesWrite,
				MemoryUsed:       memory,
			}

			// Return error back to check thread
			resultChan <- containerResult
		}(container)
	}

	// Check thread Wait for all results to come back and close the channels
	go func() {
		wg.Wait()
		close(errorChan)
		close(resultChan)
	}()

	for err := range errorChan {
		log.Errorln("Docker Check error: ", err)
	}

	for result := range resultChan {
		dockerResults = append(dockerResults, result)
	}

	return dockerResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckDocker) Configure(config *config.Configuration) (bool, error) {
	return config.Docker, nil
}

func (c *CheckDocker) calcCpuPercentageUnix(previousCPU, previousSystem float64, stats *types.StatsJSON) float64 {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L166-L183
	var cpuPercentage float64 = 0.0
	var cpuDelta float64 = float64(stats.CPUStats.CPUUsage.TotalUsage) - previousCPU
	var systemDelta float64 = float64(stats.CPUStats.SystemUsage) - previousSystem
	onlineCpus := stats.CPUStats.OnlineCPUs

	if onlineCpus == 0 {
		onlineCpus = uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		//invalid operation: mismatched types float64 and uint32compiler
		cpuPercentage = (cpuDelta / systemDelta) * float64(onlineCpus) * 100.0
	}

	return cpuPercentage
}

func (c *CheckDocker) calcCpuPercentageWindows(stats *types.StatsJSON) float64 {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L185-L199
	possibleIntervals := uint64(stats.Read.Sub(stats.PreRead).Nanoseconds())
	possibleIntervals = possibleIntervals / 100
	possibleIntervals = possibleIntervals * uint64(stats.NumProcs)

	// Used intervals
	usedIntervals := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	var cpuPercentage float64 = 0.0
	if possibleIntervals > 0 {
		cpuPercentage = float64(usedIntervals) / float64(possibleIntervals) * 100.0
	}

	return cpuPercentage
}

func (c *CheckDocker) calcDiskIoUnix(diskio types.BlkioStats) (uint64, uint64) {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L201-L215

	var bytesRead, bytesWrite uint64
	for _, blkioStatEntry := range diskio.IoServiceBytesRecursive {
		if len(blkioStatEntry.Op) == 0 {
			continue
		}
		switch blkioStatEntry.Op[0] {
		case 'r', 'R':
			bytesRead = bytesRead + blkioStatEntry.Value
		case 'w', 'W':
			bytesWrite = bytesWrite + blkioStatEntry.Value
		}
	}

	return bytesRead, bytesWrite
}

func (c *CheckDocker) calcNetworkIo(network map[string]types.NetworkStats) (float64, float64) {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L217-L225
	var networkRx, networkTx float64

	for _, stats := range network {
		networkRx = networkRx + float64(stats.RxBytes)
		networkTx = networkTx + float64(stats.TxBytes)
	}
	return networkRx, networkTx
}

func (c *CheckDocker) calcMemoryUsageUnix(memory types.MemoryStats) float64 {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L239-L249

	// cgroup v1
	// https://stackoverflow.com/a/2050629
	if inactiveFile, isCgroupV1 := memory.Stats["total_inactive_file"]; isCgroupV1 && inactiveFile < memory.Usage {
		return float64(memory.Usage - inactiveFile)
	}

	// cgroup v2
	if inactiveFile := memory.Stats["inactive_file"]; inactiveFile < memory.Usage {
		return float64(memory.Usage - inactiveFile)
	}

	return float64(memory.Usage)
}

func (c *CheckDocker) calcMemoryPercentageUnix(limit float64, noCache float64) float64 {
	// Credit to:
	// https://github.com/docker/cli/blob/e31e00585363e6a989e37fa92cf06481ea218344/cli/command/container/stats_helpers.go#L239-L249
	if limit != 0 {
		return noCache / limit * 100.0
	}
	return 0.0
}
