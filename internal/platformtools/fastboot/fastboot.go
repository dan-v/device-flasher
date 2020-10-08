package fastboot

import (
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools"
	"os/exec"
	"strings"
)

type Fastboot struct {
	executable string
}

func New(path platformtools.PlatformToolsPath) *Fastboot {
	return &Fastboot{
		executable: fmt.Sprintf("%v/%v", path, "fastboot"),
	}
}

func (f *Fastboot) Command(args []string) ([]byte, error) {
	cmd := exec.Command(f.executable, args...)
	return cmd.CombinedOutput()
}

func (f *Fastboot) GetDevices() ([]string, error) {
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