package env_os

import "os"

var env = "dev"

func SetEnv(e string) {
	if e == "" {
		env = "dev"
	}
	env = e
}

func Env() string {
	return env
}

func IsDev() bool {
	return os.Getenv("DEV") == "true"
}
