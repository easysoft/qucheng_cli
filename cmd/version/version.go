// Copyright (c) 2021-2022 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	gv "github.com/blang/semver/v4"
	"github.com/easysoft/qcadmin/common"
	"github.com/easysoft/qcadmin/internal/pkg/util/log"
	"github.com/easysoft/qcadmin/pkg/qucheng/upgrade"
	"github.com/ergoapi/util/color"
	"github.com/ergoapi/util/file"
	"github.com/imroc/req/v3"
	"github.com/pkg/errors"
)

var versionTpl = `{{with .Client -}}
Client:
 Version:           {{ .Version }}
 Go version:        {{ .GoVersion }}
 Git commit:        {{ .GitCommit }}
 Built:             {{ .BuildTime }}
 OS/Arch:           {{.Os}}/{{.Arch}}
 Experimental:      {{.Experimental}}
{{- if .CanUpgrade }}
 Note:              {{ .UpgradeMessage }}
 URL:               https://github.com/easysoft/qucheng_cli/releases/tag/v{{ .LastVersion }}
{{- end }}
{{- end}}
{{- if .ServerDeployed }}{{with .Server}}

Server:
 {{- range $component := .Components}}
 {{$component.Name}}:
{{- if $component.CanUpgrade }}
  AppVersion:       {{$component.Deploy.AppVersion}} --> {{$component.Remote.Version}}
  ChartVersion:     {{$component.Deploy.ChartVersion}} --> {{$component.Remote.Version}}
{{- else }}
  AppVersion:       {{$component.Deploy.AppVersion}}
  ChartVersion:     {{$component.Deploy.ChartVersion}}
{{- end }}
 {{- end}}
{{- if .CanUpgrade }}
  Note:              {{ .UpgradeMessage }}
{{- end }}
{{- end}}
{{- end}}
`

const (
	defaultVersion       = "0.0.0"
	defaultGitCommitHash = "a1b2c3d4"
	defaultBuildDate     = "Mon Aug  3 15:06:50 2020"
)

type versionGet struct {
	Code int `json:"code"`
	Data struct {
		Name    string    `json:"name"`
		Version string    `json:"version"`
		Sync    time.Time `json:"sync"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
}

type versionInfo struct {
	Client clientVersion
	Server *upgrade.Version
}

type clientVersion struct {
	Version        string
	LastVersion    string
	GoVersion      string
	GitCommit      string
	Os             string
	Arch           string
	BuildTime      string `json:",omitempty"`
	Experimental   bool
	CanUpgrade     bool
	UpgradeMessage string
}

// ServerDeployed returns true when the client could connect to the qucheng
func (v versionInfo) ServerDeployed() bool {
	return v.Server != nil
}

// PreCheckLatestVersion 检查最新版本
func PreCheckLatestVersion() (string, error) {
	lastVersion := &versionGet{}
	client := req.C().SetUserAgent(common.GetUG()).SetTimeout(time.Second * 5)
	_, err := client.R().SetResult(lastVersion).Get(common.GetAPI("/api/release/last/qcadmin"))
	if err != nil {
		return "", err
	}
	return lastVersion.Data.Version, nil
}

func ShowVersion() {
	// logo.PrintLogo()
	if common.Version == "" {
		common.Version = defaultVersion
	}
	if common.BuildDate == "" {
		common.BuildDate = defaultBuildDate
	}
	if common.GitCommitHash == "" {
		common.GitCommitHash = defaultGitCommitHash
	}
	tmpl, err := newVersionTemplate()
	if err != nil {
		log.Flog.Fatalf("gen version failed, reason: %v", err)
		return
	}
	vd := versionInfo{
		Client: clientVersion{
			Version:      common.Version,
			GoVersion:    runtime.Version(),
			GitCommit:    common.GitCommitHash,
			BuildTime:    common.BuildDate,
			Os:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			Experimental: true,
		},
	}
	log.Flog.StartWait("check update...")
	lastversion, err := PreCheckLatestVersion()
	log.Flog.StopWait()
	if err != nil {
		log.Flog.Debugf("get update message err: %v", err)
		return
	}
	if lastversion != "" && !strings.Contains(common.Version, lastversion) {
		nowversion := gv.MustParse(strings.TrimPrefix(common.Version, "v"))
		needupgrade := nowversion.LT(gv.MustParse(lastversion))
		// log.Flog.Debugf("lastversion: %s(%v), nowversion: %s(%v), needupgrade: %v", lastversion, gv.MustParse(lastversion), common.Version, nowversion, needupgrade)
		if needupgrade {
			vd.Client.CanUpgrade = true
			vd.Client.LastVersion = lastversion
			vd.Client.Version = color.SGreen(vd.Client.Version)
			vd.Client.UpgradeMessage = fmt.Sprintf("Now you can use %s to upgrade cli to the latest version %s", color.SGreen("%s upgrade q", os.Args[0]), color.SGreen(lastversion))
		}
	}
	if file.CheckFileExists(common.GetCustomConfig(common.InitFileName)) {
		qv, err := upgrade.QuchengVersion()
		if err == nil {
			vd.Server = &qv
			canUpgrade := false
			for _, component := range qv.Components {
				if component.CanUpgrade {
					canUpgrade = true
					break
				}
			}
			vd.Server.CanUpgrade = canUpgrade
			vd.Server.UpgradeMessage = fmt.Sprintf("Now you can use %s to upgrade  qucheng to the latest version %s", color.SGreen("%s upgrade manage upgrade ", os.Args[0]), color.SGreen(lastversion))
		}
	}
	if err := prettyPrintVersion(vd, tmpl); err != nil {
		panic(err)
	}
}

func prettyPrintVersion(vd versionInfo, tmpl *template.Template) error {
	t := tabwriter.NewWriter(os.Stdout, 20, 1, 1, ' ', 0)
	err := tmpl.Execute(t, vd)
	t.Write([]byte("\n"))
	t.Flush()
	return err
}

func newVersionTemplate() (*template.Template, error) {
	tmpl, err := template.New("version").Parse(versionTpl)
	return tmpl, errors.Wrap(err, "template parsing error")
}
