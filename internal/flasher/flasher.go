package flasher

import (
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/factoryimage"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/fastboot"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

const (
	UdevRules = "# Google\nSUBSYSTEM==\"usb\", ATTR{idVendor}==\"18d1\", GROUP=\"sudo\"\n# Xiaomi\nSUBSYSTEM==\"usb\", ATTR{idVendor}==\"2717\", GROUP=\"sudo\"\n"
	RulesFile = "98-device-flasher.rules"
	RulesPath = "/etc/udev/rules.d2/"
)

type Config struct {
	HostOS string
	DestinationDirectory string
	FactoryImage *factoryimage.FactoryImage
	PlatformTools *platformtools.PlatformTools
}

type Flasher struct {
	hostOS string
	workingDirectory string
	factoryImage *factoryimage.FactoryImage
	platformtools *platformtools.PlatformTools
}

func New(config *Config) *Flasher {
	return &Flasher{
		workingDirectory: config.DestinationDirectory,
		hostOS: config.HostOS,
		factoryImage: config.FactoryImage,
		platformtools: config.PlatformTools,
	}
}

func (f *Flasher) DiscoverDevices() ([]*Device, error) {
	if f.hostOS == "linux" {
		err := f.checkUdevRules()
		if err != nil {
			return nil, err
		}
	}

	deviceIds, err := f.platformtools.ADB.GetDeviceIds()
	if err != nil {
		return nil, err
	}
	if len(deviceIds) == 0 {
		deviceIds, err = f.platformtools.Fastboot.GetDeviceIds()
		if err != nil {
			return nil, err
		}
		if len(deviceIds) == 0 {
			return nil, fmt.Errorf("no devices detected with adb or fastboot")
		}
	}

	var devices []*Device
	for _, deviceId := range deviceIds {
		deviceCodename, err := f.platformtools.ADB.GetDeviceCodename(deviceId)
		if err != nil || deviceCodename == "" {
			deviceCodename, err = f.platformtools.Fastboot.GetDeviceCodename(deviceId)
			if err != nil || deviceCodename == "" {
				return nil, fmt.Errorf("cannot determine device model")
			}
		}
		devices = append(devices, &Device{
			ID: deviceId,
			Codename: deviceCodename,
		})
	}
	return devices, nil
}

func (f *Flasher) Flash(devices []*Device) error {
	defer f.platformtools.ADB.KillServer()

	fmt.Println("running pre extract validation")
	for _, device := range devices {
		err := f.factoryImage.PreExtractValidation(device.Codename)
		if err != nil {
			return err
		}
	}

	fmt.Println("extracting factory image")
	err := f.factoryImage.Extract()
	if err != nil {
		return err
	}

	path := os.Getenv("PATH")
	newPath := f.platformtools.Path + string(os.PathListSeparator) + path
	if f.hostOS == "windows" {
		newPath = f.platformtools.Path + string(os.PathListSeparator) + path
	}
	fmt.Printf("Setting PATH to %v...\n", newPath)
	err = os.Setenv("PATH", newPath)
	if err != nil {
		return err
	}

	for _, device := range devices {
		fmt.Println("flashing device: ", device.ID)
		err = f.flash(device)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Flasher) flash(device *Device) error {
	fmt.Println("running adb reboot bootloader")
	err := f.platformtools.ADB.RebootBootloader(device.ID)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("checking lock status with fastboot")
	lockStatus, err := f.platformtools.Fastboot.GetLockStatus(device.ID)
	if err != nil {
		return err
	}
	if lockStatus != fastboot.Unlocked {
		fmt.Println("Unlocking device " + device.ID + " bootloader...")
		fmt.Println("Please use the volume and power keys on the device to confirm.")
		err := f.platformtools.Fastboot.FlashingUnlock(device.ID)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		lockStatus, err := f.platformtools.Fastboot.GetLockStatus(device.ID)
		if err != nil {
			return err
		}
		if lockStatus != fastboot.Unlocked {
			return fmt.Errorf("failed to unlock device %v bootloader", device.ID)
		}
	}

	fmt.Printf("Flashing %v on device %v...\n", f.factoryImage.Name, device.ID)
	flashCmd := exec.Command(fmt.Sprintf("./%v", f.factoryImage.FlashAll))
	flashCmd.Dir = f.factoryImage.ExtractDirectory
	flashCmd.Stdout = os.Stdout
	flashCmd.Stderr = os.Stdout
	err = flashCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to flash device %v: %w", device.ID, err)
	}

	fmt.Printf("Locking device %v bootloader...\n", device.ID)
	fmt.Println("Please use the volume and power keys on the device to confirm.")
	err = f.platformtools.Fastboot.FlashingLock(device.ID)
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	lockStatus, err = f.platformtools.Fastboot.GetLockStatus(device.ID)
	if err != nil {
		return err
	}
	if lockStatus != fastboot.Locked {
		return fmt.Errorf("failed to lock device %v bootloader", device.ID)
	}

	fmt.Println("Rebooting " + device.ID + "...")
	err = f.platformtools.Fastboot.Reboot(device.ID)
	if err != nil {
		return err
	}
	fmt.Println("Flashing complete")

	return nil
}

func (f *Flasher) checkUdevRules() error {
	_, err := os.Stat(RulesPath)
	if os.IsNotExist(err) {
		err = exec.Command("sudo", "mkdir", RulesPath).Run()
		if err != nil {
			return err
		}
		_, err = os.Stat(RulesFile)
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(RulesFile, []byte(UdevRules), 0644)
			return err
		}
		err = exec.Command("sudo", "cp", RulesFile, RulesPath).Run()
		if err != nil {
			return err
		}
		_ = exec.Command("sudo", "udevadm", "control", "--reload-rules").Run()
		_ = exec.Command("sudo", "udevadm", "trigger").Run()
	}
	return nil
}