package config

// Limits - сколько попыток в минуту пропускается на один логин, пароль и IP.
// Границы проверяет домен.
type Limits struct {
	Login    int `env:"LIMIT_LOGIN"    env-default:"10"`
	Password int `env:"LIMIT_PASSWORD" env-default:"100"`
	IP       int `env:"LIMIT_IP"       env-default:"1000"`
}
