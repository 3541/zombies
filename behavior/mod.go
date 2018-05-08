// Entity behavior

package behavior

import "time"

func Start(log chan string) {
	tick := time.NewTicker(time.Second)
	for _ = range tick.C {
		log <- time.Now().String()
	}
}
