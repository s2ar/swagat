package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/s2ar/swagat/config"
	"github.com/s2ar/swagat/internal/handler"
	"github.com/s2ar/swagat/internal/service"
	"github.com/shopspring/decimal"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func InitConfig(filename string) (*config.Configuration, error) {
	cfg, err := config.InitConfig(filename)
	if err != nil {
		return nil, errors.Wrap(err, "cannot load config")
	}

	return cfg, nil
}

func run() error {
	decimal.MarshalJSONWithoutQuotes = true

	app := &cli.App{
		Name:  "swagat",
		Usage: "swagat service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "set config file",
				Aliases: []string{"c"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "start app in server mode",
				Action: cmdStartServer,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		return errors.Wrap(err, "cannot run app")
	}
	return nil
}

func cmdStartServer(cliContext *cli.Context) error {
	cfg, err := InitConfig(cliContext.String("config"))
	if err != nil {
		return err
	}

	services := service.New(cfg)
	handlers := handler.New(services)

	go services.Hub.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startServer(ctx, handlers, cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	fmt.Println("closing")
	return nil
}

func startServer(ctx context.Context, h *handler.Handler, cfg *config.Configuration) {
	srv := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: h.InitRouters(),
	}
	log.Printf("Starting HTTP server on %s", cfg.Server.Listen)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(errors.Wrap(err, "cannot start http server"))
		}
	}()

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Println(errors.Wrap(err, "cannot shutdown http server"))
	}
}
