package core

import (
	"blocker/serialize"
	"blocker/types"
	"fmt"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a // 10
	InstrPushByte Instruction = 0x0b // 11
	InstrAdd      Instruction = 0x0c // 12
	InstrPack     Instruction = 0x0d // 13
	InstrStore    Instruction = 0x0f // 14
	InstrGet      Instruction = 0x0e // 14
)

type VM struct {
	contractState *State // TODO: should change state to interface
	stack         *types.Stack
	data          []byte
	ip            int // Instruction pointer
	sp            int // Stack pointer
}

func NewVM(data []byte, state *State) *VM {
	return &VM{
		data:          data,
		ip:            0,
		stack:         types.NewStack(),
		sp:            -1,
		contractState: state,
	}
}

func (vm *VM) Run() error {
	if len(vm.data) == 0 {
		return nil
	}
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
	case InstrStore:
		var buf any
		val := vm.stack.Pop()
		buf = vm.stack.Pop()
		key, ok := buf.([]byte)
		if !ok {
			panic(fmt.Sprintf("vm: cannot get key as byte array, pack before get, key: (%+v)\n", buf))
		}

		switch v := val.(type) {
		case int:
			buf := serialize.SerializeUint64(uint64(v))
			err := vm.contractState.Put(string(key), buf)
			if err != nil {
				panic(fmt.Sprintf("vm: error while put to state: %v", err))
			}

		default:
			panic("vm: unknow type")
		}
	case InstrGet:
		key := vm.stack.Pop().([]byte)

		val, err := vm.contractState.Get(string(key))
		if err != nil {
			panic(fmt.Sprintf("vm: error while get from state: (%v)", err))
		}
		vm.stack.Push(val)

	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))

	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))

	case InstrAdd:
		i1 := vm.stack.Pop().(int)
		i2 := vm.stack.Pop().(int)
		i3 := i1 + i2
		vm.stack.Push(i3)

	case InstrPack:
		b := make([]byte, vm.stack.Pop().(int))
		for i := range b {
			b[i] = vm.stack.Pop().(byte)
		}

		vm.stack.Push(b)
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
	return v.(byte), nil
}
