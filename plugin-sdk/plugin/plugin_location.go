package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	ENV_SW_PROCESSOR_SOURCE = "SW_PROCESSOR_SOURCE"
	DEFAULT_HOSTNAME        = "sbl.systems"
	DEFAULT_NAMESPACE       = "synwork"
	PORT_FILE_NAME          = "synwork.port"
)

var PluginEnv = struct {
	CacheDir    string
	PluginDir   string
	OsArch      string
	ProgrammExt string
}{}

func init() {
	PluginEnv.OsArch = fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS == "windows" {
		PluginEnv.CacheDir = filepath.Join(append(filepath.SplitList(os.Getenv("HOME")), "AppData", "Roaming", "synwork.d")...)
		PluginEnv.ProgrammExt = ".exe"
	} else {
		PluginEnv.CacheDir = filepath.Join(append(filepath.SplitList(os.Getenv("HOME")), ".synwork.d")...)
	}
	PluginEnv.PluginDir = filepath.Join(PluginEnv.CacheDir, "plugins")
}

type PluginLocation struct {
	Source    string
	Hostname  string
	Namespace string
	Name      string
	Version   string
	Directory string
	Program   string
}

func NewPluginLocationFromEnv() *PluginLocation {
	return NewPluginLocation(os.Getenv(ENV_SW_PROCESSOR_SOURCE), "")
}

func NewPluginLocation(sourceStr string, versionDesc string) *PluginLocation {

	loc := &PluginLocation{
		Source:    sourceStr,
		Hostname:  DEFAULT_HOSTNAME,
		Namespace: DEFAULT_NAMESPACE,
		Version:   versionDesc,
	}
	source := strings.Split(loc.Source, "/")
	switch len(source) {
	case 1:
		loc.Name = source[0]
	case 2:
		loc.Namespace = source[0]
		loc.Name = source[1]
	case 3:
		loc.Hostname = source[0]
		loc.Namespace = source[1]
		loc.Name = source[2]
	case 4:
		loc.Hostname = source[0]
		loc.Namespace = source[1]
		loc.Name = source[2]
		loc.Version = source[3]
	default:
		return nil
	}

	pattern := NewPluginDesc(loc.Version)
	cands := loc.listVersionCandidates(pattern)
	loc.Version = loc.selectVersion(cands)
	if loc.Version == "" {
		loc.Version = "0.1"
	}
	processorName := fmt.Sprintf("synwork_processor_%s%s", loc.Name, PluginEnv.ProgrammExt)
	loc.Directory = filepath.Join(PluginEnv.PluginDir, loc.Hostname, loc.Namespace, loc.Name, loc.Version, PluginEnv.OsArch)
	loc.Program = filepath.Join(loc.Directory, processorName)
	return loc
}

func (pl *PluginLocation) listVersionCandidates(vp VersionPattern) []string {
	candidats := []string{}
	pluginDir := filepath.Join(PluginEnv.PluginDir, pl.Hostname, pl.Namespace, pl.Name)
	processorName := fmt.Sprintf("synwork_processor_%s%s", pl.Name, PluginEnv.ProgrammExt)
	existsProcess := func(vd string) bool {
		programm := filepath.Join(pluginDir, vd, PluginEnv.OsArch, processorName)
		if fileInfo, err := os.Stat(programm); err == nil && !fileInfo.IsDir() {
			return true
		}
		return false
	}
	if versDirs, err := os.ReadDir(pluginDir); err != nil {
		for _, vd := range versDirs {
			if vd.IsDir() && vp.Match(vd.Name()) && existsProcess(vd.Name()) {
				candidats = append(candidats, vd.Name())
			}
		}
	}
	return candidats
}

func (pl *PluginLocation) selectVersion(candidats []string) string {
	sort.Strings(candidats)
	if len(candidats) > 0 {
		return candidats[len(candidats)-1]
	}
	return ""
}

type VersionPattern interface {
	Match(string) bool
}

type pluginDesc struct {
}

func (pd pluginDesc) Match(v string) bool {
	return true
}

func NewPluginDesc(v string) VersionPattern {
	return pluginDesc{}
}
