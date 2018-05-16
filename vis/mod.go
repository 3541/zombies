// Graph and graph rendering

package vis

import (
	"fmt"
	"strings"

	"github.com/3541/zombies/entity"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

// Encapsulates graphics handles and such.
type VWindow struct {
	window     *pixelgl.Window
	draw       *imdraw.IMDraw
	atlas      *text.Atlas
	StatusText *text.Text

	Graph *entity.MapGraph
}

func NewVWindow(window *pixelgl.Window, draw *imdraw.IMDraw, statusAtlas *text.Atlas, labelAtlas *text.Atlas, bounds pixel.Rect) *VWindow {
	t := text.New(pixel.V(10, window.Bounds().H()-20), statusAtlas)
	t.Color = colornames.Black
	t.WriteString("Welcome!\n")
	t.WriteString("Press 's' to see the status of all people currently alive.\n")
	t.WriteString("Press 'k' to see a key detailing the shortened item names.\n")
	t.WriteString("Press 'c' to clear the status text.\n")
	t.WriteString("Use the arrow keys to move the camera, the '.' key to zoom, and the ',' key to zoom out.\n")

	return &VWindow{window, draw, statusAtlas, t /*(25.0 / bounds.W()) * window.Bounds().H()*/, entity.NewMapGraph(labelAtlas, bounds, 25)}
}

// Implements io.Writer for VWindow, allowing fmt.Println(w, ...) & co., with correct wrapping and scrolling.
func (w *VWindow) Write(p []byte) (int, error) {
	if w.StatusText.Bounds().H() >= w.window.Bounds().H()/2 {
		w.StatusText.Clear()
	}
	if w.StatusText.BoundsOf(string(p)).W() >= w.window.Bounds().W()/2 {
		fmt.Fprintln(w.StatusText, string(p[:100]))
		fmt.Fprint(w.StatusText, string(p[100:]))
	} else {
		fmt.Fprint(w.StatusText, string(p))
	}
	return len(p), nil
}

// Show a status string for all entities
func (w *VWindow) PrintEntityStatus() {
	//	w.StatusText.Clear()
	for _, v := range w.Graph.Nodes() {
		for _, p := range v.People {
			fmt.Fprintf(w, "%d (%s) is at %s with ", p.Id, p.Profession, strings.ToUpper(v.Name))
			if len(p.Items) > 0 {
				fmt.Fprint(w, p.Items[0].StringLong())
				for _, i := range p.Items[1:] {
					fmt.Fprintf(w, ", %s", i.StringLong())
				}
			} else {
				fmt.Fprint(w, "NOTHING")
			}
			fmt.Fprintln(w)
		}
	}
}

func (w *VWindow) Draw() {
	if w.Graph.Changed {
		w.draw.Reset()
		w.draw.Clear()
		w.draw.Color = colornames.Lightslategray
		for _, n := range w.Graph.Nodes() {
			// Draw vertex
			if len(n.Zombies) > 0 {
				w.draw.Color = colornames.Red
				w.draw.Push(n.Pos)
				w.draw.Circle(w.Graph.VertexSize+2, float64(2*len(n.Zombies)))
				w.draw.Color = colornames.Lightslategray
			}
			if len(n.People) > 0 {
				w.draw.Color = colornames.Green
				w.draw.Push(n.Pos)
				w.draw.Circle(w.Graph.VertexSize+4, float64(2*len(n.People)))
				w.draw.Color = colornames.Lightslategray
			}
			w.draw.Push(n.Pos)
			if n.Selected {
				w.draw.Circle(w.Graph.VertexSize, 4)
			} else {
				w.draw.Circle(w.Graph.VertexSize, 0)
			}

			// Draw edges from that vertex
			for _, t := range w.Graph.From(n) {
				t := t.(*entity.PositionedNode)
				w.draw.Push(n.Pos)
				w.draw.Push(t.Pos)
				weight, _ := w.Graph.Weight(n, t)
				w.draw.Line(weight * 2)
			}
		}

		w.draw.Push(pixel.V(0, 0))
		w.draw.Push(pixel.V(w.Graph.Bounds.W(), w.Graph.Bounds.H()))
		w.draw.Rectangle(2)

		w.Graph.Changed = false
	}

	w.draw.Draw(w.window)

	// Draw vertex names
	for _, n := range w.Graph.Nodes() {
		n.RenderedName.Draw(w.window, pixel.IM)
	}
}
