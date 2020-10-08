package adb

import (
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools"
	"os/exec"
	"strings"
)

type ADB struct {
	executable string
}

func New(path platformtools.PlatformToolsPath) *ADB {
	return &ADB{
		executable: fmt.Sprintf("%v/%v", path, "adb"),
	}
}

func (a *ADB) Command(args []string) ([]byte, error) {
	cmd := exec.Command(a.executable, args...)
	return cmd.CombinedOutput()
}

func (a *ADB) GetDevices() ([]string, error) {
	resp, err := a.Command([]string{"devices"})
	if err != nil {
		return nil, err
	}
	devices := strings.Split(string(resp), "\n")
	devices = devices[1 : len(devices)-2]
	for i, device := range devices {
		devices[i] = strings.Split(device, "\t")[0]
	}
	return devices, nil
}

func (a *ADB) GetDeviceCodename(device string) (string, error) {
	deviceCodename, err := a.GetProp("ro.product.device", device)
	if err != nil {
		return "", err
	}
	if deviceCodename == "" {
		deviceCodename, err = a.GetVar("product", device)
		if err != nil {
			return "", err
		}
		if deviceCodename == "" {
			return "", fmt.Errorf("unable to determine device code name for device %v", device)
		}
	}
	return deviceCodename, nil
}

func (a *ADB) GetProp(prop, device string) (string, error) {
	resp, err := a.Command([]string{"-s", device, "shell", "getprop", prop})
	if err != nil {
		return "", err
	}
	return strings.Trim(string(resp), "[]\n\r"), nil
}

func (a *ADB) GetVar(prop, device string) (string, error) {
	resp, err := a.Command([]string{"-s", device, "shell", "getvar", prop})
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(resp), "\n")
	for _, line := range lines {
		if strings.Contains(line, prop) {
			return strings.Trim(strings.Split(line, " ")[1], "\r"), nil
		}
	}
	return "", fmt.Errorf("var %v not found", prop)
}