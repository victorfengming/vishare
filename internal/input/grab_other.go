//go:build !linux

package input

func GrabLocalInput() error {
	return nil
}

func ReleaseLocalInput() {}
