package helm

import (
	"bytes"
	"encoding/base64"
	"io/fs"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type App struct {
	templateName string
	repo         string
	name         string
	namespace    string
	version      string
	chartDir     string

	helm *Helm
}

func (a *App) Chart() string {
	return a.repo + "/" + a.templateName
}

// TODO: use appName and templateName
func NewApp(name, namespace, templateName, repo string, version string, helm *Helm) *App {
	return &App{
		name:         name,
		namespace:    namespace,
		templateName: templateName,
		repo:         repo,
		version:      version,
		helm:         helm,
		chartDir:     "/tmp/helm/chart",
	}
}

func (a *App) Pull() error {
	client := action.NewPull()
	settings := cli.New()
	settings.RepositoryConfig = a.helm.repoFile
	settings.RepositoryCache = a.helm.repoCache
	client.Settings = settings
	client.DestDir = a.chartDir
	client.Version = a.version

	output, err := client.Run(a.chart())
	if err != nil {
		return err
	}
	logrus.Info(output)
	return nil
}

func (a *App) chart() string {
	return a.repo + "/" + a.templateName
}

func (a *App) PreInstall() error {
	var buf bytes.Buffer
	return a.helm.PreInstall(a.templateName, a.namespace, a.Chart(), &buf)
}

func (a *App) ParseChart() (string, error) {
	//chartPath := path.Join(a.chartDir, a.templateName + a.version + ".tgz")
	chartDir := path.Join(a.chartDir, a.templateName)

	var values string
	err := filepath.Walk(chartDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.Contains(path, "values.yaml") {
			return nil
		}

		file, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		values = base64.StdEncoding.EncodeToString(file)

		return nil
	})
	return values, err
}
