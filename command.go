package ckdaemon

import "os"

// 말 그대로 daemon 이 받은 플래그에 대한 set, unset 구분을 위한 interface
type Flag interface {
	IsSet() bool
}

type boolFlag struct {
	b *bool
}

type stringFlag struct {
	s* string
	v string
}

var flags = make(map[Flag]os.Signal)

// f 가 true 일때 set 상태
func BoolFlag(f *bool) Flag {
	return &boolFlag{f}
}

// f 의 값과 v 가 같을때 set 상태
func StringFlag(f *string, v string) Flag {
	return &stringFlag{f, v}
}

// 저장되어 있는 모든 flag 를 구한다.
func Flags() map[Flag]os.Signal {
	return flags
}

// signal 이 지정된 flag 를 저장 한다.
func AddFlag(f Flag, sig os.Signal) {
	flags[f] = sig
}

// active(set) 된 flag 를 구한다.
func ActiveFlags() (ret []Flag) {
	ret = make([]Flag, 0, 1)
	for f := range flags {
		if f.IsSet() {
			ret = append(ret, f)
		}
	}
	return
}

// flag 를 저장하고 flag 에 맞는 signal handler 를 매핑 한다.
func AddCommand(f Flag, sig os.Signal, handler SignalHandlerFunc) {
	if f != nil {
		AddFlag(f, sig)
	}

	if handler != nil {
		SetSigHandler(handler, sig)
	}
}

// 프로세스에 저장되어 있는 시그널을 전송 한다.
func SendCommands(p *os.Process) (err error) {
	for _, sig := range signals() {
		if err = p.Signal(sig); err != nil {

			// todo signal 전송 실패 로그를 찍자.

			return
		}
	}
	return
}

func (f *boolFlag) IsSet() bool {
	if f == nil {
		return false
	}

	return *f.b
}

func (f *stringFlag) IsSet() bool {
	if f == nil {
		return false
	}

	return *f.s == f.v
}

// set 되어 있는 signal 을 구한다.
func signals() (ret []os.Signal) {
	ret = make([]os.Signal, 0, 1)
	for f, sig := range flags {
		if f.IsSet() {
			ret = append(ret, sig)
		}
	}

	return
}
