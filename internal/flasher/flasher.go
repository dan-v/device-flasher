package flasher

import (
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/adb"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/fastboot"
	"github.com/mholt/archiver/v3"
	"os"
	"strconv"
	"strings"
)

type Flasher struct {
	imagePath string
	workingDirectory string
	devices []string
	adb *adb.ADB
	fastboot *fastboot.Fastboot
}

type Config struct {
	ImagePath string
	DestinationDirectory string
	ADB *adb.ADB
	Fastboot *fastboot.Fastboot
}

func New(config *Config) *Flasher {
	return &Flasher{
		imagePath: config.ImagePath,
		adb: config.ADB,
		fastboot: config.Fastboot,
	}
}

func (f *Flasher) Flash() error {
	fmt.Println("Connect to a wifi network and ensure that no SIM cards are installed")
	fmt.Println("Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)")
	fmt.Println("Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)")
	fmt.Println("Enable OEM Unlocking (in the same Developer Options menu)")
	fmt.Print("When done, press enter to continue")
	_, _ = fmt.Scanln()
	devices, err := f.adb.GetDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		devices, err = f.fastboot.GetDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return fmt.Errorf("No devices detected with adb or fastboot")
		}
	}
	fmt.Println("Detected " + strconv.Itoa(len(devices)) + " devices: " + strings.Join(devices, ", "))
	deviceCodename, err := f.adb.GetDeviceCodename(devices[0])
	if err != nil {
		return err
	}
	fmt.Println("deviceCodename: ", deviceCodename)
	err = f.validateImage(deviceCodename)
	if err != nil {
		return err
	}
	fmt.Println("unzipping to: ", f.workingDirectory)
	err = archiver.Unarchive(f.imagePath, f.workingDirectory)
	if err != nil {
		return err
	}
	return nil
}

func (f *Flasher) validateImage(deviceCodename string) error {
	if _, err := os.Stat(f.imagePath); os.IsNotExist(err) {
		return err
	}
	if ! strings.Contains(f.imagePath, strings.ToLower(deviceCodename)) {
		return fmt.Errorf("image filename should contain device codename %v", deviceCodename)
	}
	if ! strings.HasSuffix(f.imagePath, ".zip") {
		return fmt.Errorf("image filename should end in .zip")
	}
	if ! strings.Contains(f.imagePath, "factory") {
		return fmt.Errorf("image filename should contain factory")
	}
	return nil
}