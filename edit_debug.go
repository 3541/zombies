// +build debug

// Graph editing. For debug builds

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/3541/zombies/vis"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

type state int

const (
	Main state = iota
	Input
	CreateVertex
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
			editor.currentState = Input
			editor.input = &inputState{"Vertex name: ", CreateVertex, new(bytes.Buffer)}
			editor.tempVertex = editor.g.NewPositionedNode("", pos.X, pos.Y)
			editor.statusText.Clear()
			editor.statusText.WriteString(editor.input.prompt)
		}
	case Input:
		if editor.window.JustPressed(pixelgl.KeyEnter) {
			editor.currentState = editor.input.nextState
			break
		}

		if editor.window.JustPressed(pixelgl.KeyBackspace) {
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
		n := editor.g.NewPositionedNode(editor.input.buffer.String(), editor.tempVertex.Pos.X, editor.tempVertex.Pos.Y)
		editor.g.AddNode(n)
		editor.currentState = Main
	}

}

//func editGraph(camera pixel.Matrix) {}

func editInit(window *pixelgl.Window, g *vis.MapGraph, fontFace font.Face) {
	atlas := text.NewAtlas(fontFace, text.ASCII)
	editor = editorState{atlas, text.New(pixel.V(10, window.Bounds().H()-50), atlas), fontFace, window, g, Main, nil, nil}
	editor.statusText.Color = colornames.Black

	fmt.Fprintln(editor.statusText, "NOTE: This is a DEBUG build with the map editor enabled.")
	fmt.Fprintln(editor.statusText, "Click to add a vertex at that position.")
	fmt.Fprintln(editor.statusText, "Or press the 'a' key to add a vertex at a specific position.")
	fmt.Fprintln(editor.statusText, "Click on a vertex to select it, then press delete to delete it, or click on another vertex to connect them.")
	fmt.Fprintln(editor.statusText, "If the two vertices are already connected, this will select the edge already existing, and pressing delete will remove it.")
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
	}
}
