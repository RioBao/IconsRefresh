//go:build mage
// +build mage

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default mage target
var Default = Build

var (
	binPath = path.Join("bin")
	goEnv   = map[string]string{
		"GO111MODULE": "on",
		"GOOS":        "windows",
		"CGO_ENABLED": "0",
	}
)

// Build builds Windows binaries for amd64.
func Build() error {
	mg.Deps(Clean)
	mg.Deps(Generate)

	env := mapsClone(goEnv)
	env["GOARCH"] = "amd64"

	cliOut := releasePath()
	cliArgs := []string{
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-o", cliOut,
		"-v",
		"-ldflags", "-s -w -buildid=",
		"./cmd/iconsrefresh",
	}
	fmt.Printf("⚙️ Go build CLI (amd64) -> %s...\n", cliOut)
	if err := sh.RunWith(env, mg.GoCmd(), cliArgs...); err != nil {
		return err
	}

	trayOut := releaseTrayPath()
	trayArgs := []string{
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-o", trayOut,
		"-v",
		"-ldflags", "-s -w -buildid= -H=windowsgui",
		"./cmd/iconsrefresh-tray",
	}
	fmt.Printf("⚙️ Go build tray (amd64) -> %s...\n", trayOut)
	if err := sh.RunWith(env, mg.GoCmd(), trayArgs...); err != nil {
		return err
	}

	return nil
}

// Clean Remove files generated at build-time
func Clean() error {
	if err := createDir(binPath); err != nil {
		return err
	}
	if err := cleanDir(binPath); err != nil {
		return err
	}
	return nil
}

// Generate builds versioninfo.json, ui icon assets, and resource.syso for each cmd.
func Generate() error {
	mg.Deps(Download)
	mg.Deps(convertIcon)
	mg.Deps(versionInfo)

	fmt.Println("⚙️ Generating Windows resources...")
	for _, out := range []string{
		"cmd/iconsrefresh/resource.syso",
		"cmd/iconsrefresh-tray/resource.syso",
	} {
		fmt.Printf("  → %s\n", out)
		if err := sh.RunV(mg.GoCmd(), "run",
			"github.com/josephspurrier/goversioninfo/cmd/goversioninfo",
			"-o", out,
		); err != nil {
			return err
		}
	}
	return nil
}

// Download Run go mod download
func Download() error {
	fmt.Println("⚙️ Go mod download...")
	if err := sh.RunWith(goEnv, mg.GoCmd(), "mod", "download"); err != nil {
		return err
	}

	return nil
}

func releasePath() string {
	return path.Join(binPath, "IconsRefresh.exe")
}

func releaseTrayPath() string {
	return path.Join(binPath, "IconsRefreshUI.exe")
}

func mapsClone(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// mod returns module name
func mod() string {
	f, err := os.Open("go.mod")
	if err == nil {
		reader := bufio.NewReader(f)
		line, _, _ := reader.ReadLine()
		return strings.Replace(string(line), "module ", "", 1)
	}
	return ""
}

// version returns app version based on git tag
func version() string {
	return strings.TrimLeft(tag(), "v")
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, err := sh.Output("git", "describe", "--abbrev=0", "--tags")
	if err != nil || strings.TrimSpace(s) == "" {
		return "0.0.0"
	}
	return strings.TrimSpace(s)
}

// hash returns the git hash for the current repo or "" if none.
func hash() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}

// versionInfo generates versioninfo.json
func versionInfo() error {
	fmt.Println("🔨 Generating versioninfo.json...")

	var tpl = template.Must(template.New("").Parse(`{
	"FixedFileInfo":
	{
		"FileFlagsMask": "3f",
		"FileFlags ": "00",
		"FileOS": "040004",
		"FileType": "01",
		"FileSubType": "00"
	},
	"StringFileInfo":
	{
		"Comments": "",
		"CompanyName": "",
		"FileDescription": "Refresh icons on Desktop, Start Menu and Taskbar",
		"FileVersion": "{{ .Version }}.0",
		"InternalName": "",
		"LegalCopyright": "https://{{ .Package }}",
		"LegalTrademarks": "",
		"OriginalFilename": "IconsRefresh.exe",
		"PrivateBuild": "",
		"ProductName": "IconsRefresh",
		"ProductVersion": "{{ .Version }}.0",
		"SpecialBuild": ""
	},
	"VarFileInfo":
	{
		"Translation": {
			"LangID": "0409",
			"CharsetID": "04B0"
		}
	}
}`))

	f, err := os.Create("versioninfo.json")
	if err != nil {
		return err
	}
	defer f.Close()

	return tpl.Execute(f, struct {
		Package string
		Version string
	}{
		Package: mod(),
		Version: version(),
	})
}

// convertIcon generates the tray UI icon asset from icon.png.
// Resource icons are intentionally omitted so Windows uses the default EXE icon.
func convertIcon() error {
	if !fileExists("icon.png") {
		if fileExists("cmd/iconsrefresh-tray/icon.png") {
			fmt.Println("warning: icon.png not found, reusing existing cmd/iconsrefresh-tray/icon.png")
			return nil
		}
		return fmt.Errorf("icon.png not found and cmd/iconsrefresh-tray/icon.png is missing")
	}
	return sh.RunV(mg.GoCmd(), "run", "./cmd/genicon/main.go")
}

func createDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o777)
	}
	return nil
}

func cleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
