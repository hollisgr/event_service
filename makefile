LOGRUS := github.com/sirupsen/logrus github.com/sirupsen/logrus@v1.9.3
CLEANENV := github.com/ilyakaznacheev/cleanenv
POSTGRESQL := github.com/jackc/pgx github.com/jackc/pgx/v5/pgxpool
GOCRON := github.com/go-co-op/gocron/v2
JWT := github.com/golang-jwt/jwt/v5
VALIDATOR := github.com/go-playground/validator/v10
GIN := github.com/gin-gonic/gin
GOOSE := github.com/pressly/goose/v3/cmd/goose@latest

EXEC_EVENT_LISTENER := event_listener
EXEC_SCHEDULER := scheduler

SRC_EVENT_LISTENER := cmd/event_listener/event_listener.go
SRC_SCHEDULER := cmd/scheduler/scheduler.go

GOOSE_DRIVER := 
GOOSE_DBHOST := 
GOOSE_DBPORT := 
GOOSE_DBNAME := 
GOOSE_DBUSER := 
GOOSE_DBPASSWORD := 
GOOSE_DBSTRING := postgresql://$(GOOSE_DBNAME):$(GOOSE_DBPASSWORD)@$(GOOSE_DBHOST):$(GOOSE_DBPORT)/$(GOOSE_DBNAME)?sslmode=disable


all: clean event_listener scheduler 

event_listener: clean_event_listener build_event_listener

scheduler: clean_scheduler build_scheduler

goose_install:
	export PATH=$(PATH):home/$(USER)/go/bin/
	go install $(GOOSE)

migrations_up:
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) up

migrations_down:
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) up

docker-compose-up-silent: docker-compose-stop
	sudo docker compose -f docker-compose.yml up -d

docker-compose-stop:
	sudo docker compose -f docker-compose.yml stop

docker-compose-up: docker-compose-down
	sudo docker compose -f docker-compose.yml up

docker-compose-down:
	sudo docker compose -f docker-compose.yml down

run_event_listener:
	./$(EXEC_EVENT_LISTENER)

run_scheduler:
	./$(EXEC_SCHEDULER)

build_event_listener:
	go build $(SRC_EVENT_LISTENER)

build_scheduler:
	go build $(SRC_SCHEDULER)

mod:
	go mod init $(EXEC)

get:
	go get $(LOGRUS) $(CLEANENV) $(POSTGRESQL) $(GOCRON) $(JWT) $(VALIDATOR) $(GIN)

clean_event_listener:
	rm -rf $(EXEC_EVENT_LISTENER)

clean_scheduler:
	rm -rf $(EXEC_SCHEDULER)

clean: clean_event_listener clean_scheduler