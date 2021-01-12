package utils

import (
	"fmt"
	"reflect"

	perflibMapper "github.com/leoluk/perflib_exporter/collector"
	"github.com/leoluk/perflib_exporter/perflib"
)

// Ein einzelner Takt stellt hundert Nanosekunden oder ein Zehnmillionstel einer Sekunde dar. Es gibt 10.000 Ticks in einer Millisekunde (siehe TicksPerMillisecond )
// und 10 Millionen Ticks in einer Sekunde.
// https://docs.microsoft.com/de-de/dotnet/api/system.datetime.ticks?view=net-5.0
const WINDOWS_TICKS_PER_SECONDS = 10000000

// Windows timestamps starts at 1601-01-01T00:00:00Z which is 11644473600 before unix timestamps starts 1970-01-01T00:00:00Z.
// https://stackoverflow.com/a/6161842
// diff from Windows epoch to unix epoch in nanoseconds:
const WINDOWS_TO_UNIX_EPOCH = 11644473600 * 1000 * 1000

// Source: https://github.com/prometheus-community/windows_exporter/blob/9723aa221885f593ac77019566c1ced9d4d746fd/collector/perflib.go#L32-L99
// License: MIT License
// Author: Calle Pettersson carlpett https://github.com/carlpett
// (c) all contributors of prometheus-community/windows_exporter many thanks!
func UnmarshalObject(object *perflib.PerfObject, dst interface{}) error {
	if object == nil {
		return fmt.Errorf("perflib.PerfObject can not be nil")
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%v is nil or not a pointer to slice", reflect.TypeOf(dst))
	}
	ev := rv.Elem()
	if ev.Kind() != reflect.Slice {
		return fmt.Errorf("%v is not slice", reflect.TypeOf(dst))
	}

	// Ensure sufficient length
	if ev.Cap() < len(object.Instances) {
		nvs := reflect.MakeSlice(ev.Type(), len(object.Instances), len(object.Instances))
		ev.Set(nvs)
	}

	for idx, instance := range object.Instances {
		target := ev.Index(idx)
		rt := target.Type()

		counters := make(map[string]*perflib.PerfCounter, len(instance.Counters))
		for _, ctr := range instance.Counters {
			if ctr.Def.IsBaseValue && !ctr.Def.IsNanosecondCounter {
				counters[ctr.Def.Name+"_Base"] = ctr
			} else {
				counters[ctr.Def.Name] = ctr
			}
		}

		for i := 0; i < target.NumField(); i++ {
			f := rt.Field(i)
			tag := f.Tag.Get("perflib")
			if tag == "" {
				continue
			}

			ctr, found := counters[tag]
			if !found {
				fmt.Printf("missing counter %q, have %v", tag, counterMapKeys(counters))
				continue
			}
			if !target.Field(i).CanSet() {
				return fmt.Errorf("tagged field %v cannot be written to", f.Name)
			}
			if fieldType := target.Field(i).Type(); fieldType != reflect.TypeOf((*float64)(nil)).Elem() {
				return fmt.Errorf("tagged field %v has wrong type %v, must be float64", f.Name, fieldType)
			}

			switch ctr.Def.CounterType {
			case perflibMapper.PERF_ELAPSED_TIME:
				target.Field(i).SetFloat(float64(ctr.Value-WINDOWS_TO_UNIX_EPOCH) / float64(object.Frequency))
			case perflibMapper.PERF_100NSEC_TIMER, perflibMapper.PERF_PRECISION_100NS_TIMER:
				target.Field(i).SetFloat(float64(ctr.Value) * WINDOWS_TICKS_PER_SECONDS)
			default:
				target.Field(i).SetFloat(float64(ctr.Value))
			}
		}

		if instance.Name != "" && target.FieldByName("Name").CanSet() {
			target.FieldByName("Name").SetString(instance.Name)
		}
	}

	return nil
}

// Source: https://github.com/prometheus-community/windows_exporter/blob/9723aa221885f593ac77019566c1ced9d4d746fd/collector/perflib.go#L101-L107
// License: MIT License
// Author: Calle Pettersson carlpett https://github.com/carlpett
// (c) all contributors of prometheus-community/windows_exporter many thanks!
func counterMapKeys(m map[string]*perflib.PerfCounter) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func WindowsTicksToUnixSeconds(windowsTicks uint64) uint64 {
	return (windowsTicks/WINDOWS_TICKS_PER_SECONDS - WINDOWS_TO_UNIX_EPOCH)
}
