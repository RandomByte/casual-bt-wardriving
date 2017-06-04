package main

import (
	"bytes"
	"fmt"
	"github.com/peterbourgon/diskv"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type device struct {
	mac, name string
}

var re = regexp.MustCompile("(?im)^[^0-9a-f]*((?:[0-9a-f]{2}:){5}[0-9a-f]{2})\\s*([^\\s].*)$")

var dv *diskv.Diskv

func main() {
	setupPersistence()

	result := scan()
	parsed := parse(result)
	for _, device := range parsed {
		if checkNew(device) {
			fmt.Println("New")
		} else {
			fmt.Println("Known")
			persist(device)
		}
	}
}

func scan() string {
	// Create an *exec.Cmd
	cmd := exec.Command("hcitool", "scan", "--flush")

	// Stdout buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput

	err := cmd.Run() // will wait for command to return
	if err != nil {
		panic(err)
	}
	return cmdOutput.String()
}

func parse(rawScanResult string) []device {
	devices := re.FindAllStringSubmatch(rawScanResult, -1)
	result := make([]device, len(devices))
	for i, device := range devices {
		result[i].mac = device[1]
		result[i].name = device[2]

	}

	return result
}

func checkNew(device device) bool {
	value, err := dv.Read(device.mac)
	if err != nil {
		return false
	}
	return value != nil
}

func setupPersistence() {
	// Simplest transform function: put all the data files into the base dir.
	flatTransform := func(s string) []string { return []string{} }

	// Initialize a new diskv store, rooted at "diskv-data", with a 1MB cache.
	dv = diskv.New(diskv.Options{
		BasePath:     "diskv-data",
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})
}

func persist(device device) {
	err := dv.Write(device.mac, []byte(device.name))
	if err != nil {
		panic(err)
	}
}
