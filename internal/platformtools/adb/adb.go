package adb

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	adbExecutable = "adb"
)

type ADB struct {
	executable string
}

func New(path, osName string) (*ADB, error) {
	executable := filepath.Join(path, adbExecutable)
	if osName == "windows" {
		executable = executable + ".exe"
	}
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		return nil, err
	}
	return &ADB{
		executable: executable,
	}, nil
}

func (a *ADB) Command(args []string) ([]byte, error) {
	cmd := exec.Command(a.executable, args...)
	return cmd.CombinedOutput()
}

func (a *ADB) GetDeviceIds() ([]string, error) {
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

func (a *ADB) GetDeviceCodename(deviceId string) (string, error) {
	return  a.GetProp("ro.product.device", deviceId)
}

func (a *ADB) RebootBootloader(deviceId string) (error) {
	_, err := a.Command([]string{"-s", deviceId, "reboot", "bootloader"})
	if err != nil {
		return err
	}
	return nil
}

func (a *ADB) KillServer() error {
	_, err := a.Command([]string{"kill-server"})
	if err != nil {
		return err
	}
	return nil
}

func (a *ADB) GetProp(prop, deviceId string) (string, error) {
	resp, err := a.Command([]string{"-s", deviceId, "shell", "getprop", prop})
	if err != nil {
		return "", err
	}
	return strings.Trim(string(resp), "[]\n\r"), nil
}