package ckdaemon

import (
	"syscall"
	"fmt"
	"os"
)

// Non-block 파일 쓰기 Lock
func lockFile(fd uintptr) error {
	err := syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		err = EWouldBlock
	}
	return err
}

func unlockFile(fd uintptr) error {
	err := syscall.Flock(int(fd), syscall.LOCK_UN)
	if err == syscall.EWOULDBLOCK {
		err = EWouldBlock
	}
	return err
}

func getFdName(fd uintptr) (name string, err error) {
	path := fmt.Sprintf("/proc/self/fd/%d", int(fd))

	var (
		fileInfo os.FileInfo
		n int
	)

	if fileInfo, err = os.Lstat(path); err != nil {
		return
	}

	buff := make([]byte, fileInfo.Size() + 1)

	if n, err = syscall.Readlink(path, buff); err == nil {
		name = string(buff[:n])
	}

	return
}
