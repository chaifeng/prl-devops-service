package install

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/cjlapao/common-go/helper"
)

const (
	MAC_PLIST_DAEMON_PATH = "/Library/LaunchDaemons"
	MAC_PLIST_DAEMON_NAME = "com.parallels.pd-api-service.plist"
)

func InstallService(ctx basecontext.ApiContext) error {
	ctx.LogInfo("Installing service...")
	config := getConfigFromEnv()
	switch os := runtime.GOOS; os {
	case "darwin":
		return installServiceOnMac(ctx, config)
	case "windows":
		return installServiceOnWindows(ctx, config)
	case "linux":
		return installServiceOnLinux(ctx, config)
	default:
		errMsg := fmt.Sprintf("unsupported operating system: %s", os)
		ctx.LogError(errMsg)
		return errors.New(errMsg)
	}
}

func UninstallService(ctx basecontext.ApiContext) error {
	switch os := runtime.GOOS; os {
	case "darwin":
		return uninstallServiceOnMac(ctx)
	case "windows":
		return uninstallServiceOnWindows(ctx)
	case "linux":
		return uninstallServiceOnLinux(ctx)
	default:
		errMsg := fmt.Sprintf("unsupported operating system: %s", os)
		ctx.LogError(errMsg)
		return errors.New(errMsg)
	}
}

func IsInstalled(ctx basecontext.ApiContext) bool {
	return false
}

func installServiceOnMac(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	path, err := getExecutablePath()
	if err != nil {
		return err
	}

	plist, err := generatePlist(path, config)
	if err != nil {
		return err
	}

	if !helper.FileExists(MAC_PLIST_DAEMON_PATH) {
		return errors.New("daemon path does not exist")
	}

	daemonPath := filepath.Join(MAC_PLIST_DAEMON_PATH, MAC_PLIST_DAEMON_NAME)

	// Unload the daemon if it is already loaded
	if helper.FileExists(daemonPath) {
		uninstallServiceOnMac(ctx)
	}

	if err := helper.WriteToFile(plist, daemonPath); err != nil {
		return err
	}

	chownCmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"chown", "root:wheel", daemonPath},
	}
	chmod := helpers.Command{
		Command: "sudo",
		Args:    []string{"chmod", "644", daemonPath},
	}

	launchdLoadCmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"launchctl", "load", daemonPath},
	}

	if _, err := helpers.ExecuteWithNoOutput(chownCmd); err != nil {
		return err
	}
	if _, err := helpers.ExecuteWithNoOutput(chmod); err != nil {
		return err
	}
	if _, err := helpers.ExecuteWithNoOutput(launchdLoadCmd); err != nil {
		return err
	}

	ctx.LogInfo("Service installed successfully")

	return nil
}

func uninstallServiceOnMac(ctx basecontext.ApiContext) error {
	daemonPath := filepath.Join(MAC_PLIST_DAEMON_PATH, MAC_PLIST_DAEMON_NAME)

	cmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"launchctl", "unload", daemonPath},
	}

	if _, err := helpers.ExecuteWithNoOutput(cmd); err != nil {
		return err
	}

	if err := os.Remove(daemonPath); err != nil {
		return err
	}

	ctx.LogInfo("Service uninstalled successfully")
	return nil
}

func installServiceOnWindows(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	return errors.New("not implemented")
}

func uninstallServiceOnWindows(ctx basecontext.ApiContext) error {
	return errors.New("not implemented")
}

func installServiceOnLinux(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	return nil
}

func uninstallServiceOnLinux(ctx basecontext.ApiContext) error {
	return nil
}

func getConfigFromEnv() ApiServiceConfig {
	config := ApiServiceConfig{}
	if os.Getenv(constants.API_PORT_ENV_VAR) != "" {
		config.Port = os.Getenv(constants.API_PORT_ENV_VAR)
	} else {
		config.Port = constants.DEFAULT_API_PORT
	}

	return config
}

func getExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}
