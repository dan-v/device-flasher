package factoryimage

import (
	"fmt"
	"github.com/mholt/archiver/v3"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	HostOS string
	Name string
	ImagePath string
	WorkingDirectory string
}

type FactoryImage struct {
	ExtractDirectory string
	Name string
	FlashAll string
	imagePath string
	workingDirectory string
}

func New(config *Config) *FactoryImage {
	flashAll := "flash-all.sh"
	if config.HostOS == "windows" {
		flashAll = "flash-all.bat"
	}

	return &FactoryImage{
		Name: config.Name,
		FlashAll: flashAll,
		workingDirectory: config.WorkingDirectory,
		imagePath: config.ImagePath,
	}
}

func (f *FactoryImage) Extract() error {
	fmt.Printf("extracting image %v to %v\n", f.imagePath, f.workingDirectory)
	err := archiver.Unarchive(f.imagePath, f.workingDirectory)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(f.workingDirectory)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			_, err := os.Stat(f.workingDirectory + file.Name() + string(os.PathSeparator) + f.FlashAll)
			if err != nil {
				f.ExtractDirectory = f.workingDirectory + string(os.PathSeparator) + file.Name()
			}
		}
	}
	if f.ExtractDirectory == "" {
		return fmt.Errorf("unable to find %v in directory %v", f.FlashAll, f.workingDirectory)
	}
	fmt.Printf("FlashAll=%v ExtractDirectory%v\n", f.FlashAll, f.ExtractDirectory)

	return nil
}

func (f *FactoryImage) PreExtractValidation(deviceCodename string) error {
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