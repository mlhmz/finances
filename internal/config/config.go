package config

type Config struct {
	DBPath string
	Port   string
}

func Default() Config {
	return Config{
		DBPath: "finances.db",
		Port:   ":3000",
	}
}
