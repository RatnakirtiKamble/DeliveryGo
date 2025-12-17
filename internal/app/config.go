package app 

type Config struct {
	HTTPAddr 	string
	PostgresDSN	string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr: ":8000",
		PostgresDSN: ,
	}
}