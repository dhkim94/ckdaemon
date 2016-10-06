package ckdaemon

import (
	"errors"
	"os"
	"syscall"
	"os/signal"
)

type SignalHandlerFunc func(sig os.Signal) (err error)

var (
	ErrStop = errors.New("stop server signals")
	handlers = make(map[os.Signal]SignalHandlerFunc)
)

// 시스템 시그널 발생시 실행 될 handler 를 등록 한다.
// SIGTERM 은 handler를 등록 하지 않더라도 ErrStop 를 기본으로 리턴 한다.
func SetSigHandler(handler SignalHandlerFunc, signals ...os.Signal) {
	for _, sig := range signals {
		handlers[sig] = handler
	}
}

// 시스템 시그널 발생시 등록된 handler 를 실행 한다.
func ServeSignals() (err error) {
	signals := make([]os.Signal, 0, len(handlers))
	for sig, _ := range handlers {
		signals = append(signals, sig)
	}

	ch := make(chan os.Signal, 8)
	signal.Notify(ch, signals...)

	for sig := range ch {
		err = handlers[sig](sig)
		if err != nil {
			break
		}
	}

	signal.Stop(ch)

	if err == ErrStop {
		err = nil
	}

	return
}

func sigtermDefaultHandler(sig os.Signal) error {
	return ErrStop
}

func init() {
	handlers[syscall.SIGTERM] = sigtermDefaultHandler
}
