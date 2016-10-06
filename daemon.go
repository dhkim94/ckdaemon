package ckdaemon

import (
	"errors"
	"os"
	"syscall"
)

var errNotSupported = errors.New("daemon: non posix OS in not supported")

// 프로세스 마크 - 시스템 환경 변수 _GO_DAEMON=1 으로 설정 된다.
const (
	MARK_NAME = "_GO_DEAMON"
	MARK_VALUE = "1"
)

// log 및 pid 파일의 기본 파일 퍼미션
const FILE_PERM = os.FileMode(0644)

// 데몬 프로세스의 context
type Context struct {
	// PidFileName 이 empty 가 아니면 parent process 는 주어진 이름으로 파일을 생성하고 lock 한다.
	// child process 는 파일에 pid 를 적는다.
	PidFileName string
	// pid 파일의 파일 퍼미션
	PidFilePerm os.FileMode

	// LogFileName 이 empty 가 아니면 parent process 는 주어진 이름으로 파일을 생성하고
	// child process 를 위해서 fd 1(stdout) 에 링크를 건다.
	LogFileName string
	// 로그 파일의 파일 퍼미션
	LogFilePerm os.FileMode

	// WorkDir 가 empty 가 아니면 child process 는 working 디렉토리를 주어진 경로로 변경 한다.
	WorkDir string
	// Chroot 가 emptry 가 아니면 child process 는 루트 디렉토리를 주어진 경로로 변경 한다.
	Chroot string

	// Env 가 nil 이 아니면 데몬 프로세스를 위한 os.Environ 형태의 환경 변수를 생성 한다.
	Env []string
	// Args 가 nil 이면 커맨드 라인으로 받은 명령어를 포함하는 os.Args 를 리턴 한다.
	Args []string

	// Credential 은 데몬 프로세스의 user, group 아이디를 포함한다.
	Credential *syscall.Credential
	// Umask 가 0 가 아니면 데몬 프로세스는 주어진 값으로 umask 를 실행 한다.
	Umask int

	abspath  string
	pidFile  *LockFile
	stdoutFile *os.File
	logFile  *os.File
	nullFile *os.File

	rpipe, wpipe *os.File
}

// child 프로세스에서는 true 를 리턴 하고 parent 프로세스에서는 false 를 리턴 한다.
func WasReborn() bool {
	return os.Getenv(MARK_NAME) == MARK_VALUE
}

// 주어진 Context 로 현재 프로세스의 카피를 만든다. fork 와 비슷하게 데몬 프로세스를 만들지만 goroutine safe 하다.
// parent 프로세스에서 호출하면 child 프로세스를 리턴 하고
// child 프로세스에서 호출하면 nil 을 리턴 하며
// 그 외의 경우는 error 를 리턴 한다.
func (d *Context) Reborn() (child *os.Process, err error) {
	return d.reborn()
}

// Context 에 저장되어 있는 pid 파일에 포함된 pid 를 이용하여 데몬 프로세스를 구한다.
// 만일 pid 파일명이 empty 라면 nil 을 리턴 한다.
func (d *Context) Search() (daemon *os.Process, err error) {
	return d.search()
}

// pid 파일을 release 한다.
func (d *Context) Release() (err error) {
	return d.release()
}
