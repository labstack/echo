package logger

type (
	// Logger is the interface that declares Echo's logging system.
	Logger interface {
		Debug(...interface{})
		Debugf(string, ...interface{})

		Info(...interface{})
		Infof(string, ...interface{})

		Warn(...interface{})
		Warnf(string, ...interface{})

		Error(...interface{})
		Errorf(string, ...interface{})

		Fatal(...interface{})
		Fatalf(string, ...interface{})
	}
)
