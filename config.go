package main

import "time"

//
// configurable parameters
//

// TODO make viper able to set all these vars
var (
	runInterval        = 60 * time.Second
	maxEndpointsToTest = 5 // DEBUG, change-me for release (-1 for all).
	pingCount          = 3 // DEBUG this should be lower for debug, 5-10 for prod.

	// TODO(ainghazal) the granularity in here is too big with n=5. We could raise to n=10 and reduce the inter-ping time to keep the measurement "fast".
	healthThresholdForPingLoss = 0.8 // below this loss rate, it's healthy
)
