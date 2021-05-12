package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

//Display holds information necessary to display
type display struct {
	window   *sdl.Window
	texture  *sdl.Texture
	renderer *sdl.Renderer
	pixels   []byte
	pitch    int
	h        int
	w        int
}

//Creates a display to be used by the emulator
func InitDisplay() *display {
	//Initializes SDL
	sdl.Init(sdl.INIT_EVERYTHING)
	//Creates a window
	window, err := sdl.CreateWindow("Chip 8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 64, 32, 0)
	if err != nil {
		panic(sdl.GetError())
	}
	renderer, err := sdl.CreateRenderer(window, 0, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(sdl.GetError())
	}
	pixelFormat, err := window.GetPixelFormat()
	if err != nil {
		panic(sdl.GetError())
	}
	texture, err := renderer.CreateTexture(pixelFormat, sdl.TEXTUREACCESS_STREAMING, 64, 32)
	if err != nil {
		panic(sdl.GetError())
	}
	_, pitch, err := texture.Lock(nil)
	var pixels []byte
	if err != nil {
		panic(sdl.GetError())
	}
	texture.Unlock()
	pixels = nil
	displayOut := display{window: window, texture: texture, renderer: renderer, pixels: pixels, pitch: pitch, h: 32, w: 64}
	return &displayOut
}

//Locks texture
func (display *display) Lock() {
	var err error
	display.pixels, display.pitch, err = display.texture.Lock(nil)
	if err != nil {
		panic(sdl.GetError())
	}
}

//Unlocks texture
func (display *display) UnLock() {
	display.texture.Unlock()
}

//Opcode 00E0
//Clears the screen
func (display *display) Clear() {
	display.Lock()
	for i := 0; i < (display.pitch * display.h); i += 4 {
		display.pixels[i] = 0xFF
		display.pixels[i+1] = 0xFF
		display.pixels[i+2] = 0xFF
		display.pixels[i+3] = 0xFF
	}
	display.UnLock()
}
