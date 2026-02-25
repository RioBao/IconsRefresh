//go:build mage
// +build mage

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
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

// Build builds one or more Windows binaries.
//
// Defaults to amd64 only.
// Optional targets:
//   - set ICONSREFRESH_BUILD_ARM64=1 to include arm64
//   - set ICONSREFRESH_BUILD_386=1 to include 386
func Build() error {
	mg.Deps(Clean)
	mg.Deps(Generate)

	resourcePath := "resource.syso"
	hasResource := fileExists(resourcePath)

	for _, arch := range buildArchMatrix() {
		resourceDisabled := false
		if arch == "arm64" && hasResource {
			if err := os.Rename(resourcePath, resourcePath+".bak"); err != nil {
				return err
			}
			resourceDisabled = true
		}
		output := releasePath(arch)
		args := []string{
			"build",
			"-trimpath",
			"-buildvcs=false",
			"-o", output,
			"-v",
			"-ldflags", "-s -w -buildid= -H=windowsgui",
		}

		env := mapsClone(goEnv)
		env["GOARCH"] = arch

		fmt.Printf("⚙️ Go build (%s) -> %s...\n", arch, output)
		if err := sh.RunWith(env, mg.GoCmd(), args...); err != nil {
			if resourceDisabled {
				_ = os.Rename(resourcePath+".bak", resourcePath)
			}
			return err
		}
		if resourceDisabled {
			if err := os.Rename(resourcePath+".bak", resourcePath); err != nil {
				return err
			}
		}
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

// Generate Run go generate
func Generate() error {
	mg.Deps(Download)
	mg.Deps(versionInfo)

	fmt.Println("⚙️ Go generate...")
	if err := sh.RunV(mg.GoCmd(), "generate", "-v"); err != nil {
		return err
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

func buildArchMatrix() []string {
	arches := []string{"amd64"}
	if envTrue("ICONSREFRESH_BUILD_ARM64") {
		arches = append(arches, "arm64")
	}
	if envTrue("ICONSREFRESH_BUILD_386") {
		arches = append(arches, "386")
	}
	return arches
}

func envTrue(key string) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return slices.Contains([]string{"1", "true", "yes", "on"}, value)
}

func releasePath(arch string) string {
	return path.Join(binPath, fmt.Sprintf("IconsRefresh_windows_%s.exe", arch))
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
	s, _ := sh.Output("bash", "-c", "git describe --abbrev=0 --tags 2> /dev/null")
	if s == "" {
		return "0.0.0"
	}
	return s
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

func copyDir(src string, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = os.ReadDir(src); err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}

func copyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
