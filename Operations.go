package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

//Struct that holds all information relating to Chip-8 Instance
type Chip8 struct {
	stack     Stack
	memory    []byte //Size of 4096
	registers []byte //Size of 16
	I         uint16 //Address Register
	PC        uint16 //Program Counter
	IR        uint16 //Instruction Register
}

//A stack that does stack things
type Stack struct {
	stack   []uint16
	pointer int
}

//Creates a new instance of a Chip-8
//Should be used whenever a new ROM is used
func newInstance(ROM []byte) Chip8 {
	instance := Chip8{stack: Stack{stack: make([]uint16, 30), pointer: -1},
		memory:    make([]byte, 4096),
		registers: make([]byte, 16),
		I:         uint16(0),
		PC:        uint16(0x200),
		IR:        uint16(0)}
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

//Does the Fetch, Decode, Execute Cycle for the Chip-8
//Also runs a clock to make sure it is running at a correct speed.
func (instance *Chip8) Cycle(debugMode bool) {
	//Fetches
	instance.Fetch()
	//Decodes
	instance.DecodeExecute()
	if debugMode {
		fmt.Println("IR: ", instance.IR)
		fmt.Println("PC: ", instance.PC)
		fmt.Println("OPcode:", instance.IR>>12)
	}
}

//Fetches the next instruction
func (instance *Chip8) Fetch() {
	//Moves first intruction 8 to the left and then XOR the following byte
	//To mix both of them into one uint16
	instance.IR = uint16(instance.memory[instance.PC])<<8 ^ uint16(instance.memory[instance.PC+1])
	instance.PC += 2
}

//Decodes the current instruction
func (instance *Chip8) DecodeExecute() {
	//Gets first 4 bits of the Instruction and matches it with an opcode
	switch instance.IR >> 12 {
	case 0x0:
		//Different opcodes start at 0 so matches it to the correct one
		switch instance.IR {
		case 0x00E0:
			//Display Clear
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
			sum := Vx + Vy
			if sum > 255 {
				instance.registers[15] = 1
			}
			instance.registers[Vx] = byte(sum)
		case 0x5: //Vx -= Vy
			Vx := instance.IR << 4 >> 12
			Vy := instance.IR << 8 >> 12
			sum := Vx - Vy
			if sum < 0 {
				instance.registers[15] = 1
			}
			instance.registers[Vx] = byte(sum)
		case 0x6:
			
		}
	case 0x9:
		//if Vx j
	//More missing
	case 0xA:
		//I = NNN
	case 0xD:
		//Display
	}
}

//Main function which runs the whole thing
func main() {
	//Opens file and checks for errors
	reader, err := os.Open("IBM Logo.ch8")
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

}
