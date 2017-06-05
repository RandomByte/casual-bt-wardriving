package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	devices := parse("Scanning ...\n	12:34:56:78:90:42	Device 1\n	13:37:13:37:13:37	//Device $('2 ")

	if len(devices) != 2 {
		t.Errorf("Expected two devices, got %v", len(devices))
	}
	device1, ok := devices["12:34:56:78:90:42"]
	if ok != true || device1.Name != "Device 1" {
		t.Error("Wrong values for Device 1")
	}
	device2, ok := devices["13:37:13:37:13:37"]
	if ok != true || device2.Name != "//Device $('2 " {
		t.Error("Wrong values for Device 2")
	}

	devices = parse("    12:34:56:78:90:42      ")
	if len(devices) != 1 {
		t.Errorf("Expected one device, got %v", len(devices))
	}
	device3, ok := devices["12:34:56:78:90:42"]
	if ok != true || device3.Name != "12:34:56:78:90:42" { // name should fall back to MAC
		t.Error("Wrong values for Device 3")
	}
}

func TestReadDevice(t *testing.T) {
	setupPersistence()
	mac1 := "12:34:56:78:90:42"
	device1 := device{Name: "Device 1"}
	persist(mac1, device1)
	defer dv.Erase(mac1)

	if readDevice(mac1) == nil {
		t.Error("Device 1 should be known, but is new")
	}

	mac2 := "13:37:13:37:13:37"
	if readDevice(mac2) != nil {
		t.Error("Device 2 should be new, but is known")
	}
}

func TestPersist(t *testing.T) {
	setupPersistence()
	mac := "12:34:56:78:90:42"
	device := device{Name: "Device 1"}
	persist(mac, device)
	defer dv.Erase(mac)

	result := readDevice(mac)
	if result == nil {
		t.Error("Device not found in persistence")
	}

	if result.Name != "Device 1" {
		t.Errorf("Device name should be Device 1, but is %s", result.Name)
	}
}

func TestHandleKnownDevice(t *testing.T) {
	mac := "12:34:56:78:90:42"
	device1 := device{Name: "Device 1", LastSeen: time.Now().Unix() - (5 * 60 * 60) + 1}
	handleKnownDevice(mac, device1, device1)
	defer dv.Erase(mac)

	if readDevice(mac) != nil {
		t.Error("Device 1 should not have been handled")
	}

	device2 := device{Name: "Device 2", LastSeen: time.Now().Unix() - (5 * 60 * 60) - 1, Count: 5}

	handleKnownDevice(mac, device2, device2)
	result := readDevice(mac)
	if result == nil {
		t.Error("Device 2 should have been handled")
	}

	if result.Count != 6 {
		t.Errorf("Device 2 should have an increased count of 6, but is %s", result.Count)
	}

	// device3 := device{Name: "Device 3", LastSeen: time.Now().Unix() - 5*60*59}
	// device4 := device{Name: "Device 4", LastSeen: time.Now().Unix() - 5*60*59}
	// TODO check nameclash handling
}

func TestWriteOled(t *testing.T) {
	displayBuffer = make([]string, 8) // initialize
	device1 := device{Name: "Device 1", LastSeen: time.Now().Unix() - (5 * 60 * 60) - 1, Count: 8}
	writeOled(device1)

	if displayBuffer[1] != "" {
		t.Errorf("Expected second line to be empty, but got %s", displayBuffer[1])
	}
	if displayBuffer[0] != "Device 1 (8x)        " {
		t.Errorf("Expected first line to be Device 1 (8x)        , but got %s", displayBuffer[0])
	}

	device2 := device{Name: "Device 2", LastSeen: time.Now().Unix() - (5 * 60 * 60) - 1, Count: 18}
	writeOled(device2)

	if displayBuffer[2] != "" {
		t.Errorf("Expected third line to be empty, but got %s", displayBuffer[2])
	}
	if displayBuffer[1] != "Device 1 (8x)        " {
		t.Errorf("Expected second line to be Device 1 (8x)        , but got %s", displayBuffer[1])
	}
	if displayBuffer[0] != "Device 2 (18x)       " {
		t.Errorf("Expected first line to be Device 2 (18x)       , but got %s", displayBuffer[0])
	}
}

func TestSendToEndpoint(t *testing.T) {
	device1 := deviceFlat{Mac: "12:34:56:78:90:42"}
	device1.Name = "Device 1"
	device2 := deviceFlat{Mac: "13:37:13:37:13:37"}
	device1.Name = "//Device $('2 "

	devices := []deviceFlat{device1, device2}
	done := make(chan error)

	handlePostRequest := func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var requestDevices []deviceFlat
		err := decoder.Decode(&requestDevices)
		if err != nil {
			panic(err)
		}
		if reflect.DeepEqual(requestDevices, devices) != true {
			t.Error("Requested devices don't match original list")
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(handlePostRequest))
	defer ts.Close()
	go sendToEndpoint(ts.URL, devices, done)
	<-done
}

func TestCollectEntries(t *testing.T) {
	setupPersistence()
	mac1 := "12:34:56:78:90:42"
	device1 := device{Name: "Device 1"}
	persist(mac1, device1)
	defer dv.Erase(mac1)

	mac2 := "13:37:13:37:13:37"
	device2 := device{Name: "//Device $('2 "}
	persist(mac2, device2)
	defer dv.Erase(mac2)

	devices := make(chan deviceFlat, 2)
	go collectEntries(devices)

	result1 := <-devices
	if result1.Mac != mac1 || result1.Name != device1.Name {
		t.Error("Wrong values for Device 1")
	}

	result2 := <-devices
	if result2.Mac != mac2 || result2.Name != device2.Name {
		t.Error("Wrong values for Device 2")
	}
}
