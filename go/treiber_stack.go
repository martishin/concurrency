package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type item[T any] struct {
	value T
	next  unsafe.Pointer
}

type Stack[T any] struct {
	head unsafe.Pointer
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(value T) {
	node := &item[T]{value: value}

	for {
		head := atomic.LoadPointer(&s.head)
		node.next = head

		if atomic.CompareAndSwapPointer(&s.head, head, unsafe.Pointer(node)) {
			return
		}
	}
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T

	for {
		head := atomic.LoadPointer(&s.head)
		if head == nil {
			return zero, false
		}

		next := atomic.LoadPointer(&(*item[T])(head).next)
		if atomic.CompareAndSwapPointer(&s.head, head, next) {
			return (*item[T])(head).value, true
		}
	}
}

func main() {
	stack := NewStack[int]()

	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 50; i++ {
		go func(value int) {
			defer wg.Done()
			stack.Push(value)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			val, ok := stack.Pop()
			fmt.Println(val, ok)
		}()
	}

	wg.Wait()

	v, ok := stack.Pop()
	if ok {
		fmt.Println("Popped:", v)
	} else {
		fmt.Println("Stack is empty")
	}
}
