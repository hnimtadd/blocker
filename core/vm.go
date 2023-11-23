package core

import (
	"blocker/types"
	"fmt"
)

type Instruction byte

const (
	InstrPush Instruction = 0x0a // 10
	InstrAdd  Instruction = 0x0b // 11
)

type VM struct {
	data  []byte
	ip    int // Instruction pointer
	stack *types.Stack[byte]
	sp    int // Stack pointer
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: types.NewStack[byte](),
		sp:    -1,
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])
		if err := vm.ExecInstruction(instr); err != nil {
			return err
		}
		vm.ip++
		if vm.ip > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) ExecInstruction(instr Instruction) error {
	switch instr {
	case InstrPush:
		vm.pushStack(vm.data[vm.ip-1])
	case InstrAdd:
		i1 := vm.stack.Pop()
		i2 := vm.stack.Pop()
		i3 := i1 + i2
		vm.pushStack(i3)
		// default:
		// 	fmt.Println(instr)
	}
	return nil
}
func (vm *VM) pushStack(b byte) {
	// vm.sp++
	vm.stack.Push(b)
	// vm.stack[vm.sp] = b
}

func (vm *VM) done() (byte, error) {
	v := vm.stack.Pop()
	if vm.stack.Len() != 0 {
		return 0, fmt.Errorf("stack not nil")
	}
	return v, nil
}
