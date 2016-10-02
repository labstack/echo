package gracenet

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/freeport"
)

func TestEmptyCountEnvVariable(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	ensure.Nil(t, n.inherit())
}

func TestZeroCountEnvVariable(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "0")
	ensure.Nil(t, n.inherit())
}

func TestInvalidCountEnvVariable(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "a")
	expected := regexp.MustCompile("^found invalid count value: LISTEN_FDS=a$")
	ensure.Err(t, n.inherit(), expected)
}

func TestInvalidFileInherit(t *testing.T) {
	var n Net
	tmpfile, err := ioutil.TempFile("", "TestInvalidFileInherit-")
	ensure.Nil(t, err)
	defer os.Remove(tmpfile.Name())
	n.fdStart = dup(t, int(tmpfile.Fd()))
	os.Setenv(envCountKey, "1")
	ensure.Err(t, n.inherit(), regexp.MustCompile("^error inheriting socket fd"))
	ensure.DeepEqual(t, len(n.inherited), 0)
	ensure.Nil(t, tmpfile.Close())
}

func TestInheritErrorOnListenTCPWithInvalidCount(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "a")
	_, err := n.Listen("tcp", ":0")
	ensure.NotNil(t, err)
}

func TestInheritErrorOnListenUnixWithInvalidCount(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "a")
	tmpdir, err := ioutil.TempDir("", "TestInheritErrorOnListenUnixWithInvalidCount-")
	ensure.Nil(t, err)
	ensure.Nil(t, os.RemoveAll(tmpdir))
	_, err = n.Listen("unix", filepath.Join(tmpdir, "socket"))
	ensure.NotNil(t, err)
}

func TestOneTcpInherit(t *testing.T) {
	var n Net
	l, err := net.Listen("tcp", ":0")
	ensure.Nil(t, err)
	file, err := l.(*net.TCPListener).File()
	ensure.Nil(t, err)
	ensure.Nil(t, l.Close())
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l, err = n.Listen("tcp", l.Addr().String())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	active, err := n.activeListeners()
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(active), 1)
	ensure.Nil(t, l.Close())
}

func TestSecondTcpListen(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	l, err := n.Listen("tcp", ":0")
	ensure.Nil(t, err)
	_, err = n.Listen("tcp", l.Addr().String())
	ensure.Err(t, err, regexp.MustCompile("bind: address already in use$"))
	ensure.Nil(t, l.Close())
}

func TestSecondTcpListenInherited(t *testing.T) {
	var n Net
	l, err := net.Listen("tcp", ":0")
	ensure.Nil(t, err)
	file, err := l.(*net.TCPListener).File()
	ensure.Nil(t, err)
	ensure.Nil(t, l.Close())
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l, err = n.Listen("tcp", l.Addr().String())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	_, err = n.Listen("tcp", l.Addr().String())
	ensure.Err(t, err, regexp.MustCompile("bind: address already in use$"))
	ensure.Nil(t, l.Close())
}

func TestInvalidNetwork(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	_, err := n.Listen("foo", "")
	ensure.Err(t, err, regexp.MustCompile("^unknown network foo$"))
}

func TestInvalidNetworkUnix(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	_, err := n.Listen("invalid_unix_net_for_test", "")
	ensure.Err(t, err, regexp.MustCompile("^unknown network invalid_unix_net_for_test$"))
}

func TestWithTcp0000(t *testing.T) {
	var n Net
	port, err := freeport.Get()
	ensure.Nil(t, err)
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	l, err := net.Listen("tcp", addr)
	ensure.Nil(t, err)
	file, err := l.(*net.TCPListener).File()
	ensure.Nil(t, err)
	ensure.Nil(t, l.Close())
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l, err = n.Listen("tcp", addr)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	ensure.Nil(t, l.Close())
}

func TestWithTcpIPv6Loal(t *testing.T) {
	var n Net
	l, err := net.Listen("tcp", "[::]:0")
	ensure.Nil(t, err)
	file, err := l.(*net.TCPListener).File()
	ensure.Nil(t, err)
	ensure.Nil(t, l.Close())
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l, err = n.Listen("tcp", l.Addr().String())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	ensure.Nil(t, l.Close())
}

func TestOneUnixInherit(t *testing.T) {
	var n Net
	tmpfile, err := ioutil.TempFile("", "TestOneUnixInherit-")
	ensure.Nil(t, err)
	ensure.Nil(t, tmpfile.Close())
	ensure.Nil(t, os.Remove(tmpfile.Name()))
	defer os.Remove(tmpfile.Name())
	l, err := net.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	file, err := l.(*net.UnixListener).File()
	ensure.Nil(t, err)
	ensure.Nil(t, l.Close())
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l, err = n.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	ensure.Nil(t, l.Close())
}

func TestInvalidTcpAddr(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	_, err := n.Listen("tcp", "abc")
	ensure.Err(t, err, regexp.MustCompile("^missing port in address abc$"))
}

func TestTwoTCP(t *testing.T) {
	var n Net

	port1, err := freeport.Get()
	ensure.Nil(t, err)
	addr1 := fmt.Sprintf(":%d", port1)
	l1, err := net.Listen("tcp", addr1)
	ensure.Nil(t, err)

	port2, err := freeport.Get()
	ensure.Nil(t, err)
	addr2 := fmt.Sprintf(":%d", port2)
	l2, err := net.Listen("tcp", addr2)
	ensure.Nil(t, err)

	file1, err := l1.(*net.TCPListener).File()
	ensure.Nil(t, err)
	file2, err := l2.(*net.TCPListener).File()
	ensure.Nil(t, err)

	// assign both to prevent GC from kicking in the finalizer
	fds := []int{dup(t, int(file1.Fd())), dup(t, int(file2.Fd()))}
	n.fdStart = fds[0]
	os.Setenv(envCountKey, "2")

	// Close these after to ensure we get coalaced file descriptors.
	ensure.Nil(t, l1.Close())
	ensure.Nil(t, l2.Close())

	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 2)

	l1, err = n.Listen("tcp", addr1)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	ensure.Nil(t, l1.Close())
	ensure.Nil(t, file1.Close())

	l2, err = n.Listen("tcp", addr2)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 2)
	ensure.DeepEqual(t, n.inherited[1], nil)
	ensure.Nil(t, l2.Close())
	ensure.Nil(t, file2.Close())
}

func TestOneUnixAndOneTcpInherit(t *testing.T) {
	var n Net

	tmpfile, err := ioutil.TempFile("", "TestOneUnixAndOneTcpInherit-")
	ensure.Nil(t, err)
	ensure.Nil(t, tmpfile.Close())
	ensure.Nil(t, os.Remove(tmpfile.Name()))
	defer os.Remove(tmpfile.Name())
	unixL, err := net.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)

	port, err := freeport.Get()
	ensure.Nil(t, err)
	addr := fmt.Sprintf(":%d", port)
	tcpL, err := net.Listen("tcp", addr)
	ensure.Nil(t, err)

	tcpF, err := tcpL.(*net.TCPListener).File()
	ensure.Nil(t, err)
	unixF, err := unixL.(*net.UnixListener).File()
	ensure.Nil(t, err)

	// assign both to prevent GC from kicking in the finalizer
	fds := []int{dup(t, int(tcpF.Fd())), dup(t, int(unixF.Fd()))}
	n.fdStart = fds[0]
	os.Setenv(envCountKey, "2")

	// Close these after to ensure we get coalaced file descriptors.
	ensure.Nil(t, tcpL.Close())
	ensure.Nil(t, unixL.Close())

	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 2)

	unixL, err = n.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[1], nil)
	ensure.Nil(t, unixL.Close())
	ensure.Nil(t, unixF.Close())

	tcpL, err = n.Listen("tcp", addr)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 2)
	ensure.DeepEqual(t, n.inherited[0], nil)
	ensure.Nil(t, tcpL.Close())
	ensure.Nil(t, tcpF.Close())
}

func TestSecondUnixListen(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	tmpfile, err := ioutil.TempFile("", "TestSecondUnixListen-")
	ensure.Nil(t, err)
	ensure.Nil(t, tmpfile.Close())
	ensure.Nil(t, os.Remove(tmpfile.Name()))
	defer os.Remove(tmpfile.Name())
	l, err := n.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	_, err = n.Listen("unix", tmpfile.Name())
	ensure.Err(t, err, regexp.MustCompile("bind: address already in use$"))
	ensure.Nil(t, l.Close())
}

func TestSecondUnixListenInherited(t *testing.T) {
	var n Net
	tmpfile, err := ioutil.TempFile("", "TestSecondUnixListenInherited-")
	ensure.Nil(t, err)
	ensure.Nil(t, tmpfile.Close())
	ensure.Nil(t, os.Remove(tmpfile.Name()))
	defer os.Remove(tmpfile.Name())
	l1, err := net.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	file, err := l1.(*net.UnixListener).File()
	ensure.Nil(t, err)
	n.fdStart = dup(t, int(file.Fd()))
	ensure.Nil(t, file.Close())
	os.Setenv(envCountKey, "1")
	ensure.Nil(t, n.inherit())
	ensure.DeepEqual(t, len(n.inherited), 1)
	l2, err := n.Listen("unix", tmpfile.Name())
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(n.active), 1)
	ensure.DeepEqual(t, n.inherited[0], nil)
	_, err = n.Listen("unix", tmpfile.Name())
	ensure.Err(t, err, regexp.MustCompile("bind: address already in use$"))
	ensure.Nil(t, l1.Close())
	ensure.Nil(t, l2.Close())
}

func TestPortZeroTwice(t *testing.T) {
	var n Net
	os.Setenv(envCountKey, "")
	l1, err := n.Listen("tcp", ":0")
	ensure.Nil(t, err)
	l2, err := n.Listen("tcp", ":0")
	ensure.Nil(t, err)
	ensure.Nil(t, l1.Close())
	ensure.Nil(t, l2.Close())
}

// We dup file descriptors because the os.Files are closed by a finalizer when
// they are GCed, which interacts badly with the fact that the OS reuses fds,
// and that we emulating inheriting the fd by it's integer value in our tests.
func dup(t *testing.T, fd int) int {
	nfd, err := syscall.Dup(fd)
	ensure.Nil(t, err)
	return nfd
}
