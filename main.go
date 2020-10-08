package main

import (
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/adb"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/fastboot"
	"github.com/AOSPAlliance/device-flasher/internal/flasher"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools"
	"log"
	"net/http"
	"time"
)

func main() {
	platformTools := platformtools.New(&platformtools.Config{
		HttpClient: &http.Client{Timeout: time.Second * 60},
		OS: "darwin",
		ToolsVersion: "latest",
		DestinationDirectory: "/tmp/testing",
	})
	defer platformTools.Cleanup()
	platformToolsPath, err := platformTools.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	flashTool := flasher.New(&flasher.Config{
		ImagePath: "/tmp/test.zip",
		DestinationDirectory: "/tmp/testing",
		ADB: adb.New(platformToolsPath),
		Fastboot: fastboot.New(platformToolsPath),
	})
	err = flashTool.Flash()
	if err != nil {
		log.Println(err)
	}
}
