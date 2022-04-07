package exts

import (
	"context"

	"github.com/hashicorp/go-version"
)

type (
	ProcessorKey struct {
		Hostname      string
		Namespace     string
		Name          string
		Version       string
		OsArch        string
		Ext           string
		ParsedVersion *version.Version
	}

	ProcessorSource struct {
		Version          *version.Version
		ProcessorProgram func(ctx context.Context) ([]byte, error)
	}
	ProcessorSources          []ProcessorSource
	ProcessorProvider         func(ctx context.Context, key *ProcessorKey) (ProcessorSources, error)
	ProcessorProviderRegistry map[string]ProcessorProvider
)

var (
	_processorProviderRegistry = ProcessorProviderRegistry{}
	_processorProviders        = []ProcessorProvider{}
)

func RegisterProcessorProvider(providerName string, provider ProcessorProvider) {
	if _, ok := _processorProviderRegistry[providerName]; !ok {
		_processorProviderRegistry[providerName] = provider
		_processorProviders = append(_processorProviders, provider)
	}
}

func ProcessorProviders() []ProcessorProvider {
	return _processorProviders
}
