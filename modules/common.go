package modules

import "time"

// Only keep TimeNowUTC here
func TimeNowUTC() time.Time {
	return time.Now().UTC()
}
