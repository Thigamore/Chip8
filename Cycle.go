package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/thigamore/Chip8/Display"
	"github.com/veandco/go-sdl2/sdl"
)

//Struct that holds all information relating to Chip-8 Instance
type Chip8 struct {
	stack      Stack
	memory     []byte           //Size of 4096
	registers  []byte           //Size of 16
	I          uint16           //Address Register
	PC         uint16           //Program Counter
	IR         uint16           //Instruction Register
	display    *Display.Display //Dislpay Struct
	delayTimer Display.Timer    //Delay Timer
	soundTimer Display.Timer    //Sound Timer
}

//A stack that does stack things
type Stack struct {
	stack   []uint16
	pointer int
}

//Log struct to log things
type Log struct {
	toWrite string
	file    *os.File
}

//Creates a new instance of a Chip-8
//Should be used whenever a new ROM is used
func newInstance(romPath string) Chip8 {
	//Opens file and checks for errors
	reader, err := os.Open(romPath)
	if err != nil {
		fmt.Println("Error Opening the file")
		panic(0)
	}
	//Reads file and looks for errors
	ROM, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println("Error reading the file")
		panic(0)
	}
	//If ROM is too big to be loaded into memory, returns error
	if len(ROM) > 3584 {
		fmt.Println("ROM too large")
		panic(0)
	}
	instance := Chip8{stack: Stack{stack: make([]uint16, 30), pointer: -1},
		memory:     make([]byte, 4096),
		registers:  make([]byte, 16),
		I:          uint16(0),
		PC:         uint16(0x200),
		IR:         uint16(0),
		display:    Display.InitDisplay(),
		delayTimer: Display.InitTimer(),
		soundTimer: Display.InitTimer(),
	}
	fonts := []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80} //F
	for pos, i := range fonts {
		instance.memory[pos] = i
	}
	for pos, i := range ROM {
		instance.memory[pos+0x200] = i
	}
	return instance
}

//Push stack operation
//Accept an input and pushes it to top of stack | Increments PC by 1
func (Stack *Stack) push(toPush uint16) {
	Stack.stack[Stack.pointer+1] = toPush
	Stack.pointer += 1
}

//Pop stack operation
//Returns the top of stack | Decreases PC by 1
func (Stack *Stack) pop() uint16 {
	popped := Stack.stack[Stack.pointer]
	Stack.pointer -= 1
	return popped
}

//Adds to log file
func (log *Log) Add(toAdd string) {
	log.toWrite += toAdd
}

//Writes the to Add to the file
func (log *Log) toFile() {
	log.file.Write([]byte(log.toWrite))
}

//Opcode DXYN
//Draws on the screen
func (instance *Chip8) Draw(Vx byte, Vy byte, n uint8, debugLog *Log) {
	//Sets register VF to 0
	instance.registers[0xF] = 0
	//Locks to write only
	instance.display.Lock()
	//Sets address of PC to top of stack
	instance.stack.push(instance.PC)
	//Sets PC to I
	instance.PC = instance.I
	//Gets the bits of the texture
	bits := instance.display.GetPixels(instance.registers[Vy], instance.registers[Vx], n)
	if debugLog != nil {
		debugLog.Add("\n")
		debugLog.Add(fmt.Sprintf("Vx: %b\n", instance.registers[Vx]))
		debugLog.Add(fmt.Sprintf("Vy: %b\n", instance.registers[Vy]))
		debugLog.Add(fmt.Sprintf("I: %b\n", instance.I))
		debugLog.Add(fmt.Sprintf("PC: %b\n", instance.PC))
		debugLog.Add(fmt.Sprintf("Bit Before: %b\n", bits))
	}
	for i := uint8(0); i < n; i++ {
		if debugLog != nil {
			debugLog.Add(fmt.Sprintf("Bits Before "+strconv.FormatUint(uint64(i), 10)+": %b\n", (*bits)[i]))
		}
		//Register VF is set to 1 if a bit gets flipped
		//This is done by a boolean algebra statement and an or statement to check if a bit has already been flipped
		//Boolean Algebra statement where A = bits, B = memory, and C is VF register
		//AB | C
		instance.registers[0xF] = (*bits)[i]&instance.memory[instance.I+uint16(i)] | instance.registers[0xF]
		//Flips bit if memory number is 1, bit stays the same if memory is 0
		//Boolean Algebra statement where A = bits and B = memory
		//(-A)B + A(-B)
		(*bits)[i] = ((^(*bits)[i]) & instance.memory[instance.I+uint16(i)]) | ((*bits)[i] & (^instance.memory[instance.I+uint16(i)]))
		if debugLog != nil {
			debugLog.Add(fmt.Sprintf("Memory "+strconv.FormatUint(uint64(i), 10)+": %b\n", instance.memory[instance.I+uint16(i)]))
			debugLog.Add(fmt.Sprintf("Bits After "+strconv.FormatUint(uint64(i), 10)+": %b\n", (*bits)[i]))
		}
	}
	//Transforms the bits that were altered to bytes to be displayed
	instance.display.BitToByte()
	if debugLog != nil {
		debugLog.Add(fmt.Sprintf("Bit After: %b\n", bits))
		debugLog.Add(fmt.Sprintf("Pixels: %b\n", instance.display.Pixels))
		debugLog.Add(fmt.Sprintf("VF: %b\n", instance.registers[0xF]))
		debugLog.Add("\n\n")
	}
	//Unlocks texture to update teture
	instance.display.UnLock()
	//Renders the texture
	instance.display.Renderer.Copy(instance.display.Texture, nil, &sdl.Rect{W: int32(instance.display.Dw), H: int32(instance.display.Dh)})
	instance.display.Renderer.Present()
	//Returns the pc to its original value
	instance.PC = instance.stack.pop()
	//sdl.Delay(1000)
}

//Does the Fetch, Decode, Execute Cycle for the Chip-8
//Also runs a clock to make sure it is running at a correct speed.
func (instance *Chip8) Cycle(debugMode bool) {
	//Creates logs for debugging purposes
	log := Log{}
	if debugMode {
		file, err := os.Create("Log.txt")
		if err != nil {
			panic("Unable to create file")
		}
		log = Log{file: file, toWrite: ""}
	}
	var e sdl.Event
	running := true
	for running {
		//Fetches
		instance.Fetch()
		//Decodes
		instance.DecodeExecute(&log)
		if debugMode {
			log.Add(fmt.Sprintf("IR: %X\n", instance.IR))
			log.Add(fmt.Sprintf("PC: %X\n", instance.PC))
			log.Add(fmt.Sprintf("OPcode: %b\n", instance.IR>>12))
			log.Add(fmt.Sprintf("Registers: %X\n", instance.registers))
			log.Add("\n")
		}
		e = sdl.PollEvent()
		for e != nil {
			if e.GetType() == sdl.QUIT {
				running = false
			}
			e = sdl.PollEvent()
		}
	}
	log.toFile()
}

//Fetches the next instruction
func (instance *Chip8) Fetch() {
	//Moves first intruction 8 to the left and then XOR the following byte
	//To mix both of them into one uint16
	instance.IR = uint16(instance.memory[instance.PC])<<8 ^ uint16(instance.memory[instance.PC+1])
	instance.PC += 2
}

//Decodes the current instruction
func (instance *Chip8) DecodeExecute(debugLog *Log) {
	//Gets first 4 bits of the Instruction and matches it with an opcode
	switch instance.IR >> 12 {
	case 0x0:
		//Different opcodes start at 0 so matches it to the correct one
		switch instance.IR {
		case 0x00E0:
			instance.display.Clear()
		case 0x00EE: //return from a subroutine
			instance.PC = instance.stack.pop()
		default:
			//Call (Omitted)
		}
	case 0x1: //Jump to NNN
		instance.PC = instance.IR << 4 >> 4
	case 0x2: //Call subroutine at NNN
		instance.stack.push(instance.PC)
		instance.PC = instance.IR << 4 >> 4
	case 0x3: //If Vx == NN, skip
		Vx := instance.IR << 4 >> 12
		if instance.registers[Vx] == byte(instance.IR<<8>>8) {
			instance.PC += 2
		}
	case 0x4: //If Vx != NN, skip
		Vx := instance.IR << 4 >> 12
		if instance.registers[Vx] != byte(instance.IR<<8>>8) {
			instance.PC += 2
		}
	case 0x5: //If Vx == Vy, skip
		Vx := instance.IR << 4 >> 12
		Vy := instance.IR << 8 >> 12
		if instance.registers[Vx] == instance.registers[Vy] {
			instance.PC += 2
		}
	case 0x6: //Vx = NN
		Vx := instance.IR << 4 >> 12
		instance.registers[Vx] = byte(instance.IR << 8 >> 8)
	case 0x7: //Vx += NN
		Vx := instance.IR << 4 >> 12
		instance.registers[Vx] += byte(instance.IR << 8 >> 8)
	case 0x8:
		//Different opcodes start with 8 so match to the correct one
		switch instance.IR << 12 >> 12 {
		case 0x0: //Vx = Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			instance.registers[Vx] = instance.registers[Vy]
		case 0x1: //Vx = Vx|Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			instance.registers[Vx] = instance.registers[Vx] | instance.registers[Vy]
		case 0x2: //Vx = Vx & Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			instance.registers[Vx] = instance.registers[Vx] & instance.registers[Vy]
		case 0x3: //Vx = Vx ^ Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			instance.registers[Vx] = instance.registers[Vx] ^ instance.registers[Vy]
		case 0x4: //Vx += Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			sum := instance.registers[Vx] + instance.registers[Vy]
			if int(sum) != (int(instance.registers[Vx]) + int(instance.registers[Vy])) {
				instance.registers[15] = 1
			}
			instance.registers[Vx] = sum
		case 0x5: //Vx -= Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			sum := instance.registers[Vx] - instance.registers[Vy]
			if int(sum) != (int(instance.registers[Vx]) - int(instance.registers[Vy])) {
				instance.registers[0xF] = 1
			}
			instance.registers[Vx] = sum
		case 0x6: //Vx >>= 1
			Vx := instance.IR << 4 >> 12
			instance.registers[0xF] = instance.registers[Vx] << 7 >> 7
			instance.registers[Vx] = instance.registers[Vx] >> 1
		case 0x7: //Vx = Vy - Vx
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			sum := instance.registers[Vy] - instance.registers[Vx]
			if int(sum) != (int(instance.registers[Vy]) - int(instance.registers[Vx])) {
				instance.registers[0xF] = 1
			}
			instance.registers[Vx] = sum
		case 0x8: // Vx <<= 1
			Vx := instance.IR << 4 >> 12
			instance.registers[0xF] = instance.registers[Vx] >> 7
			instance.registers[Vx] = instance.registers[Vx] << 1
		}
	case 0x9: //if Vx != Vy, skip
		Vx := instance.IR << 4 >> 12
		Vy := instance.IR << 8 >> 12
		if instance.registers[Vx] != instance.registers[Vy] {
			instance.PC += 2
		}
	case 0xA: // I = NNN
		instance.I = instance.IR << 4 >> 4
	case 0xB: // PC = V0 + NNN
		instance.PC = uint16(instance.registers[0x0]) + (instance.IR << 4 >> 4)
	case 0xC: // Vx = rand() & NN
		rand.Seed(time.Now().UnixNano())
		Vx := instance.IR << 4 >> 12
		instance.registers[Vx] = byte(rand.Intn(255)) & byte(instance.IR<<8>>12)
	case 0xD: /// Draw( Vx, Vy, N )
		Vx := instance.IR << 4 >> 12
		Vy := instance.IR << 8 >> 12
		n := instance.IR << 12 >> 12
		instance.Draw(byte(Vx), byte(Vy), uint8(n), debugLog)
	case 0xE: // Multiple Skip Instructions
		switch instance.IR << 8 >> 8 {
		case 0x9E: // if key() == Vx, skip
			Vx := instance.IR << 4 >> 12
			e := sdl.PollEvent()
			key := instance.registers[Vx]
			switch t := e.(type) {
			case *sdl.KeyboardEvent:
				if sdl.GetKeyFromName(strconv.FormatUint(uint64(key), 10)) == t.Keysym.Sym {
					instance.PC += 2
				}
			}
		case 0xA1: // if key() != Vx, skip
			Vx := instance.IR << 4 >> 12
			e := sdl.PollEvent()
			key := instance.registers[Vx]
			switch t := e.(type) {
			case *sdl.KeyboardEvent:
				if sdl.GetKeyFromName(strconv.FormatUint(uint64(key), 10)) != t.Keysym.Sym {
					instance.PC += 2
				}
			}
		}
	case 0xF: // Multitude of different instructions
		switch instance.IR << 8 >> 12 {
		case 0x07: // Vx = get_delay()
			Vx := instance.IR << 4 >> 12
			instance.registers[Vx] = instance.delayTimer.GetValue()
		case 0x0A: // Vx = get_key()
			Vx := instance.IR << 4 >> 12
			var e sdl.Event
			running := true
			for running {
				e = sdl.PollEvent()
				for e != nil {
					switch t := e.(type) {
					case *sdl.KeyboardEvent:
						key, err := strconv.ParseUint(sdl.GetKeyName(t.Keysym.Sym), 10, 8)
						if err != nil {
							panic(sdl.GetError())
						}
						instance.registers[Vx] = byte(key)
					}
					e = sdl.PollEvent()
				}
			}
		case 0x15: // delay_timer(Vx)
			Vx := instance.IR << 4 >> 12
			instance.delayTimer.Start(instance.registers[Vx])
		case 0x18: // sound_timer(Vx)
			Vx := instance.IR << 4 >> 12
			instance.soundTimer.Start(instance.registers[Vx])
		case 0x1E: // I += Vx
			Vx := instance.IR << 4 >> 12
			instance.I += uint16(instance.registers[Vx])
		case 0x29: // I = sprite_address[Vx]
			Vx := instance.IR << 4 >> 12
			spriteAddress := instance.memory[instance.registers[Vx]*5]
			instance.I = uint16(spriteAddress)
		case 0x33: // Take decimal version of Vx, put hundreds in position I, tens in I + 1, ones in I + 2
			Vx := instance.IR << 4 >> 12
			hundreds := instance.registers[Vx] / 100
			tens := (instance.registers[Vx] / 10) - (hundreds * 10)
			ones := (instance.registers[Vx]) - (tens * 10) - (hundreds * 100)
			instance.memory[instance.I] = hundreds
			instance.memory[instance.I+1] = tens
			instance.memory[instance.I+2] = ones
		case 0x55: // reg_dump(Vx, &I)
			Vx := instance.IR << 4 >> 12
			for i := uint16(0); i < (Vx + 1); i++ {
				instance.memory[instance.I+i] = instance.registers[i]
			}
		case 0x65: // reg_kiad(Vx, &I)
			Vx := instance.IR << 4 >> 12
			for i := uint16(16); i < (Vx + 1); i++ {
				instance.registers[i] = instance.memory[instance.I+i]
			}
		}
	}
}

//Main function which runs the whole thing
func main() {
	sdl.Init(sdl.INIT_EVERYTHING)
	instance := newInstance("test_opcode.ch8")
	instance.display.Renderer.SetDrawColor(0xFF, 0xFF, 0xFF, 0xFF)
	instance.Cycle(true)
	/*
		instance.display.Lock()
		instance.display.Pixels[14*256+19] = 0xFF
		instance.display.Pixels[14*256+20] = 0xFF
		instance.display.Pixels[14*256+21] = 0xFF
		instance.display.Pixels[14*256+22] = 0xFF
		instance.display.Pixels[14*256+23] = 0xFF
		instance.display.Pixels[14*256+24] = 0xFF
		instance.display.Pixels[14*256+25] = 0xFF
		instance.display.Pixels[14*256+26] = 0xFF
		instance.display.Pixels[14*256+27] = 0xFF
		instance.display.Pixels[14*256+28] = 0xFF
		instance.display.Pixels[14*256+29] = 0xFF
		instance.display.Pixels[14*256+30] = 0xFF
		instance.display.Pixels[14*256+31] = 0xFF
		instance.display.Pixels[14*256+32] = 0xFF
		instance.display.Pixels[14*256+33] = 0xFF
		instance.display.Pixels[14*256+34] = 0xFF
		fmt.Println(instance.display.Pixels)
		fmt.Printf("%b", *instance.display.GetPixels(15, 15, 10))
		instance.display.UnLock()
		instance.display.Renderer.Copy(instance.display.Texture, nil, &sdl.Rect{W: int32(instance.display.Dw), H: int32(instance.display.Dh)})
		instance.display.Renderer.Present()
		sdl.Delay(1000)
		instance.display.Clear()
		instance.display.Lock()
		fmt.Println(instance.display.Pixels)
		fmt.Print(14*256 + 31)
		fmt.Println(instance.display.Pixels[14*256+31])
		instance.display.UnLock()
		sdl.Delay(1000)
	*/

}
