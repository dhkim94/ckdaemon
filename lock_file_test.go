package ckdaemon

import (
	"testing"
	"fmt"
	"os"
	"io/ioutil"
	"runtime"
	"os/exec"
	"io"
	"strings"
)

var (
	filename string		= os.TempDir() + "test.lock"
	fileperm os.FileMode	= 0644
	invalidname string	= "/a/b/c/d"
)

type script struct {
	cmd *exec.Cmd
	stdout io.ReadCloser
	stdin io.WriteCloser
}

func setdata() {
	if strings.Index(runtime.GOOS, "darwin") != -1 {
		filename = os.TempDir() + "test.lock"
	} else {
		filename = os.TempDir() + "/test.lock"
	}
}

func TestCreatePidFile(t *testing.T) {
	setdata()

	if _, err := CreatePidFile(invalidname, fileperm); err == nil {
		t.Fatal("err1: ", err)
	}

	fmt.Println("> filename [", filename, "]")

	lock, err := CreatePidFile(filename, fileperm)
	if err != nil {
		t.Fatal("err2: ", err)
	}
	defer lock.Remove()

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal("err3: ", err)
	}

	if string(data) != fmt.Sprint(os.Getpid()) {
		t.Fatal("err4: pid not equal")
	}

	file, err := os.OpenFile(filename, os.O_RDONLY, fileperm)
	if err != nil {
		t.Fatal("err5: ", err)
	}

	if err = NewLockFile(file).WritePid(); err == nil {
		t.Fatal("err6: ", err)
	}
}

func TestNewLockFile(t *testing.T) {
	setdata()

	lock := NewLockFile(os.NewFile(1001, ""))
	err := lock.Remove()
	if err == nil {
		t.Fatal("case1: ", err)
	}

	err = lock.WritePid()
	if err == nil {
		t.Fatal("case2: ", err)
	}
}

func TestGetFdName(t *testing.T) {
	setdata()

	name, err := GetFdName(0)
	if err != nil {
		t.Error(err)
	} else {
		if name != "/dev/null" {
			t.Errorf("Filename of fd 0: [%s]", name)
		}
	}

	name, err = GetFdName(1011)
	if err == nil {
		t.Errorf("GetFdName(): detected invalid fd. name [%s]", name)
	}
}

func TestReadPidFile(t *testing.T) {
	setdata()

	lock, err := CreatePidFile(filename, fileperm)
	if err != nil {
		t.Fatal("case1: ", err)
	}
	defer lock.Remove()

	pid, err := ReadPidFile(filename)
	if err != nil {
		t.Fatal("case2: ", err)
	}

	if pid != os.Getpid() {
		t.Fatal("case3: pid not equal")
	}
}

// OSX 에서는 flock 실행 파일이 없어서 스크립트로 파일 lock 을 걸수 없다.
// 때문에 OSX 에서는 테스트 하지 않는다.
func TestLockFileLock(t *testing.T) {
	if strings.Index(runtime.GOOS, "darwin") != -1 {
		fmt.Println("ignore test. no flock on osx.")
		return
	}


	lock, err := OpenLockFile(filename, fileperm)
	if err != nil {
		t.Fatal("case1: ", err)
	}
	defer lock.Remove()

	if err = lock.Lock(); err != nil {
		t.Fatal("case2: ", err)
	}

	scr, msg, err := createLockScriptAndStart()
	if err != nil {
		t.Fatal("case3: ", err)
	}
	if msg != "error" {
		t.Fatal("script was able lock file")
	}
	if err = terminateLockScript(scr); err != nil {
		t.Fatal(err)
	}

	if err = lock.Unlock(); err != nil {
		t.Fatal(err)
	}
	lock.Close()

	scr, msg, err = createLockScriptAndStart()
	if err != nil {
		t.Fatal(err)
	}
	if msg != "locked" {
		t.Fatal("script can not lock file")
	}
	lock, err = CreatePidFile(filename, fileperm)
	if err != EWouldBlock {
		t.Fatal("Lock() not work properly")
	}
	if err = terminateLockScript(scr); err != nil {
		t.Error(err)
	}
}

func createLockScriptAndStart() (scr *script, msg string, err error) {
	var text = fmt.Sprintf(`
		set -e
		exec 222>"%s"
		flock -n 222||echo "error"
		echo "locked"
		read inp`, filename)

	scr, err = createScript(text, true)
	if err != nil {
		return
	}

	if err = scr.cmd.Start(); err != nil {
		return
	}
	// wait until the script does not try to lock the file
	msg, err = scr.get()

	return
}

func createScript(text string, createPipes bool) (scr *script, err error) {
	var scrName string
	if scrName, err = createScriptFile(text); err != nil {
		return
	}

	scr = &script{cmd: exec.Command("bash", scrName)}
	if createPipes {
		if scr.stdout, err = scr.cmd.StdoutPipe(); err != nil {
			return
		}
		if scr.stdin, err = scr.cmd.StdinPipe(); err != nil {
			return
		}
	}

	return
}

func createScriptFile(text string) (name string, err error) {
	var scr *os.File
	if scr, err = ioutil.TempFile(os.TempDir(), "scr"); err != nil {
		return
	}
	defer scr.Close()

	name = scr.Name()
	_, err = scr.WriteString(text)

	return
}

func terminateLockScript(scr *script) (err error) {
	if err = scr.send(""); err != nil {
		return
	}
	err = scr.cmd.Wait()

	return
}

func (scr *script) get() (line string, err error) {
	_, err = fmt.Fscanln(scr.stdout, &line)
	return
}

func (scr *script) send(line string) (err error) {
	_, err = fmt.Fprintln(scr.stdin, line)
	return
}
