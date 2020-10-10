package platformtools

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/adb"
	"github.com/AOSPAlliance/device-flasher/internal/platformtools/fastboot"
	"github.com/mholt/archiver/v3"
	"io"
	"net/http"
	"os"
)

const (
	DefaultBaseURI = "https://dl.google.com/android/repository"
	PlatformToolsFilenameTemplate = "platform-tools-%v-%v.zip"
)

const (
	LinuxSha256   = "f7306a7c66d8149c4430aff270d6ed644c720ea29ef799dc613d3dc537485c6e"
	DarwinSha256  = "ab9dbab873fff677deb2cfd95ea60b9295ebd53b58ec8533e9e1110b2451e540"
	WindowsSha256 = "265dd7b55f58dff1a5ad5073a92f4a5308bd070b72bd8b0d604674add6db8a41"
)

type Config struct {
	BaseURI string
	HttpClient *http.Client
	OS string
	ToolsVersion string
	DestinationDirectory string
}

type PlatformTools struct {
	ADB *adb.ADB
	Fastboot *fastboot.Fastboot
	Path string
	httpClient *http.Client
	os string
	downloadURI string
	sha256 string
	workingDirectory string
	zipFile string
}

func New(config *Config) *PlatformTools {
	platformToolsFilename := fmt.Sprintf(PlatformToolsFilenameTemplate, config.ToolsVersion, config.OS)
	downloadURI := fmt.Sprintf("%v/%v", DefaultBaseURI, platformToolsFilename)
	workingDirectory := config.DestinationDirectory
	zipFile := fmt.Sprintf("%v/%v", workingDirectory, "platform-tools.zip")
	path := fmt.Sprintf("%v/platform-tools", workingDirectory)

	var sha256 string
	switch config.OS {
	case "linux":
		sha256 = LinuxSha256
	case "darwin":
		sha256 = DarwinSha256
	case "windows":
		sha256 = WindowsSha256
	}

	return &PlatformTools{
		Path: path,
		httpClient: config.HttpClient,
		downloadURI: downloadURI,
		sha256: sha256,
		os: config.OS,
		workingDirectory: workingDirectory,
		zipFile: zipFile,
	}
}

func (p *PlatformTools) Initialize() error {
	err := p.download()
	if err != nil {
		return err
	}

	err = p.extract()
	if err != nil {
		return err
	}

	// TODO: add back verify
	//err = p.verify()
	//if err != nil {
	//	return err
	//}

	adbTool, err := adb.New(p.Path, p.os)
	if err != nil {
		return err
	}
	p.ADB = adbTool

	fastbootTool, err := fastboot.New(p.Path, p.os)
	if err != nil {
		return err
	}
	p.Fastboot = fastbootTool

	return nil
}

func (p *PlatformTools) download() error {
	_ = os.Mkdir(p.workingDirectory, os.ModePerm)

	out, err := os.Create(p.zipFile)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := p.httpClient.Get(p.downloadURI)
	if err != nil {

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad download status from %v: %v", p.zipFile, resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p *PlatformTools) extract() error {
	return archiver.Unarchive(p.zipFile, p.workingDirectory)
}

func (p *PlatformTools) verify() error {
	f, err := os.Open(p.zipFile)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if p.sha256 != sum {
		return errors.New("SHA256 mismatch")
	}
	return nil
}