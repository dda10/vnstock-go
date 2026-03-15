// Package all imports all connector implementations to register them.
// Import this package in your main application to enable all connectors.
package all

import (
	_ "github.com/dda10/vnstock-go/connector/binance" // Register Binance connector
	_ "github.com/dda10/vnstock-go/connector/dnse"    // Register DNSE connector
	_ "github.com/dda10/vnstock-go/connector/fmp"     // Register FMP connector
	_ "github.com/dda10/vnstock-go/connector/gold"    // Register GOLD connector
	_ "github.com/dda10/vnstock-go/connector/vci"     // Register VCI connector
)
