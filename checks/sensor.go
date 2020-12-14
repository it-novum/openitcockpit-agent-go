package checks

import (
	"context"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/v3/host"
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

type temperatureSensor struct {
	Label    string  `json:"label"`
	Current  float64 `json:"current"`
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
func (c *CheckSensor) Run(ctx context.Context) (*CheckResult, error) {
	sensors, err := host.SensorsTemperaturesWithContext(ctx)
	sensorResults := make([]*temperatureSensor, 0, len(sensors))

	if err != nil {
		return nil, err
	}
	for _, sensor := range sensors {
		sensorResult := &temperatureSensor{
			Label:    sensor.SensorKey,
			Current:  sensor.Temperature,
			High:     sensor.High,
			Critical: sensor.Critical}
		sensorResults = append(sensorResults, sensorResult)
	}

	batteries, err := battery.GetAll()
	batteriesResults := make([]*batterySensor, 0, len(batteries))
	if err != nil {
		return nil, err
	}

	for i, battery := range batteries {
		batResult := &batterySensor{
			ID:           i,
			Percent:      battery.Current / battery.Full * 100,
			PowerPlugged: (battery.State.String() == "Full" || battery.State.String() == "Charging"),
		}
		batteriesResults = append(batteriesResults, batResult)
	}

	result := &resultSensor{
		Temperatures: sensorResults,
		Batteries:    batteriesResults,
	}
	return &CheckResult{Result: result}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckSensor) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckSensor) Configure(_ interface{}) error {
	return nil
}
