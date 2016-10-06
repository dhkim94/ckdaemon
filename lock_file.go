package ckdaemon

import (
	"os"
	"fmt"
	"errors"
	"syscall"
)

var EWouldBlock = errors.New("File already locked another process")

// 이러면 기본으로 제공하는 os.File 에 원하는 function 을 연결 할 수 있다.
// 이 방법 좋네.
type LockFile struct {
	*os.File
}

// 파라미터로 주어진 파일로 부터 파일 Lock을 생성 한다.
func NewLockFile(file *os.File) *LockFile {
	return &LockFile{file}
}

// 잠긴 파일을 오픈 한다.
func OpenLockFile(name string, perm os.FileMode) (lock *LockFile, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm); err == nil {
		lock = NewLockFile(file)
	}

	return
}

// 주어진 이름의 pid file 을 생성 하고 프로세스 아이디를 적는다.
func CreatePidFile(name string, perm os.FileMode) (lock *LockFile, err error) {
	if lock, err = OpenLockFile(name, perm); err != nil {
		return
	}

	if err = lock.Lock(); err != nil {
		lock.Remove()
		return
	}

	if err = lock.WritePid(); err != nil {
		lock.Remove()
	}

	return
}

// 파일을 Unlock 하고 close 한다.
func (file *LockFile) Remove() error {
	defer file.Close()

	if err := file.Unlock(); err != nil {
		return err
	}

	name, err := GetFdName(file.Fd())
	if err != nil {
		return err
	}

	err = syscall.Unlink(name)

	return err
}

// 현재 프로세스의 pid 를 파일에 적는다.
func (file *LockFile) WritePid() (err error) {
	if _, err = file.Seek(0, os.SEEK_SET); err != nil {
		return
	}

	var fileLen int
	if fileLen, err = fmt.Fprint(file, os.Getpid()); err != nil {
		return
	}

	if err = file.Truncate(int64(fileLen)); err != nil {
		return
	}

	err = file.Sync()
	return
}

// pid 파일에서 pid 를 구한다.
func ReadPidFile(name string) (pid int, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDONLY, 0640); err != nil {
		return
	}
	defer file.Close()

	lock := NewLockFile(file)
	pid, err = lock.ReadPid()

	return
}

// Pid 를 저장하고 있는 파일에서 pid 를 읽는다.
func (file *LockFile) ReadPid() (pid int, err error) {
	if _, err = file.Seek(0, os.SEEK_SET); err != nil {
		return
	}

	_, err = fmt.Fscan(file, &pid)

	return
}

// 파일을 lock 한다.
func (file *LockFile) Lock() error {
	return lockFile(file.Fd())
}

// 파일을 unlock 한다.
func (file *LockFile) Unlock() error {
	return unlockFile(file.Fd())
}

// 파일 descriptor 를 이용하여 파일명을 구한다.
func GetFdName(fd uintptr) (name string, err error) {
	return getFdName(fd)
}
