package application

import (
	"fmt"
	"github.com/kientink26/go-json-api/cmd/api/config"
	"github.com/kientink26/go-json-api/internal/data"
	"github.com/kientink26/go-json-api/internal/mailer"
	"log"
)

type Application struct {
	Config config.Config
	Logger *log.Logger
	Models data.Models
	Mailer mailer.Mailer
}

func (app *Application) background(fn func()) {
	// Launch a background goroutine.
	go func() {
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.logError(fmt.Errorf("%s", err))
			}
		}()
		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}
