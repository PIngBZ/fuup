package fuup

import (
	"time"

	"github.com/PIngBZ/kcp-go/v5"
	"github.com/xtaci/smux"
)

const (
	SALT          = "HappyFUUP"
	DATA_BUF_SIZE = 1024 * 1024
)

func SmuxConfig() *smux.Config {
	smuxConfig := smux.DefaultConfig()
	smuxConfig.Version = 1
	smuxConfig.MaxFrameSize = 1024 * 32
	smuxConfig.MaxReceiveBuffer = 1024 * 1024 * 32
	smuxConfig.MaxStreamBuffer = 1024 * 1024 * 4
	smuxConfig.KeepAliveDisabled = false
	smuxConfig.KeepAliveInterval = time.Second * 3
	smuxConfig.KeepAliveTimeout = time.Second * 15

	err := smux.VerifyConfig(smuxConfig)
	CheckError(err)

	return smuxConfig
}

func KcpConfig() kcp.KCPOptions {
	kcpConfig := kcp.KCPOptions{
		InitialTXRTOBackoff:       2,
		InitialTXRTOBackoffThresh: 128,
		EarlyRetransmit:           true,
	}
	return kcpConfig
}
