// +build !windows

package prompt

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	escClearLine = "\x1B[2K"
	escMoveUp    = "\x1B[1A"
	escMoveDown  = "\x1B[1B"
	escMoveLeft  = "\x1B[1D"
	escMoveRight = "\x1B[1C"
	escMoveStart = "\x1B[G"
	escBold      = "\x1B[1m"
	escRed       = "\x1B[31m"
	escReset     = "\x1B[0m"
	escShow      = "\x1B[?25h"
	escHide      = "\x1B[?25l"
)

func MakeRaw(hide bool) (func() error, error) {
	if hide {
		fmt.Printf(escHide)
	}
	oldState := syscall.Termios{}
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), syscall.TCGETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		if hide {
			fmt.Printf(escShow)
		}
		return nil, err
	}

	newState := syscall.Termios{}
	newState.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG
	// Because we are clearing canonical mode, we need to ensure VMIN & VTIME are
	// set to the values we expect. This combination puts things in standard
	// "blocking read" mode (see termios(3)).
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), syscall.TCSETS, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		if hide {
			fmt.Printf(escShow)
		}
		return nil, err
	}

	return func() error {
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), syscall.TCSETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
			if hide {
				fmt.Printf(escShow)
			}
			return err
		}
		if hide {
			fmt.Printf(escShow)
		}
		return nil
	}, nil
}
