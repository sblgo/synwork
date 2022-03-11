package cfg

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	OsArch     string
	WorkDir    string
	PortFrom   int
	PortTo     int
	CacheDir   string
	PluginDir  string
	ProgramExt string
}

func ParseArgs(cmds []string) ([]string, *Config, error) {
	prgs := map[string]struct{}{}
	posCmd := 0
	for _, a := range cmds {
		prgs[a] = struct{}{}
	}
	for p, a := range os.Args {
		if _, ok := prgs[a]; ok {
			posCmd = p
			break
		}
	}
	cfg, err := parseConfig(os.Args[1:posCmd])
	if err != nil {
		return nil, nil, err
	}
	if err = cfg.evalOsArch(); err != nil {
		return nil, nil, err
	}
	if err = cfg.evalCacheDir(); err != nil {
		return nil, nil, err
	}
	return os.Args[posCmd:], cfg, nil
}

func parseConfig(args []string) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("general", flag.ExitOnError)
	fs.StringVar(&cfg.WorkDir, "f", ".", "directory with configuration")
	fs.IntVar(&cfg.PortFrom, "pf", 50000, "start evaluation for free ports")
	fs.IntVar(&cfg.PortTo, "pt", 60000, "start evaluation for free ports")
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) evalOsArch() error {
	os, arch := runtime.GOOS, runtime.GOARCH
	cfg.OsArch = fmt.Sprintf("%s_%s", os, arch)
	return nil
}

func (cfg *Config) evalCacheDir() error {
	log.Println(os.Getenv("HOME"))
	home := filepath.SplitList(os.Getenv("HOME"))
	if len(home) == 0 {
		home = []string{"~"}
	}
	if runtime.GOOS == "windows" {
		cfg.CacheDir = filepath.Join(append(home, "AppData", "Roaming", "synwork.d")...)
		cfg.ProgramExt = ".exe"
	} else {
		cfg.CacheDir = filepath.Join(append(home, ".synwork.d")...)
	}
	cfg.PluginDir = filepath.Join(cfg.CacheDir, "plugins")
	return nil
}
