package plugin

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type PluginOptions struct {
	Provider func() schema.Processor
}

func Serve(opts PluginOptions) {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	if !Debug(opts) {
		Run(opts)
	}
}
func Run(opts PluginOptions) {
	port, err := findPort()
	if err != nil {
		log.Fatal(err)
	}
	plg := &Plugin{
		shutdown: make(chan struct{}, 1),
		provider: opts.Provider,
	}
	rpc.Register(plg)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
	<-plg.shutdown
}

func Debug(opts PluginOptions) bool {
	portString := os.Getenv("PLUGIN_DEBUG_PORT")
	if port, err := strconv.Atoi(portString); err != nil {
		return false
	} else {
		pluginLocation := NewPluginLocationFromEnv()
		if pluginLocation == nil {
			return false
		}
		fileName := filepath.Join(pluginLocation.Directory, PORT_FILE_NAME)
		if err := os.WriteFile(fileName, []byte(portString), 0644); err != nil {
			panic(err)
		}
		defer func() { os.Remove(fileName) }()
		plg := &Plugin{
			shutdown: make(chan struct{}, 1),
			provider: opts.Provider,
		}
		rpc.Register(plg)
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if e != nil {
			log.Fatal("listen error:", e)
		}
		go http.Serve(l, nil)
		<-plg.shutdown
	}

	return true
}

func findPort() (int, error) {
	args := os.Args
	port, err := strconv.Atoi(args[len(args)-1])
	return port, err
}
