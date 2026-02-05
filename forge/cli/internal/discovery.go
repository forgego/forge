package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverApps discovers all apps in a project
func DiscoverApps(projectRoot string) ([]string, error) {
	appDir := filepath.Join(projectRoot, "app")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("app directory not found")
	}

	var apps []string
	err := filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a direct child of app/ directory
		relPath, err := filepath.Rel(appDir, path)
		if err != nil {
			return err
		}

		// If it's a directory and directly under app/, it's an app
		if info.IsDir() && !strings.Contains(relPath, string(filepath.Separator)) && relPath != "." {
			apps = append(apps, relPath)
		}

		return nil
	})

	return apps, err
}

// DiscoverModels discovers model files in apps
func DiscoverModels(projectRoot string) (map[string][]string, error) {
	apps, err := DiscoverApps(projectRoot)
	if err != nil {
		return nil, err
	}

	models := make(map[string][]string)
	for _, app := range apps {
		modelsPath := filepath.Join(projectRoot, "app", app, "models.go")
		if _, err := os.Stat(modelsPath); err == nil {
			models[app] = append(models[app], modelsPath)
		}
	}

	return models, nil
}

// DiscoverAdminFiles discovers admin.go files
func DiscoverAdminFiles(projectRoot string) ([]string, error) {
	apps, err := DiscoverApps(projectRoot)
	if err != nil {
		return nil, err
	}

	var adminFiles []string
	for _, app := range apps {
		adminPath := filepath.Join(projectRoot, "app", app, "admin.go")
		if _, err := os.Stat(adminPath); err == nil {
			adminFiles = append(adminFiles, adminPath)
		}
	}

	return adminFiles, nil
}

// DiscoverAPIFiles discovers api.go files
func DiscoverAPIFiles(projectRoot string) ([]string, error) {
	apps, err := DiscoverApps(projectRoot)
	if err != nil {
		return nil, err
	}

	var apiFiles []string
	for _, app := range apps {
		apiPath := filepath.Join(projectRoot, "app", app, "api.go")
		if _, err := os.Stat(apiPath); err == nil {
			apiFiles = append(apiFiles, apiPath)
		}
	}

	return apiFiles, nil
}
