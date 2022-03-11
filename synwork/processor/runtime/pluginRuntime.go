package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	PluginKey
	SelectedVersion string
	pluginProgram   string
	name            string
	Config          *cfg.Config
}

func (pr *PluginRuntime) Start(ctx context.Context, from, to int) (*Plugin, error) {
	if port, ok := pr.debugMode(); ok {
		pl, err := NewPlugin(port, pr.Name, pr.name)
		if err == nil {
			err = pl.DebugEnv()
		}
		return pl, err
	}

	port, ok := pr.findPort(from, to)
	if !ok {
		return nil, fmt.Errorf("can't find free port in range %d to %d", from, to)
	}
	if name, err := pr.startPluginServer(ctx, port); err != nil {
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

func (pr *PluginRuntime) startPluginServer(ctx context.Context, port int) (string, error) {
	var err error
	for cnt := 0; cnt < 5; cnt++ {
		cmd := exec.Command(pr.pluginProgram, strconv.Itoa(port))
		cmd.Env = os.Environ()
		go pr.copyCmdOut(ctx, cmd.StderrPipe, os.Stderr)
		go pr.copyCmdOut(ctx, cmd.StdoutPipe, os.Stderr)
		err = cmd.Start()
		if err == nil {
			return pr.name, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return pr.name, err
}

func (pr *PluginRuntime) copyCmdOut(ctx context.Context, p func() (io.ReadCloser, error), w io.Writer) {
	r, err := p()
	if err != nil {
		return
	}
	reader := textproto.NewReader(bufio.NewReader(r))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line, err := reader.ReadContinuedLine()
		if err == nil {
			out := fmt.Sprintf("[%s] %s\n", pr.name, line)
			w.Write([]byte(out))
		}
	}
}
