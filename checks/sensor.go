package checks

import (
	"context"
	"fmt"
	"runtime"

	"github.com/distatus/battery"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"
)

// CheckSensor gathers information about system sensors
type CheckSensor struct {
}

// Name will be used in the response as check name
func (c *CheckSensor) Name() string {
	return "sensors"
}

type resultSensor struct {
	Temperatures []*temperatureSensor
	Batteries    []*batterySensor
}

//For Mac systems a list of SMC sensors: https://logi.wiki/index.php/SMC_Sensor_Codes
type temperatureSensor struct {
	Label    string  `json:"label"`   // e.g.: TB2T
	Current  float64 `json:"current"` // e.g.: 31 (value in °C)
	High     float64 `json:"high"`
	Critical float64 `json:"critical"`
}

type batterySensor struct {
	ID           int     `json:"id"`
	Percent      float64 `json:"percent"`
	Secsleft     float64 `json:"secsleft"`
	PowerPlugged bool    `json:"power_plugged"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSensor) Run(ctx context.Context) (interface{}, error) {
	sensors, err := host.SensorsTemperaturesWithContext(ctx)
	sensorResults := make([]*temperatureSensor, 0, len(sensors))

	if err != nil {
		log.Errorln("Check Sonsors: Temperatures: ", err)
	} else {
		for _, sensor := range sensors {
			label := sensor.SensorKey
			if runtime.GOOS == "darwin" {
				if smcName, ok := utils.SmcSensorNames[sensor.SensorKey]; ok {
					label = fmt.Sprintf("%v (%v)", smcName, sensor.SensorKey)
				}
			}

			sensorResult := &temperatureSensor{
				Label:    label,
				Current:  sensor.Temperature,
				High:     sensor.High,
				Critical: sensor.Critical}
			sensorResults = append(sensorResults, sensorResult)
		}
	}

	batteries, err := battery.GetAll()
	batteriesResults := make([]*batterySensor, 0, len(batteries))
	if err != nil {
		log.Errorln("Check Sensors: Batteries: ", err)
	} else {
		for i, battery := range batteries {
			batResult := &batterySensor{
				ID:           i,
				Percent:      battery.Current / battery.Full * 100,
				PowerPlugged: (battery.State.String() == "Full" || battery.State.String() == "Charging"),
			}
			batteriesResults = append(batteriesResults, batResult)
		}
	}

	result := &resultSensor{
		Temperatures: sensorResults,
		Batteries:    batteriesResults,
	}
	return result, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckSensor) Configure(config *config.Configuration) (bool, error) {
	return config.Sensors, nil
}
