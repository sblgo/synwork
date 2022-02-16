package runtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sbl.systems/go/synwork/synwork/processor/cfg"
)

type PluginKey struct {
	Hostname  string
	Namespace string
	Name      string
	Version   string
	OsArch    string
}

type PluginSource struct {
	PluginKey
	Source          string
	VersionSelector string
	Config          *cfg.Config
	PluginProgram   string
}

func (ps *PluginSource) verifyAndLoadPlugin() error {
	if err := ps.evalPluginKey(); err != nil {
		return err
	}
	if err := ps.evalRemotePluginProgram(); err != nil {
		return err
	}
	if err := ps.evalLocalPluginProgram(); err != nil {
		return err
	}

	return nil
}

func (ps *PluginSource) verifyPlugin() error {
	if err := ps.evalPluginKey(); err != nil {
		return err
	}
	if err := ps.evalLocalPluginProgram(); err != nil {
		return err
	}

	return nil
}

func (pr *PluginSource) evalPluginKey() error {
	parts := strings.Split(pr.Source, "/")
	pr.Namespace = INTERN_PROVIDER
	pr.Hostname = INTERN_HOSTNAME
	pr.OsArch = pr.Config.OsArch
	switch len(parts) {
	case 1:
		pr.Name = parts[0]
	case 2:
		pr.Namespace = parts[0]
		pr.Name = parts[1]
	case 3:
		pr.Hostname = parts[0]
		pr.Namespace = parts[1]
		pr.Name = parts[2]
	default:
		return fmt.Errorf("plugin %s not found", pr.Name)
	}
	return nil
}

func (pr *PluginSource) evalLocalPluginProgram() error {

	parts := append(filepath.SplitList(pr.Config.PluginDir), pr.Hostname, pr.Namespace, pr.Name)
	progName := func(ver string) string {
		progParts := append(parts, ver, pr.Config.OsArch, fmt.Sprintf("synwork-processor-%s%s", pr.Name, pr.Config.ProgramExt))
		return filepath.Join(progParts...)
	}
	if child, err := ioutil.ReadDir(filepath.Join(parts...)); err != nil {
		return err
	} else {
		checkFile := func(finfo os.FileInfo) bool {
			fileName := progName(finfo.Name())
			if fi, err := os.Stat(fileName); err != nil {
				return false
			} else if !fi.IsDir() {
				return true
			}
			return false
		}
		availableVersions := []string{}
		for _, dir := range child {
			if dir.IsDir() && checkFile(dir) {
				availableVersions = append(availableVersions, dir.Name())
			}
		}
		sort.Strings(availableVersions)
		if len(availableVersions) == 0 {
			return fmt.Errorf("plugin %s not found", pr.Source)
		}
		pr.Version = availableVersions[len(availableVersions)-1]
		pr.PluginProgram = progName(pr.Version)

		return nil
	}
}

func (ps *PluginSource) evalRemotePluginProgram() error {
	return nil
}
