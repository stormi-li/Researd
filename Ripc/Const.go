package ripc

import "time"

const const_separator = ":"
const const_lockExipireTime = 15
const const_lockWatchdogInterval = 5 * time.Second
const const_lockBlockingTime = 5 * time.Second
