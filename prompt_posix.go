// +build !windows

package prompt

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	escClearLine  = "\x1B[2K"
	escClearToEnd = "\x1B[0K"
	escMoveUp     = "\x1B[1A"
	escMoveDown   = "\x1B[1B"
	escMoveLeft   = "\x1B[1D"
	escMoveRight  = "\x1B[1C"
	escMoveStart  = "\x1B[G"
	escMoveToRow  = "\x1B[%dH"
	escBold       = "\x1B[1m"
	escRed        = "\x1B[31m"
	escReset      = "\x1B[0m"
	escShow       = "\x1B[?25h"
	escHide       = "\x1B[?25l"
)

func TerminalSize() (int, int, error) {
	data := struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&data))); err != 0 {
		return 0, 0, err
	}
	return int(data.Row), int(data.Col), nil
}

func MakeRawTerminal(hide bool) (func() error, error) {
	if hide {
		fmt.Printf(escHide)
	}
	oldState := syscall.Termios{}
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TCGETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
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

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TCSETS, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		if hide {
			fmt.Printf(escShow)
		}
		return nil, err
	}

	return func() error {
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TCSETS, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
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
