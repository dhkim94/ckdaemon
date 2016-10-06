package ckdaemon

import (
	"syscall"
	"unsafe"
	"bytes"
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
	_path := make([]byte, 256)
	// 가장 마지막에 char* 가 들어가야 하기 때문에 slice 를 char* 로 변경하는 방법은 아래와 같다.
	// 이거....찾기 무지 힘들었는데, 그냥 cgo 보고 연상할 것을 왜 생각 못 했는지...
	if _, _, _errno := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), syscall.F_GETPATH, uintptr(unsafe.Pointer(&_path[0]))); _errno != 0 {
		err = _errno
		return
	}

	name = string(bytes.Trim(_path, "\x00"))

	return
}