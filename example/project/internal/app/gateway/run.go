package gateway

import (
	"github.com/culionbear/configure"
	"github.com/culionbear/configure/dirver/dir"
	"log"
	"project/internal/pkg/dba"
	"project/internal/pkg/logger"
)

func readConfig() error {
	dirDriver, err := dir.New("./conf")
	if err != nil {
		return err
	}
	conf, err := configure.New(dirDriver)
	if err != nil {
		return err
	}
	return conf.Listen(
		dba.Unit(),
		logger.Unit(),
	)
}

func Run() {
	err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}
}
