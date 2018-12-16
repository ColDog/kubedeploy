package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	minio "github.com/minio/minio-go"
	"github.com/urfave/cli"
)

var version = "UNRELEASED"

const (
	appFile = "app.yaml"
	bucket  = "builds"
)

// App holds all of our app configuration.
type App struct {
	Version      string                   `json:"version"`
	Name         string                   `json:"name"`
	Namespace    string                   `json:"namespace"`
	Source       string                   `json:"source"`
	Runtime      string                   `json:"runtime"`
	Build        []string                 `json:"build"`
	Requirements []map[string]interface{} `json:"requirements"`
}

type storeConfig struct {
	accessKey    string
	accessSecret string
	namespace    string
	service      string
}

func app() (*App, error) {
	f, err := os.Open(appFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", appFile, err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", appFile, err)
	}
	app := &App{}
	err = yaml.Unmarshal(data, app)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %v", appFile, err)
	}
	return app, nil
}

// buildCode uploades code described in app.Source as a tar gzip to minio inside
// the cluster.
func buildCode(buildDir string, app *App, c storeConfig) error {
	code := filepath.Join(buildDir, "build.tgz")
	err := sh("tar", "-czf", code, app.Source)
	if err != nil {
		return fmt.Errorf("packaging source failed: %v", err)
	}
	f, err := os.Open(code)
	if err != nil {
		return fmt.Errorf("reading source failed %s: %v", code, err)
	}
	defer f.Close()

	err = portForward(c.namespace, c.service, "9000", func(port string) error {
		endpoint := "127.0.0.1:" + port
		client, err := minio.New(endpoint, c.accessKey, c.accessSecret, false)
		if err != nil {
			return err
		}

		client.MakeBucket(bucket, "")
		_, err = client.FPutObject(
			bucket, app.Name+"-"+app.Version+".tgz", code,
			minio.PutObjectOptions{ContentType: "application/tar+gzip"},
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("uploading source failed %s: %v", code, err)
	}
	return nil
}

// buildChart builds a newly packaged chart that will live at
// <buildDir>/<app.Name>-<app.Version>.tgz
func buildChart(buildDir, chartURI string, app *App) (err error) {
	// Fetch a helm chart which should build a folder in <buildDir> named
	// <chartName>.
	err = sh("helm", "fetch", chartURI, "-d", buildDir, "--untar")
	if err != nil {
		return fmt.Errorf("helm fetch failed: %v", err)
	}
	chartName := filepath.Base(firstDir(buildDir))

	// Rename the chart to the new app name.
	err = os.Rename(
		filepath.Join(buildDir, chartName),
		filepath.Join(buildDir, app.Name),
	)
	if err != nil {
		return fmt.Errorf("chart rename failed: %v", err)
	}
	err = renameChart(app.Name, filepath.Join(buildDir, app.Name))
	if err != nil {
		return err
	}

	// Copy over the app file to store always in the helm chart.
	if err = copy(
		appFile, filepath.Join(buildDir, app.Name, appFile)); err != nil {
		return err
	}

	// Copy requirements over.
	reqs, err := yaml.Marshal(app.Requirements)
	if err != nil {
		return fmt.Errorf("failed to marshal requirements: %v", err)
	}
	err = ioutil.WriteFile(
		filepath.Join(buildDir, app.Name, "requirements.yaml"), reqs, 0644)
	if err != nil {
		return fmt.Errorf("failed to write requirements: %v", err)
	}

	// Package up this newly built chart updating any dependencies needed.
	err = sh(
		"helm", "package", "-d", buildDir, "-u", "--version", app.Version,
		filepath.Join(buildDir, app.Name),
	)
	if err != nil {
		return fmt.Errorf("helm package failed: %v", err)
	}
	return nil
}

func release(chartURI string, a *App, cfg storeConfig) error {
	return sh(
		"helm", "upgrade", "--install",
		"--set", "version="+a.Version,
		"--set", "store.key="+cfg.accessKey,
		"--set", "store.secret="+cfg.accessSecret,
		"--set", "store.service="+cfg.service,
		"--set", "store.namespace="+cfg.namespace,
		"-f", appFile, "--namespace", a.Namespace, a.Name, chartURI,
	)
}

func remove(a *App) error {
	return sh(
		"helm", "delete", "--purge", a.Name,
	)
}

func deployCmd(c *cli.Context) error {
	baseChartURI := c.GlobalString("base-chart")
	cfg := storeConfig{
		namespace:    c.GlobalString("store-namespace"),
		service:      c.GlobalString("store-service"),
		accessKey:    c.GlobalString("store-access-key"),
		accessSecret: c.GlobalString("store-secret-key"),
	}

	tmp, err := ioutil.TempDir("", "kube-")
	if err != nil {
		return err
	}

	a, err := app()
	if err != nil {
		return err
	}

	log("Packaging code", tmp)
	err = buildCode(tmp, a, cfg)
	if err != nil {
		return err
	}

	chartURI := tmp + "/" + a.Name + "-" + a.Version + ".tgz"
	log("Building a chart", chartURI)
	err = buildChart(tmp, baseChartURI, a)
	if err != nil {
		return err
	}

	log("Releasing chart", chartURI)
	err = release(chartURI, a, cfg)
	if err != nil {
		return err
	}

	log("deploy complete")
	return nil
}

func deleteCmd(c *cli.Context) error {
	a, err := app()
	if err != nil {
		return err
	}
	return remove(a)
}

func main() {
	app := cli.NewApp()
	app.Name = "kubedeploy"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "base-chart",
			Usage: "base chart which is extended with unique app definitions",
			Value: "https://github.com/ColDog/kubedeploy/releases/download/" + version + "/app-" + version + ".tgz",
		},
		cli.StringFlag{
			Name:  "store-namespace",
			Usage: "minio store namespace",
			Value: "kubedeploy",
		},
		cli.StringFlag{
			Name:  "store-service",
			Usage: "minio store service name",
			Value: "kubedeploy-minio",
		},
		cli.StringFlag{
			Name:  "store-access-key",
			Usage: "minio store access key",
			Value: "kubedeploy-key",
		},
		cli.StringFlag{
			Name:  "store-secret-key",
			Usage: "minio store secret access key",
			Value: "kubedeploy-secret",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "deploy",
			Usage:  "deploy a project in the current kubernetes cluster",
			Action: deployCmd,
		},
		cli.Command{
			Name:   "delete",
			Usage:  "delete the project in the current kubernetes cluster",
			Action: deleteCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
