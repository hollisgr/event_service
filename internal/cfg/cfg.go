package cfg

import (
	"event_service/pkg/logger"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Cfg struct {
	JwtSecretKey string `env:"JWT_SECRET_KEY"`
	Scheduler    struct {
		Host                  string `env:"S_HOST"`
		Token                 string `env:"S_TOKEN"`
		CheckEventsTimeOutSec int    `env:"CHECK_EVENTS_SEC"`
		SendMessageTimeOutSec int    `env:"SEND_MSG_SEC"`
	} `yaml:"scheduler"`
	Listen struct {
		BindIP string `env:"BIND_IP" env-default:"127.0.0.1"`
		Port   string `env:"LISTEN_PORT" env-default:"8080"`
	} `yaml:"listen"`
	Postgresql struct {
		Host     string `env:"PSQL_HOST"`
		Port     string `env:"PSQL_PORT"`
		Database string `env:"PSQL_NAME"`
		Username string `env:"PSQL_USER"`
		Password string `env:"PSQL_PASSWORD"`
	} `json:"Postgresql"`
}

var instance *Cfg
var once sync.Once

func GetConfig() *Cfg {
	once.Do(func() {
		logger := logger.GetLogger()
		logger.Infoln("read app configuration")
		instance = &Cfg{}
		err := cleanenv.ReadEnv(instance)
		if err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Infoln(help)
			logger.Fatal(err)
		}
	})
	return instance
}
