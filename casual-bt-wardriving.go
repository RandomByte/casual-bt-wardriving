package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/peterbourgon/diskv"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

type device struct {
	Name     string
	Count    int
	LastSeen int64
}

var re = regexp.MustCompile("(?im)^[^0-9a-f]*((?:[0-9a-f]{2}:){5}[0-9a-f]{2})\\s*([^\\s].*)$")

var dv *diskv.Diskv

func main() {
	setupPersistence()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer signal.Stop(sig)

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			loop()
		case s := <-sig:
			fmt.Println("Got signal:", s)
			fmt.Println("Quitting...")
			return
		}
	}
}

func loop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()
	result := scan()
	parsed := parse(result)
	for mac, device := range parsed {
		knownDevice := readDevice(mac)
		if knownDevice == nil {
			handleNewDevice(mac, device)
		} else {
			handleKnownDevice(mac, device, *knownDevice)
		}
	}
}

func handleNewDevice(mac string, device device) {
	fmt.Printf("New device %s: %s\n", device.Name, mac)
	persist(mac, device)
}

func handleKnownDevice(mac string, device device, knownDevice device) {
	if time.Since(time.Unix(knownDevice.LastSeen, 0)).Hours() < 5 {
		// Last seen less then five hours ago
		return
	}

	if device.Name != knownDevice.Name {
		fmt.Printf("Same MAC but different name: %s (new) vs. %s (known)\n", device.Name, knownDevice.Name)

		err := dv.Write("nameclash"+mac+string(time.Now().Unix()), []byte(fmt.Sprintf("%s, %s (new) vs. %s (known)", mac, device.Name, knownDevice.Name)))
		if err != nil {
			fmt.Println(err)
		}
	}

	device.Count = knownDevice.Count + 1
	fmt.Printf("%vx Known device %s: %s\n", device.Count, device.Name, mac)
	persist(mac, device)
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

func parse(rawScanResult string) map[string]device {
	matches := re.FindAllStringSubmatch(rawScanResult, -1)
	devices := make(map[string]device)
	for _, match := range matches {
		devices[match[1]] = device{Name: match[2], LastSeen: time.Now().Unix()}
	}

	return devices
}

func readDevice(mac string) *device {
	value, err := dv.Read(mac)
	if err != nil {
		return nil
	}

	res := &device{}
	json.Unmarshal([]byte(value), res)

	return res
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

func persist(mac string, device device) {
	serialized, _ := json.Marshal(device)
	err := dv.Write(mac, []byte(serialized))
	if err != nil {
		panic(err)
	}
}
