package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/prometheus-community/windows_exporter/perflib"
)

// Ein einzelner Takt stellt hundert Nanosekunden oder ein Zehnmillionstel einer Sekunde dar. Es gibt 10.000 Ticks in einer Millisekunde (siehe TicksPerMillisecond )
// und 10 Millionen Ticks in einer Sekunde.
// https://docs.microsoft.com/de-de/dotnet/api/system.datetime.ticks?view=net-5.0
const WINDOWS_TICKS_PER_SECONDS float64 = 1.0e-7 // = 1 / 10000000

const WINDOWS_TICKS_PER_MILISECONDS float64 = 1.0e-6 // = 1 / 1000000

// The github.com/leoluk/perflib_exporter/collector dependency requires github.com/prometheus/common/log
// which does not exists anymore. We only need the constants, so we copy&paste this instead.
// Source: https://github.com/leoluk/perflib_exporter/blob/da7e746e1ac6876bbb7dac385b6a415fee2a8172/collector/mapper.go
// License: MIT License
// Author: Bismarck Paliz bismarck https://github.com/bismarck
// (c) all contributors lofeoluk/perflib_exporter many thanks!
const (
	PERF_COUNTER_RAWCOUNT_HEX           = 0x00000000
	PERF_COUNTER_LARGE_RAWCOUNT_HEX     = 0x00000100
	PERF_COUNTER_TEXT                   = 0x00000b00
	PERF_COUNTER_RAWCOUNT               = 0x00010000
	PERF_COUNTER_LARGE_RAWCOUNT         = 0x00010100
	PERF_DOUBLE_RAW                     = 0x00012000
	PERF_COUNTER_DELTA                  = 0x00400400
	PERF_COUNTER_LARGE_DELTA            = 0x00400500
	PERF_SAMPLE_COUNTER                 = 0x00410400
	PERF_COUNTER_QUEUELEN_TYPE          = 0x00450400
	PERF_COUNTER_LARGE_QUEUELEN_TYPE    = 0x00450500
	PERF_COUNTER_100NS_QUEUELEN_TYPE    = 0x00550500
	PERF_COUNTER_OBJ_TIME_QUEUELEN_TYPE = 0x00650500
	PERF_COUNTER_COUNTER                = 0x10410400
	PERF_COUNTER_BULK_COUNT             = 0x10410500
	PERF_RAW_FRACTION                   = 0x20020400
	PERF_LARGE_RAW_FRACTION             = 0x20020500
	PERF_COUNTER_TIMER                  = 0x20410500
	PERF_PRECISION_SYSTEM_TIMER         = 0x20470500
	PERF_100NSEC_TIMER                  = 0x20510500
	PERF_PRECISION_100NS_TIMER          = 0x20570500
	PERF_OBJ_TIME_TIMER                 = 0x20610500
	PERF_PRECISION_OBJECT_TIMER         = 0x20670500
	PERF_SAMPLE_FRACTION                = 0x20c20400
	PERF_COUNTER_TIMER_INV              = 0x21410500
	PERF_100NSEC_TIMER_INV              = 0x21510500
	PERF_COUNTER_MULTI_TIMER            = 0x22410500
	PERF_100NSEC_MULTI_TIMER            = 0x22510500
	PERF_COUNTER_MULTI_TIMER_INV        = 0x23410500
	PERF_100NSEC_MULTI_TIMER_INV        = 0x23510500
	PERF_AVERAGE_TIMER                  = 0x30020400
	PERF_ELAPSED_TIME                   = 0x30240500
	PERF_COUNTER_NODATA                 = 0x40000200
	PERF_AVERAGE_BULK                   = 0x40020500
	PERF_SAMPLE_BASE                    = 0x40030401
	PERF_AVERAGE_BASE                   = 0x40030402
	PERF_RAW_BASE                       = 0x40030403
	PERF_PRECISION_TIMESTAMP            = 0x40030500
	PERF_LARGE_RAW_BASE                 = 0x40030503
	PERF_COUNTER_MULTI_BASE             = 0x42030500
	PERF_COUNTER_HISTOGRAM_TYPE         = 0x80000000
)

// Credit to: https://github.com/bosun-monitor/bosun/blob/master/cmd/scollector/collectors/disk_windows.go#L15-L22
// Converts 100ns samples to 0-100 Percent samples
const WIDOWS_TICKS_TO_PERCENT = 100000

// Windows timestamps starts at 1601-01-01T00:00:00Z which is 11644473600 before unix timestamps starts 1970-01-01T00:00:00Z.
// https://stackoverflow.com/a/6161842
// diff from Windows epoch to unix epoch in nanoseconds:
const WINDOWS_TO_UNIX_EPOCH = 11644473600 * 1000 * 1000

// Source: https://github.com/prometheus-community/windows_exporter/blob/b5284aca85433c097fdbd64671b4c6dcaff10037/pkg/perflib/unmarshal.go#L18-L111
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

			secondValue := false

			st := strings.Split(tag, ",")
			tag = st[0]

			for _, t := range st {
				if t == "secondvalue" {
					secondValue = true
				}
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

			if secondValue {
				if !ctr.Def.HasSecondValue {
					return fmt.Errorf("tagged field %v expected a SecondValue, which was not present", f.Name)
				}
				target.Field(i).SetFloat(float64(ctr.SecondValue))
				continue
			}

			switch ctr.Def.CounterType {
			case PERF_ELAPSED_TIME:
				target.Field(i).SetFloat(float64(ctr.Value-WINDOWS_TO_UNIX_EPOCH) / float64(object.Frequency))
			case PERF_100NSEC_TIMER, PERF_PRECISION_100NS_TIMER:
				//target.Field(i).SetFloat(float64(ctr.Value) / WINDOWS_TICKS_PER_SECONDS)
				target.Field(i).SetFloat(float64(ctr.Value) * WINDOWS_TICKS_PER_SECONDS)
				//target.Field(i).SetFloat(float64(ctr.Value))

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
	return uint64((float64(windowsTicks)*WINDOWS_TICKS_PER_SECONDS - WINDOWS_TO_UNIX_EPOCH))
}
