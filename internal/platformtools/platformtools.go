package platformtools

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io"
	"net/http"
	"os"
)

const (
	DefaultBaseURI = "https://dl.google.com/android/repository"
	PlatformToolsFilenameTemplate = "platform-tools-%v-%v.zip"
)

type PlatformToolsPath string

type Config struct {
	BaseURI string
	HttpClient *http.Client
	OS string
	ToolsVersion string
	DestinationDirectory string
}

type PlatformTools struct {
	httpClient *http.Client
	downloadURI string
	workingDirectory string
	zipFile string
	path PlatformToolsPath
}

func New(config *Config) *PlatformTools {
	platformToolsFilename := fmt.Sprintf(PlatformToolsFilenameTemplate, config.ToolsVersion, config.OS)
	downloadURI := fmt.Sprintf("%v/%v", DefaultBaseURI, platformToolsFilename)
	workingDirectory := config.DestinationDirectory
	zipFile := fmt.Sprintf("%v/%v", workingDirectory, "platform-tools.zip")
	path := fmt.Sprintf("%v/platform-tools", workingDirectory)
	return &PlatformTools{
		httpClient: config.HttpClient,
		downloadURI: downloadURI,
		workingDirectory: workingDirectory,
		zipFile: zipFile,
		path: PlatformToolsPath(path),
	}
}

func (p *PlatformTools) Initialize() (PlatformToolsPath, error) {
	err := p.download()
	if err != nil {
		return "", err
	}

	err = p.unzip()
	if err != nil {
		return "", err
	}

	return p.path, nil
}

func (p *PlatformTools) Cleanup() error {
	return os.RemoveAll(p.workingDirectory)
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

func (p *PlatformTools) unzip() error {
	return archiver.Unarchive(p.zipFile, p.workingDirectory)
}