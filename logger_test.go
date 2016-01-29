package echo

import (
	"bytes"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
)

func TestGommonLogger(t *testing.T) {
	assert.Implements(t, (*Logger)(nil), &GommonLogger{})

	buf := new(bytes.Buffer)
	glogger := log.New("echo")
	glogger.SetLevel(log.DEBUG)
	logger := &GommonLogger{logger: glogger}

	switch os.Getenv("TESm_LOGGER_FATAL") {
	case "fatal":
		loggerFatal(logger, buf)
		return
	case "fatalf":
		loggerFatalf(logger, buf)
		return
	}

	glogger.SetOutput(buf)

	logger.Debug("m_debug")
	assert.Contains(t, buf.String(), "m_debug")
	buf.Reset()

	logger.Debugf("%s_debugf", "m")
	assert.Contains(t, buf.String(), fmt.Sprintf("%s_debugf", "m"))
	buf.Reset()

	logger.Info("m_info")
	assert.Contains(t, buf.String(), "m_info")
	buf.Reset()

	logger.Infof("%s_infof", "m")
	assert.Contains(t, buf.String(), fmt.Sprintf("%s_infof", "m"))
	buf.Reset()

	logger.Warn("m_warning")
	assert.Contains(t, buf.String(), "m_warning")
	buf.Reset()

	logger.Warnf("%s_warningf", "m")
	assert.Contains(t, buf.String(), fmt.Sprintf("%s_warningf", "m"))
	buf.Reset()

	logger.Error("m_error")
	assert.Contains(t, buf.String(), "m_error")
	buf.Reset()

	logger.Errorf("%s_errorf", "m")
	assert.Contains(t, buf.String(), fmt.Sprintf("%s_errorf", "m"))
	buf.Reset()

	loggerFatalTest(t, logger, buf, "fatal", "m_fatal")
	loggerFatalTest(t, logger, buf, "fatalf", "m_fatalf")
}

func loggerFatalTest(t *testing.T, logger Logger, buf *bytes.Buffer, env string, contains string) {
	cmd := exec.Command(os.Args[0], "-test.run=TestGommonLogger")
	cmd.Env = append(os.Environ(), "TESm_LOGGER_FATAL="+env)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		assert.Contains(t, buf.String(), contains)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func loggerFatal(logger Logger, buf *bytes.Buffer) {
	logger.Fatal("m_fatal")
	buf.Reset()
}

func loggerFatalf(logger Logger, buf *bytes.Buffer) {
	logger.Fatalf("%s_fatalf", "m")
	buf.Reset()
}
