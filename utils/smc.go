package utils

//SmcSensorNames maps the names from the SMC chip to human readable names
// this list is taken from: https://github.com/shirou/gopsutil/blob/a1e77476b2bb7bbe6130fa1cbc39fe1635ff459c/v3/host/smc_darwin.h#L6-L26
// Also a good source is: https://superuser.com/a/967056
// or maybe https://github.com/dkorunic/iSMC
var SmcSensorNames = map[string]string{
	"TA0P": "Ambient 1",
	"TA1P": "Ambient 2",
	"TC0D": "CPU 0 Diode",
	"TC0H": "CPU 0 Heatsink",
	"TC0P": "CPU 0 Proximity",
	"TB0T": "Battery",
	"TB1T": "Battery 1",
	"TB2T": "Battery 2",
	"TB3T": "Battery",
	"TG0D": "GPU 0 Diode",
	"TG0H": "GPU 0 Heatsink",
	"TG0P": "GPU 0 Proximity",
	"TH0P": "TH0P Bay 1",
	"TM0S": "Memory Module 0",
	"TM0P": "Mainboard Proximity",
	"TN0H": "Northbridge Heatsink",
	"TN0D": "Northbridge Diode",
	"TN0P": "Northbridge Proximity",
	"TI0P": "Thunderbolt 0",
	"TI1P": "Thunderbolt 1",
	"TW0P": "Airport Proximity",
}
