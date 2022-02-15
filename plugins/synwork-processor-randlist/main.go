package main

import (
	"sbl.systems/go/synwork/plugin-sdk/plugin"
	"sbl.systems/go/synwork/plugins/synwork-processor-randlist/randlist"
)

func main() {
	plugin.Serve(randlist.Opts)
}
