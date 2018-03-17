// +build debug

// Graph editing. For debug builds

package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/3541/zombies/vis"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type state int

const (
	Main state = iota
	Input
)

type editorState struct {
	atlas      *text.Atlas
	statusText *text.Text

	window *pixelgl.Window
	g      *vis.MapGraph

	currentState state
}

var editor editorState

/*func editGraph(camera pixel.Matrix) {
	editor.statusText.Draw(editor.window, pixel.IM.Scaled(editor.statusText.Orig, 2))

	var pos pixel.Vec
	var clicked bool

	if editor.window.JustPressed(pixelgl.MouseButtonLeft) {
		pos = camera.Unproject(editor.window.MousePosition())
		clicked = true
		fmt.Fprintln(editor.statusText, pos)
	}

	switch state {
		case Main:
			if clicked {

			}
		case Input:

	}
}*/

func editGraph(camera pixel.Matrix) {}

func editInit(window *pixelgl.Window, g *vis.MapGraph) {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	editor = editorState{atlas, text.New(pixel.V(10, window.Bounds().H()-50), atlas), window, g, Input}
	editor.statusText.Color = colornames.Black

	fmt.Fprintln(editor.statusText, "NOTE: This is a DEBUG build with the map editor enabled.")
	fmt.Fprintln(editor.statusText, "Click to add a vertex at that position.")
	fmt.Fprintln(editor.statusText, "Or press the 'a' key to add a vertex at a specific position.")
	fmt.Fprintln(editor.statusText, "Click on a vertex to select it, then press delete to delete it, or click on another vertex to connect them.")
	fmt.Fprintln(editor.statusText, "If the two vertices are already connected, this will select the edge already existing, and pressing delete will remove it.")
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
