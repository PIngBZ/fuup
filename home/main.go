package main

import (
	"errors"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/PIngBZ/fuup"
	"github.com/xtaci/kcp-go"
)

var (
	configFile string
	config     *Config
)

func init() {
	rand.Seed(time.Now().Unix())
	fuup.OpenLog()

	flag.StringVar(&configFile, "c", "", "configure file")
	flag.Parse()

	if configFile == "" {
		fuup.CheckError(errors.New("no config file"))
	}

	var err error
	config, err = parseConfig(configFile)
	fuup.CheckError(err)
}

func main() {
	f := fuup.NewFuup(true, config.AllowProxy, config.ListenSocks, config.FakeSubnet, config.LocalSubnet)

	crypt, err := kcp.NewAESBlockCrypt([]byte(config.Key)[:16])
	fuup.CheckError(err)

	listener, err := kcp.ListenWithOptions(config.ListenKcp, crypt, 0, 0)
	fuup.CheckError(err)

	for {
		c, err := listener.AcceptKCP()
		if err != nil {
			log.Printf("AcceptKCP: %+v\n", err)
			continue
		}

		log.Printf("AcceptKCP %s\n", c.RemoteAddr().String())

		go f.HandleKCP(c)
	}
}
