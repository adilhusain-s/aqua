package install_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	"github.com/clivm/clivm/pkg/controller/install"
	"github.com/clivm/clivm/pkg/download"
	"github.com/clivm/clivm/pkg/exec"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/clivm/clivm/pkg/installpackage"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestController_Install(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		links             map[string]string
		param             *config.Param
		rt                *runtime.Runtime
		registryInstaller registry.Installer
		isErr             bool
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: clivm
  repo_name: clivm-installer
  path: clivm-installer
`,
				"/home/foo/.local/share/clivm/pkgs/github_content/github.com/clivm/clivm-installer/v1.0.0/clivm-installer/clivm-installer":                                              ``,
				fmt.Sprintf("/home/foo/.local/share/clivm/pkgs/github_release/github.com/clivm/clivm-proxy/%s/clivm-proxy_linux_amd64.tar.gz/clivm-proxy", installpackage.ProxyVersion): ``,
				"/home/foo/.local/share/clivm/bin/clivm-installer": ``,
				"/home/foo/.local/share/clivm/bin/clivm-proxy":     ``,
			},
			links: map[string]string{
				"clivm-proxy": "/home/foo/.local/share/clivm/bin/clivm-installer",
				fmt.Sprintf("../pkgs/github_release/github.com/clivm/clivm-proxy/%s/clivm-proxy_linux_amd64.tar.gz/clivm-proxy", installpackage.ProxyVersion): "/home/foo/.local/share/clivm/bin/clivm-proxy",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	registryDownloader := download.NewRegistryDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			linker := link.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			executor := exec.NewMock(0, nil)
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, fs, linker, executor)
			ctrl := install.New(d.param, finder.NewConfigFinder(fs), reader.New(fs), registry.New(d.param, registryDownloader, fs), pkgInstaller, fs, d.rt)
			if err := ctrl.Install(ctx, d.param, logE); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
