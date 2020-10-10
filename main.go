package main

import (
	"flag"
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/factoryimage"
	"github.com/AOSPAlliance/device-flasher/internal/flasher"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	toolsVersionPtr := flag.String("tools-version", "latest", "platform tools version")
	namePtr := flag.String("name", "CalyxOS", "os name")
	imagePtr := flag.String("image", "", "factory image to flash")
	flag.Parse()

	if *imagePtr == "" {
		log.Fatal("must specify factory image")
	}

	err := run(*namePtr, *imagePtr, *toolsVersionPtr, runtime.GOOS)
	if err != nil {
		log.Fatal(err)
	}
}

func run(name, image, toolsVersion, hostOS string) error {
	platformToolsDir, err := ioutil.TempDir("", "device-flasher-platformtools")
	if err != nil {
		return err
	}
	defer os.RemoveAll(platformToolsDir)

	platformTools := platformtools.New(&platformtools.Config{
		HttpClient: &http.Client{Timeout: time.Second * 60},
		OS: hostOS,
		ToolsVersion: toolsVersion,
		DestinationDirectory: platformToolsDir,
	})
	err = platformTools.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	factoryImageDir, err := ioutil.TempDir("", "device-flasher-factory")
	if err != nil {
		return err
	}
	defer os.RemoveAll(factoryImageDir)

	factoryImage := factoryimage.New(&factoryimage.Config{
		HostOS: hostOS,
		Name: name,
		ImagePath: image,
		WorkingDirectory: factoryImageDir,
	})

	flashTool := flasher.New(&flasher.Config{
		HostOS: "darwin",
		FactoryImage: factoryImage,
		PlatformTools: platformTools,
	})

	fmt.Println("Connect to a wifi network and ensure that no SIM cards are installed")
	fmt.Println("Enable Developer Options on device (Settings -> About Phone -> tap \"Build number\" 7 times)")
	fmt.Println("Enable USB debugging on device (Settings -> System -> Advanced -> Developer Options) and allow the computer to debug (hit \"OK\" on the popup when USB is connected)")
	fmt.Println("Enable OEM Unlocking (in the same Developer Options menu)")
	fmt.Print("When done, press enter to continue")
	_, _ = fmt.Scanln()
	devices, err := flashTool.DiscoverDevices()
	if err != nil {
		return err
	}

	fmt.Println("running flash devices")
	err = flashTool.Flash(devices)
	if err != nil {
		return err
	}
	return nil
}
