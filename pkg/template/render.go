package template

import (
	"text/template"

	"github.com/aquaproj/aqua/pkg/runtime"
)

type Artifact struct {
	Version string
	OS      string
	Arch    string
	Format  string
	Asset   string
}

func Render(s string, art *Artifact, rt *runtime.Runtime) (string, error) {
	return Execute(s, map[string]interface{}{
		"Version": art.Version,
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      art.OS,
		"Arch":    art.Arch,
		"Format":  art.Format,
		"Asset":   art.Asset,
	})
}

func RenderTemplate(tpl *template.Template, art *Artifact, rt *runtime.Runtime) (string, error) {
	return ExecuteTemplate(tpl, map[string]interface{}{
		"Version": art.Version,
		"GOOS":    rt.GOOS,
		"GOARCH":  rt.GOARCH,
		"OS":      art.OS,
		"Arch":    art.Arch,
		"Format":  art.Format,
		"Asset":   art.Asset,
	})
}
