package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"

	"github.com/syscrypt/boring-calendar/internal/handler"
	"github.com/syscrypt/boring-calendar/internal/model"
	"github.com/syscrypt/boring-calendar/internal/service"
)

const (
	DB_PATH = "./build/calendar.db"
)

func main() {
	config := &model.Config{}
	configPtr := flag.String("config", "./init/config.json", "path to the json config file")
	flag.Parse()
	if configJson, err := os.ReadFile(*configPtr); err == nil {
		if err = json.Unmarshal(configJson, config); err != nil {
			logrus.WithError(err).Fatalln("loaded config.json")
		}
	}
	logrus.Infoln("loaded config.json")

	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		logrus.WithError(err).Fatalf("unable to open db file %s", DB_PATH)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		logrus.WithError(err).Fatalln("error while validating database connection")
	}

	ctx := context.Background()
	calendarService, err := service.NewCalendarService(db, ctx, logrus.StandardLogger())
	if err != nil {
		logrus.WithError(err).Fatalln("error creating calendar service")
	}
	defer calendarService.Close()

	if exists, err := calendarService.UserExists(config.TestUserToken, ctx); !exists && err == nil {
		err = calendarService.CreateUser(config.TestUserToken, ctx)
		if err != nil {
			logrus.WithError(err).Fatalln("error creating test user")
		}
	} else if err != nil {
		logrus.WithError(err).Fatalln("error creating test user")
	}

	logrus.Infof("starting server on port %s", config.Port)
	http.ListenAndServe(":"+config.Port, handler.NewHandler(db, calendarService))
}
