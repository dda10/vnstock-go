package vnstock

import (
	"log/slog"
	"net/http"
	"sync"
)

// ConnectorFactory is a function that creates a new Connector instance.
type ConnectorFactory func(httpClient *http.Client, logger *slog.Logger) Connector

var (
	connectorRegistry = make(map[string]ConnectorFactory)
	registryMu        sync.RWMutex
)

// RegisterConnector registers a connector factory with the given name.
// This is called by connector packages in their init() functions.
func RegisterConnector(name string, factory ConnectorFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	connectorRegistry[name] = factory
}

// getConnectorFactory retrieves a registered connector factory by name.
func getConnectorFactory(name string) (ConnectorFactory, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, ok := connectorRegistry[name]
	return factory, ok
}
