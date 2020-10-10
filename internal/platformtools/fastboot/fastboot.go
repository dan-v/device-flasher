package fastboot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	fastbootExecutable = "fastboot"
)

type FastbootLockStatus int

const (
	Unknown FastbootLockStatus = iota
	Unlocked
	Locked
)

type Fastboot struct {
	executable string
}

func New(path, osName string) (*Fastboot, error) {
	executable := filepath.Join(path, fastbootExecutable)
	if osName == "windows" {
		executable = executable + ".exe"
	}
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		return nil, err
	}
	return &Fastboot{
		executable: fmt.Sprintf("%v/%v", path, fastbootExecutable),
	}, nil
}

func (f *Fastboot) Command(args []string) ([]byte, error) {
	cmd := exec.Command(f.executable, args...)
	return cmd.CombinedOutput()
}

func (f *Fastboot) GetDeviceCodename(deviceId string) (string, error) {
	return f.GetVar("product", deviceId)
}

func (f *Fastboot) GetDeviceIds() ([]string, error) {
	resp, err := f.Command([]string{"devices"})
	if err != nil {
		return nil, err
	}
	devices := strings.Split(string(resp), "\n")
	devices = devices[:len(devices)-1]
	for i, device := range devices {
		devices[i] = strings.Split(device, "\t")[0]
	}
	return devices, nil
}

func (f *Fastboot) FlashingUnlock(deviceId string) error {
	_, err := f.Command([]string{"-s", deviceId, "flashing", "unlock"})
	if err != nil {
		return err
	}
	return nil
}

func (f *Fastboot) FlashingLock(deviceId string) error {
	_, err := f.Command([]string{"-s", deviceId, "flashing", "lock"})
	if err != nil {
		return err
	}
	return nil
}

func (f *Fastboot) Reboot(deviceId string) error {
	_, err := f.Command([]string{"-s", deviceId, "reboot"})
	if err != nil {
		return err
	}
	return nil
}

func (a *Fastboot) GetLockStatus(deviceId string) (FastbootLockStatus, error) {
	unlocked, err := a.GetVar("unlocked", deviceId)
	if err != nil {
		return Unknown, err
	}
	switch unlocked {
	case "yes":
		return Unlocked, nil
	case "no":
		return Locked, nil
	}
	return Unknown, fmt.Errorf("unknown fastboot unlocked value returned: %v", unlocked)
}

func (a *Fastboot) GetVar(prop, deviceId string) (string, error) {
	resp, err := a.Command([]string{"-s", deviceId, "getvar", prop})
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