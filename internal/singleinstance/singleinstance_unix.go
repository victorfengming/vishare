//go:build linux || darwin

package singleinstance

import (
	"fmt"
	"os"
	"syscall"
)

var (
	lockFd   = -1
	lockPath string
)

// Acquire tries to obtain an exclusive process lock for the given app name.
// Returns an error if another instance is already running.
func Acquire(name string) error {
	lockPath = fmt.Sprintf("/tmp/%s.lock", name)
	fd, err := syscall.Open(lockPath, syscall.O_CREAT|syscall.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open lock file %s: %w", lockPath, err)
	}
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		syscall.Close(fd)
		if err == syscall.EWOULDBLOCK {
			return fmt.Errorf("vishare is already running")
		}
		return fmt.Errorf("acquire lock: %w", err)
	}
	lockFd = fd
	// Overwrite with current PID for debugging convenience
	syscall.Ftruncate(fd, 0)
	syscall.Write(fd, []byte(fmt.Sprintf("%d\n", os.Getpid())))
	return nil
}

// Release releases the lock and removes the lock file.
func Release() {
	if lockFd >= 0 {
		syscall.Flock(lockFd, syscall.LOCK_UN)
		syscall.Close(lockFd)
		os.Remove(lockPath)
		lockFd = -1
	}
}
