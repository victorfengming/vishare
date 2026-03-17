//go:build windows

package singleinstance

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32    = syscall.NewLazyDLL("kernel32.dll")
	createMutex = kernel32.NewProc("CreateMutexW")
	mutexHandle uintptr
)

// Acquire tries to obtain an exclusive named mutex for the given app name.
// Returns an error if another instance is already running.
func Acquire(name string) error {
	mutexName, err := syscall.UTF16PtrFromString("Global\\Vishare-" + name)
	if err != nil {
		return fmt.Errorf("encode mutex name: %w", err)
	}
	h, _, callErr := createMutex.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))
	if h == 0 {
		return fmt.Errorf("create mutex: %w", callErr)
	}
	if callErr == syscall.ERROR_ALREADY_EXISTS {
		syscall.CloseHandle(syscall.Handle(h))
		return fmt.Errorf("vishare is already running")
	}
	mutexHandle = h
	return nil
}

// Release releases the named mutex.
func Release() {
	if mutexHandle != 0 {
		syscall.CloseHandle(syscall.Handle(mutexHandle))
		mutexHandle = 0
	}
}
