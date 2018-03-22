// +build debug

// Graph editing. For debug builds

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/3541/zombies/vis"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/gonum/graph/simple"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

type state int

const (
	Main state = iota
	Input
	Selected
	CreateVertex
	CreateEdge
)

type editorState struct {
	atlas      *text.Atlas
	statusText *text.Text
	fontFace   font.Face

	window *pixelgl.Window
	g      *vis.MapGraph

	currentState state
	input        *inputState

	tempVertex *vis.PositionedNode
	tempEdge   *simple.Edge
	selected   *vis.PositionedNode
}

type inputState struct {
	prompt    string
	nextState state
	buffer    *bytes.Buffer
}

var editor editorState

func editGraph(camera pixel.Matrix) {
	editor.statusText.Draw(editor.window, pixel.IM)

	var pos pixel.Vec
	var clicked bool

	// Reset editor if escape is pressed
	if editor.window.JustPressed(pixelgl.KeyEscape) {
		editInit(editor.window, editor.g, editor.fontFace)
		return
	}

	if editor.window.JustPressed(pixelgl.MouseButtonLeft) {
		pos = camera.Unproject(editor.window.MousePosition())
		clicked = true
	}

	// Obviously input is annoying because this is running in a loop.
	// When user keyboard input is required, the next state (e.g., CreateVertex)
	// is stored, and the current state is switched to Input.
	// Post-Input states then return to Main once finished.
	switch editor.currentState {
	case Main:
		if clicked {
			v := clickedVertex(pos)
			if v == nil {
				editor.currentState = Input
				editor.input = &inputState{"Vertex name: ", CreateVertex, new(bytes.Buffer)}
				editor.tempVertex = editor.g.NewPositionedNode("", pos.X, pos.Y, -1)
				editor.statusText.Clear()
				editor.statusText.WriteString(editor.input.prompt)
			} else {
				editor.selected = v
				v.Selected = true
				editor.currentState = Selected
				editor.g.Changed = true
			}
		}
	case Input:
		if editor.window.JustPressed(pixelgl.KeyEnter) {
			editor.currentState = editor.input.nextState
			editor.statusText.Clear()
			break
		}

		if editor.window.JustPressed(pixelgl.KeyBackspace) && editor.input.buffer.Len() > 0 {
			editor.input.buffer.Truncate(editor.input.buffer.Len() - 1)
			editor.statusText.Clear()
			editor.statusText.WriteString(editor.input.prompt)
			editor.statusText.WriteString(editor.input.buffer.String())
			break
		}

		t := editor.window.Typed()
		editor.input.buffer.WriteString(t)
		editor.statusText.WriteString(t)
	case CreateVertex:
		if editor.tempVertex.Weight != -1 {
			n := editor.g.NewPositionedNode(editor.input.buffer.String(), editor.tempVertex.Pos.X, editor.tempVertex.Pos.Y, editor.tempVertex.Weight)
			editor.g.AddNode(n)
			editor.currentState = Main
			editor.tempVertex = nil
		} else {
			editor.input = &inputState{"Vertex weight: ", CreateVertex, new(bytes.Buffer)}
			editor.statusText.Clear()
			editor.statusText.WriteString(editor.input.prompt)
		}
	case Selected:
		if clicked {
			cv := clickedVertex(pos)
			if cv == nil {
				editor.currentState = Main
				editor.selected.Selected = false
				editor.selected = nil
				editor.g.Changed = true
			} else {
				editor.currentState = Input
				editor.input = &inputState{"Edge weight: ", CreateEdge, new(bytes.Buffer)}
				editor.tempEdge = &simple.Edge{simple.Node(editor.selected.ID()), simple.Node(cv.ID()), 1}
				editor.selected.Selected = false
				editor.selected = nil
				editor.statusText.Clear()
				editor.statusText.WriteString(editor.input.prompt)
			}
		} else if editor.window.JustPressed(pixelgl.KeyBackspace) || editor.window.JustPressed(pixelgl.KeyDelete) {
			for _, n := range editor.g.From(editor.selected) {
				editor.g.RemoveEdge(editor.g.EdgeBetween(editor.selected, n))
			}
			editor.g.RemoveNode(editor.selected)
			editor.currentState = Main
			editor.selected = nil
			editor.g.Changed = true
		}
	case CreateEdge:
		w, err := strconv.ParseFloat(editor.input.buffer.String(), 64)
		if err != nil {
			editor.statusText.WriteString("Must enter a valid number.")
		} else {
			editor.tempEdge.W = w
			if !editor.g.HasEdgeBetween(editor.tempEdge.F, editor.tempEdge.T) {
				editor.g.SetEdge(editor.tempEdge)
			} else if w == 0 {
				editor.g.RemoveEdge(editor.g.EdgeBetween(editor.tempEdge.T, editor.tempEdge.F))
			}
			editor.tempEdge = nil
			editor.g.Changed = true
		}
		editor.currentState = Main
	}
}

func clickedVertex(pos pixel.Vec) *vis.PositionedNode {
	for _, v := range editor.g.Nodes() {
		if math.Sqrt(math.Pow(pos.X-v.Pos.X, 2)+math.Pow(pos.Y-v.Pos.Y, 2)) < editor.g.VertexSize {
			return v
		}
	}
	return nil
}

//func editGraph(camera pixel.Matrix) {}

func editInit(window *pixelgl.Window, g *vis.MapGraph, fontFace font.Face) {
	atlas := text.NewAtlas(fontFace, text.ASCII)
	editor = editorState{atlas, text.New(pixel.V(10, window.Bounds().H()-50), atlas), fontFace, window, g, Main, nil, nil, nil, nil}
	editor.statusText.Color = colornames.Black

	fmt.Fprintln(editor.statusText, "NOTE: This is a DEBUG build with the map editor enabled.")
	fmt.Fprintln(editor.statusText, "Click to add a vertex at that position.")
	fmt.Fprintln(editor.statusText, "Or press the 'a' key to add a vertex at a specific position.")
	fmt.Fprintln(editor.statusText, "Click on a vertex to select it, then press delete to delete it, or click on another vertex to connect them.")
	fmt.Fprintln(editor.statusText, "Clicking two vertices already connected by an edge and entering a weight of 0 deletes the edge.")
	fmt.Fprintln(editor.statusText, "Press the escape key to reset the editor. No graph data will be lost, but any current editing actions will be removed,\nand this message will display again.")
}

func editEnd(window *pixelgl.Window, g *vis.MapGraph) {
	window.SetMonitor(nil)
	window.SetBounds(pixel.R(0, 0, 1, 1))
	var in string
	fmt.Print("Save [Y/n]? ")
	fmt.Scanln(&in)
	fmt.Println(in)

	if strings.ToLower(in) != "n" {
		s, err := g.Serialize()
		if err != nil {
			panic(err)
		}

		fmt.Print("Save to the following file [map.json]: ")
		fmt.Scanln(&in)
		if len(in) == 0 {
			in = "map.json"
		}
		err = ioutil.WriteFile(fmt.Sprintf("./%s", in), s, 0666)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Saved to %s\n", in)
	}
}
