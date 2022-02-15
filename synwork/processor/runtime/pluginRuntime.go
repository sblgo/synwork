package runtime

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"sbl.systems/go/synwork/plugin-sdk/plugin"
	"sbl.systems/go/synwork/synwork/processor/cfg"
)

const (
	OS_ENV_INTERN_PLUGINS string = "PLUGIN_PATH_INTERN"
	OS_ENV_EXTERN_PLUGINS string = "PLUGIN_PATH_EXTERN"
	INTERN_PROVIDER              = "synwork"
	INTERN_HOSTNAME              = "sbl.systems"
)

type PluginRuntime struct {
	Source          string
	Hostname        string
	Namespace       string
	Name            string
	Version         string
	SelectedVersion string
	OsArch          string
	Programm        string
	pluginProgram   string
	name            string
	Config          *cfg.Config
}

func (pr *PluginRuntime) Start(from, to int) (*Plugin, error) {
	if err := pr.evalPluginProgram(); err != nil {
		return nil, err
	}
	if port, ok := pr.debugMode(); ok {
		return NewPlugin(port, pr.Name, pr.name)
	}

	port, ok := pr.findPort(from, to)
	if !ok {
		return nil, fmt.Errorf("can't find free port in range %d to %d", from, to)
	}
	if name, err := pr.startPluginServer(port); err != nil {
		return nil, err
	} else {

		return NewPlugin(port, pr.Name, name)
	}

}

func (pr *PluginRuntime) debugMode() (int, bool) {
	fileName := filepath.Join(filepath.Dir(pr.pluginProgram), plugin.PORT_FILE_NAME)
	if portBytes, err := os.ReadFile(fileName); err == nil {
		if port, err := strconv.Atoi(string(portBytes)); err != nil {
			panic(err)
		} else {
			return port, true
		}
	}

	return 0, false
}

func (pr *PluginRuntime) findPort(from, to int) (int, bool) {
	port := from
	for ; port <= to; port++ {
		if l, e := net.Listen("tcp", fmt.Sprintf(":%d", port)); e == nil {
			l.Close()
			return port, true
		}
	}
	return 0, false
}

func (pr *PluginRuntime) evalPluginProgram() error {
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
	parts = append(filepath.SplitList(pr.Config.PluginDir), pr.Hostname, pr.Namespace, pr.Name)
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
		pr.SelectedVersion = availableVersions[len(availableVersions)-1]
		pr.pluginProgram = progName(pr.SelectedVersion)
		return nil
	}
}

func (pr *PluginRuntime) startPluginServer(port int) (string, error) {
	var err error
	for cnt := 0; cnt < 5; cnt++ {
		cmd := exec.Command(pr.pluginProgram, strconv.Itoa(port))
		err = cmd.Start()
		if err == nil {
			return pr.name, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pr.name, err
}
