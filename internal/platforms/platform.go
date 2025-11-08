package platforms

import (
	"fmt"
	"time"

	"panoptic/internal/config"
)

type Platform interface {
	Initialize(app config.AppConfig) error
	Navigate(url string) error
	Click(selector string) error
	Fill(selector, value string) error
	Submit(selector string) error
	Wait(duration int) error
	Screenshot(filename string) error
	StartRecording(filename string) error
	StopRecording() error
	GetMetrics() map[string]interface{}
	Close() error
}

type PlatformFactory struct{}

func NewPlatformFactory() *PlatformFactory {
	return &PlatformFactory{}
}

func (f *PlatformFactory) CreatePlatform(appType string) (Platform, error) {
	switch appType {
	case "web":
		return NewWebPlatform(), nil
	case "desktop":
		return NewDesktopPlatform(), nil
	case "mobile":
		return NewMobilePlatform(), nil
	default:
		return nil, fmt.Errorf("unsupported platform type: %s", appType)
	}
}

func waitForPageLoad() {
	time.Sleep(2 * time.Second)
}