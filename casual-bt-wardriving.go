package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var re = regexp.MustCompile("(?im)^[^0-9a-f]*((?:[0-9a-f]{2}:){5}[0-9a-f]{2})\\s*([^\\s].*)$")

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

func main() {

	// cmdOutput.WriteString("Scanning ...\n	00:26:4A:A3:23:3C	Ares\n	11:26:4A:A3:23:3C	Ares2")
	parse(scan())
}

func scan() string {
	// Create an *exec.Cmd
	cmd := exec.Command("hcitool", "scan", "--flush")

	// Stdout buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput

	err := cmd.Run() // will wait for command to return
	printError(err)
	return cmdOutput.String()
}

func parse(rawScanResult string) {
	addresses := re.FindAllStringSubmatch(rawScanResult, -1)
	for _, addr := range addresses {
		mac := addr[1]
		name := addr[2]

		fmt.Printf("MAC: %s - Name: %s\n", mac, name)
	}
}
