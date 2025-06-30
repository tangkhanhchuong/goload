package configs

import "time"

type Hash struct {
	Cost int `yaml:"cost"`
}

type Token struct {
	ExpiresIn string `yaml:"expires_in"`
}

func (t Token) GetExpiresInDuration() (time.Duration, error) {
	return time.ParseDuration(t.ExpiresIn)
}

type Auth struct {
	Hash  Hash
	Token Token
}
