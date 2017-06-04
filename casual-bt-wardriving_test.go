package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	devices := parse("Scanning ...\n	12:34:56:78:90:42	Device 1\n	13:37:13:37:13:37	//Device $('2 ")

	if len(devices) != 2 {
		t.Errorf("Expected two devices, got %s", len(devices))
	}
	device1 := devices[0]
	if device1.mac != "12:34:56:78:90:42" || device1.name != "Device 1" {
		t.Error("Wrong values for Device 1")
	}
	device2 := devices[1]
	if device2.mac != "13:37:13:37:13:37" || device2.name != "//Device $('2 " {
		t.Error("Wrong values for Device 2")
	}
}
