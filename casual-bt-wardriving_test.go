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

func TestCheckNew(t *testing.T) {
	setupPersistence()
	device1 := device{mac: "12:34:56:78:90:42", name: "Device 1"}
	persist(device1)
	defer dv.Erase(device1.mac)

	if checkNew(device1) != true {
		t.Error("Device 1 should be known, but is new")
	}

	device2 := device{mac: "13:37:13:37:13:37", name: "//Device $('2 "}
	if checkNew(device2) != false {
		t.Error("Device 2 should be new, but is known")
	}
}

func TestPersist(t *testing.T) {
	setupPersistence()
	device := device{mac: "12:34:56:78:90:42", name: "Device 1"}
	persist(device)
	defer dv.Erase(device.mac)

	if checkNew(device) != true {
		t.Error("Device not found in persistence")
	}
}
