package sys

import (
	"syscall"
)

// Prctl exposes the prctl syscall which enables various operations on a process.
//
// Reference:
// http://man7.org/linux/man-pages/man2/prctl.2.html
func Prctl(option int, arg2, arg3, arg4, arg5 uintptr) error {
	_, _, e1 := syscall.Syscall6(syscall.SYS_PRCTL, uintptr(option), arg2, arg3, arg4, arg5, 0)
	if e1 != 0 {
		return fmt.Errorf("errno %d", e1)
	}
	return nil
}
