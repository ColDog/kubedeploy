package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
)

func copy(src, dst string) error {
	srcData, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", src, err)
	}
	err = ioutil.WriteFile(dst, srcData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", dst, err)
	}
	return nil
}

func sh(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func portForward(ns, svc, port string, fn func(port string) error) error {
	local := "9000"
	cmd := exec.Command("kubectl", "-n", ns, "port-forward", "svc/"+svc, local+":"+port)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	err := fn("9000")
	cmd.Process.Kill()
	cmd.Wait()
	return err
}

func renameChart(name, path string) error {
	chartPath := filepath.Join(path, "Chart.yaml")

	f, err := os.Open(chartPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", chartPath, err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", chartPath, err)
	}

	var chart map[string]interface{}
	err = yaml.Unmarshal(data, &chart)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", chartPath, err)
	}

	chart["name"] = name

	data, err = yaml.Marshal(chart)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", chartPath, err)
	}
	err = ioutil.WriteFile(chartPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", chartPath, err)
	}
	return nil
}

func firstDir(dir string) string {
	var out string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != dir {
			out = path
			return filepath.SkipDir
		}
		return nil
	})
	return out
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

func log(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
