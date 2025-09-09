package cfg

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Cfg struct {
	Token     string `env:"S_TOKEN"`
	Scheduler struct {
		Host                  string `env:"S_HOST"`
		CheckEventsTimeOutSec int    `env:"CHECK_EVENTS_SEC"`
		SendMessageTimeOutSec int    `env:"SEND_MSG_SEC"`
	}
	Listen struct {
		BindIP string `env:"BIND_IP" env-default:"0.0.0.0"`
		Port   string `env:"LISTEN_PORT" env-default:"8080"`
	}
	Postgresql struct {
		Host     string `env:"PSQL_HOST"`
		Port     string `env:"PSQL_PORT"`
		Database string `env:"PSQL_NAME"`
		Username string `env:"PSQL_USER"`
		Password string `env:"PSQL_PASSWORD"`
	}
}

var instance *Cfg
var once sync.Once

func GetConfig() *Cfg {
	once.Do(func() {
		log.Println("read app configuration")
		instance = &Cfg{}
		err := cleanenv.ReadConfig(".env", instance)
		if err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatalln(err)
		}
	})
	return instance
}
