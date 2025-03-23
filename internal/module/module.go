package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"io"
)

func GetModuleEnvVar() (string, string, error) {
	moduleDir := os.Getenv("CODEC_MODULE_DIR")
	systemdPath := os.Getenv("CODEC_SYSTEMD_PATH")

	if moduleDir == ""  {
		return "", "", fmt.Errorf("CODEC_MODULE_DIR is not set")
	}else if systemdPath == "" {
		return "", "", fmt.Errorf("CODEC_SYSTEMD_PATH is not set")
	}

	return moduleDir, systemdPath, nil
}

func LoadModules(
	modulesDir string, 
) ([]CodecModule, error) {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("module system: could not read module directory: %w", err)
	}

	var modules []CodecModule

	for _, entry := range entries {
		if entry.IsDir() {
			module := CodecModule{
				Path: filepath.Join(modulesDir, entry.Name()),
				Name: entry.Name(),
			}
			if !module.Check() {
				continue
			}

			modules = append(modules, module)
		}
	}
	return modules, nil
}

func ProcessModules(
	modules []CodecModule,
	systemdPath string,
) error {
	for _, module := range modules {
		err := module.Process(systemdPath)
		if err != nil {
			return fmt.Errorf("module system: error processing module %q: %w", module.Name, err)
		}
	}
	return nil
}

type CodecModule struct {
	Path string
	Name string
}

func (module *CodecModule) Check() bool {
	files, err := os.ReadDir(module.Path)
	if err != nil {
		fmt.Printf("module system: error reading module directory %q: %v\n", module.Path, err)
		return false
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".service") || file.Name() == "exec.sh" || file.Name() == "daemon.sh" {
			return true
		}
	}

	return false
}

func (module *CodecModule) Process(
	systemdPath string,
) error {
	files, err := os.ReadDir(module.Path)
	if err != nil {
		return fmt.Errorf("module system: error reading module directory %q: %w", module.Path, err)
	}

	var toCopy []string
	exec := false
	daemon := false

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".service") {
			toCopy = append(toCopy, file.Name())
		}else if file.Name() == "exec.sh" {
			exec = true
		}else if file.Name() == "daemon.sh" {
			daemon = true
		}
	}

	if len(toCopy) != 0 {
		err := module.CopyServiceFiles(systemdPath, toCopy)
		if err != nil {
			return fmt.Errorf("module system: error copying service files: %w", err)
		}
	}

	if exec {
		err := module.CreateExecService(systemdPath)
		if err != nil {
			return fmt.Errorf("module system: error creating exec service: %w", err)
		}
	}

	if daemon {
		err := module.CreateDaemonService(systemdPath)
		if err != nil {
			return fmt.Errorf("module system: error creating daemon service: %w", err)
		}
	}

	return nil
}

func (module *CodecModule) CopyServiceFiles(systemdPath string, toCopy []string) error {
	for _, fileName := range toCopy {
		srcPath := filepath.Join(module.Path, fileName)
		destPath := filepath.Join(systemdPath, fileName)

		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("module system: could not open source service file %q: %w", srcPath, err)
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("module system: could not create destination service file %q: %w", destPath, err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("module system: error copying service file %q -> %q: %w", srcPath, destPath, err)
		}
	}

	return nil
}

func (module *CodecModule) CreateExecService(systemdPath string) error {
	serviceFilePath := filepath.Join(systemdPath, module.Name+".service")
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s service
After=network.target

[Service]
Type=oneshot
ExecStart=%s
Restart=always
User=root

[Install]
WantedBy=multi-user.target
`, module.Name, filepath.Join(module.Path, "exec.sh"))

	err := os.WriteFile(serviceFilePath, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("module system: could not write exec service file (%q): %w", serviceFilePath, err)
	}

	return nil
}

func (module *CodecModule) CreateDaemonService(systemdPath string) error {
	serviceFilePath := filepath.Join(systemdPath, module.Name+".service")
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s service
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
User=root

[Install]
WantedBy=multi-user.target
`, module.Name, filepath.Join(module.Path, "daemon.sh"))

	err := os.WriteFile(serviceFilePath, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("module system: could not write daemon service file (%q): %w", serviceFilePath, err)
	}

	return nil
}