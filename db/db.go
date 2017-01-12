package db

type (
	// DB defines the interface for general database operations.
	DB interface {
		Logger
	}

	// Logger defines the interface for logger middleware.
	Logger interface {
		Log(*Request) error
	}

	// Mongo implements `DB`
	Mongo struct{}
)
