// Entity behavior

package behavior

import (
	"time"

	"github.com/3541/zombies/vis"
)

func Start(log chan string, graph *vis.MapGraph) {
	tick := time.NewTicker(time.Second)
	for _ = range tick.C {
		/*for v := range graph.Nodes() {

		}*/

	}
}
