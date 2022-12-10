package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
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
	addr := &atomic.Value{}

	f := fuup.NewFuup(false, config.AllowProxy, config.ListenSocks, config.FakeSubnet, config.LocalSubnet)

	go daemon(f, addr)

	for addr.Load() == nil {
		time.Sleep(time.Second)
	}

	for {
		crypt, err := kcp.NewAESBlockCrypt([]byte(config.Key)[:16])
		fuup.CheckError(err)

		dst := addr.Load().(string)
		c, err := kcp.DialWithOptions(dst, crypt, 0, 0)
		if err != nil {
			log.Printf("Dial KCP: %s %+v\n", dst, err)
			time.Sleep(time.Second * 10)
			continue
		}

		log.Printf("Dail KCP: %s\n", dst)

		f.HandleKCP(c)
	}
}

func daemon(f *fuup.Fuup, addr *atomic.Value) {
	for {
		resp, err := http.Get(config.ServerIpGetter)
		if err == nil {
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil && len(data) > 6 && len(data) < 16 {
				fuup.Xor(data, []byte(config.ServerIpKey))
				if d := addr.Load(); d == nil || d.(string) != string(data) {
					addr.Store(string(data))
					f.Close()
				}
			}
		}
		time.Sleep(time.Minute * 15)
	}
}