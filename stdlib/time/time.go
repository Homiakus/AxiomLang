package time

import "time"

// NowNs returns Unix timestamp in nanoseconds
func NowNs() int64 {
	return time.Now().UnixNano()
}

// NowMs returns Unix timestamp in milliseconds
func NowMs() int64 {
	return time.Now().UnixMilli()
}

// SleepMs halts current actor execution for ms milliseconds
func SleepMs(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
