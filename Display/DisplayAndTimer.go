package Display

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

//Display holds information necessary to display
type Display struct {
	window   *sdl.Window
	Texture  *sdl.Texture
	Renderer *sdl.Renderer
	Pixels   []byte
	bits     []byte
	bitPosY  byte
	bitPosX  byte
	pitch    int
	H        int
	W        int
	Dh       int
	Dw       int
}

//Creates a display to be used by the emulator
func InitDisplay() *Display {
	//Initializes SDL
	sdl.Init(sdl.INIT_EVERYTHING)
	//Creates a window
	window, err := sdl.CreateWindow("Chip 8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 320, 0)
	if err != nil {
		panic(sdl.GetError())
	}
	//Creates a renderer
	renderer, err := sdl.CreateRenderer(window, 0, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(sdl.GetError())
	}
	//Creates a texture which allows pixel manipulation
	pixelFormat, err := window.GetPixelFormat()
	if err != nil {
		panic(sdl.GetError())
	}
	texture, err := renderer.CreateTexture(pixelFormat, sdl.TEXTUREACCESS_STREAMING, 64, 32)
	if err != nil {
		panic(sdl.GetError())
	}
	//Sets pitch and Pixels
	_, pitch, err := texture.Lock(nil)
	var Pixels []byte
	if err != nil {
		panic(sdl.GetError())
	}
	texture.Unlock()
	Pixels = nil
	//Creates display and returns it
	displayOut := Display{window: window, Texture: texture, Renderer: renderer, Pixels: Pixels, pitch: pitch, H: 32, W: 64, Dh: 320, Dw: 640}
	return &displayOut
}

//Locks texture to write only
func (display *Display) Lock() {
	var err error
	display.Pixels, display.pitch, err = display.Texture.Lock(nil)
	if err != nil {
		panic(sdl.GetError())
	}
}

//Unlocks texture for reading and applies update to texture
func (display *Display) UnLock() {
	display.Texture.Unlock()
	display.Pixels = nil
	display.bits = nil
	display.pitch = 0
}

//Turns bits into bytes
func (display *Display) BitToByte() {
	bits := display.bits
	//Iterates through the height of the bits changed
	for i := 0; i < len(bits); i++ {
		//Iterates through each byte that will be changed
		for j := 0; j < 8; j++ {
			//Formula for finding position of pixel based on x and y
			//  (y - 1)(l) + (i * l) + Vx + (j * 4)
			//Sets it equal to the bit in the position we want times 255
			fmt.Printf("%b : ", (display.bits[i] << j >> 7))
			fmt.Println(int(display.bitPosY-1)*(display.pitch) + (int(i) * (display.pitch)) + (int(display.bitPosX) * 4) + (int(j) * 4))
			pos := (int(display.bitPosY-1) * display.pitch) + (i * display.pitch) + (int(display.bitPosX) * 4) + (j * 4)
			display.Pixels[pos] = ((^(bits[i] << j)) >> 7) * 0xFF
			display.Pixels[pos+1] = ((^(bits[i] << j)) >> 7) * 0xFF
			display.Pixels[pos+2] = ((^(bits[i] << j)) >> 7) * 0xFF
			display.Pixels[pos+3] = ((^(bits[i] << j)) >> 7) * 0xFF
		}
		fmt.Println()
	}
	fmt.Println()
}

//Opcode 00E0
//Clears the screen
func (display *Display) Clear() {
	//Locks display to write only
	display.Lock()
	//Changes every pixel to be white
	for i := 0; i < (display.pitch * display.H); i += 4 {
		display.Pixels[i] = 0xFF
		display.Pixels[i+1] = 0xFF
		display.Pixels[i+2] = 0xFF
		display.Pixels[i+3] = 0xFF
	}
	//Unlocks display from write to read
	display.UnLock()
	//Displays changes
	display.Renderer.Clear()
	display.Renderer.Copy(display.Texture, nil, &sdl.Rect{W: int32(display.Dw), H: int32(display.Dh)})
	display.Renderer.Present()
	display.Lock()
	display.UnLock()
}

//Gets the pitch
func (display *Display) GetPitch() int {
	return display.pitch
}

//Gets Pixels and returns the bit byte version
func (display *Display) GetPixels(Vy byte, Vx byte, n uint8) *[]byte {
	//Resets display.bits
	display.bits = make([]byte, n)
	//Loops through all the Pixels starting at (Vx,Vy) and ending at n + 1
	toBits := byte(0)
	for i := uint8(0); i < (n); i++ {
		toBits = 0
		//Loops through 8 Pixels of display and converts them into a byte
		for j := uint8(0); j < 8; j++ {
			//Switches through the positions
			//Position calculated by formula (y - 1)(l) + (i * l) + Vx + (j * 4)
			switch display.Pixels[(int(Vy-1)*(display.pitch+1))+(int(i)*(display.pitch))+int(Vx)+(int(j)*4)] {
			case 0xFF:
				toBits = toBits << 1
			case 0x0:
				toBits = toBits << 1
				toBits = toBits ^ (1 << j)
			}
		}
		display.bits[i] = toBits
	}
	display.bitPosX = Vx
	display.bitPosY = Vy
	return &display.bits
}

//--------------------------------Timer---------------------

func InitTimer() Timer {
	return Timer{ticksStart: 0, value: 0}
}

type Timer struct {
	ticksStart uint32
	value      byte
}

func (timer *Timer) Start(startValue byte) {
	timer.ticksStart = sdl.GetTicks()
	timer.value = startValue
}

func (timer *Timer) Run() {
	if timer.value != 0 {
		timer.value = timer.value - byte((sdl.GetTicks()-timer.ticksStart)/17)
	}
}

func (timer *Timer) GetValue() byte {
	timer.Run()
	return timer.value
}

func (timer *Timer) SetValue(value byte) {
	timer.value = value
}
