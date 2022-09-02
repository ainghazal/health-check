package health

import "time"

//
// configurable parameters
//

// TODO make viper able to set all these vars
var (
	maxConcurrentCheckers      = 20
	runInterval                = 10 * time.Minute
	pingCount                  = 5   // DEBUG this should be lower for debug, 5-10 for prod.
	maxEndpointsToTest         = -1  // DEBUG, change-me for release (-1 for all).
	healthThresholdForPingLoss = 0.8 // below this loss rate, it's healthy
)
