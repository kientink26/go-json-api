package config

type Config struct {
	Port int
	Env  string
	Db   struct {
		Dsn string
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
	Cors struct {
		TrustedOrigin string
	}
}
