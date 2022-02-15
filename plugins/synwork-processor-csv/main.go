package main

import (
	"sbl.systems/go/synwork/plugin-sdk/plugin"
	"sbl.systems/go/synwork/plugins/synwork-processor-csv/csv"
)

func main() {

	plugin.Serve(csv.Opts)
}
