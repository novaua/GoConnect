package serverimpl

// Config is a configuration properties
type Config struct {
	HttpsPort     int
	MaxConnection int
}

// NewConfig creates default configuration
func NewConfig(secured bool) Config {
	defaultPort := 7002
	if secured {
		defaultPort = 7443
	}

	return Config{
		HttpsPort:     defaultPort,
		MaxConnection: 128,
	}
}
