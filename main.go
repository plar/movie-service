package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/cihub/seelog"

	"github.com/go-ini/ini"
	"github.com/plar/movie-service/rest"
)

var (
	config      = flag.String("config", "movie-service.ini", "Configuration file")
	logFileName = flag.String("log", "log.xml", "use custom log.xml")
)

func init() {
	flag.Parse()

	logger, err := log.LoggerFromConfigAsFile(*logFileName)
	if err == nil {
		log.ReplaceLogger(logger)
		return
	}

	logger, err = log.LoggerFromConfigAsString(_log_default)
	if err != nil {
		panic(err)
	}
}

func main() {
	defer log.Flush()

	cfg, err := ini.LooseLoad(_conf_default, *config)
	if err != nil {
		log.Error(fmt.Errorf("Cannot load settings: %s", err))
		return
	}

	//server := rest.NewRestServerWithLogger(log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile))
	ctx := rest.MovieServerContext{
		MessageQueueURI:      cfg.Section("rabbitmq").Key("uri").String(),
		ServiceURI:           cfg.Section("movie-service").Key("uri").String(),
		RottenTomatoesAPIKey: cfg.Section("rottentomatoes").Key("rottentomatoes_api_key").String(),
	}

	server, err := rest.NewMovieServer(ctx)
	if err != nil {
		log.Error(fmt.Errorf("Cannot create service: %s", err))
		return
	}

	// handle Ctrl-C & system.SIGTERM
	// TBD: HUP reload configs
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Infof("scraper: received system '%v' signal", sig)
		server.Quit()
	}()

	server.Start()
}
