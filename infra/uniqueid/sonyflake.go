package uniqueid

import (
	"errors"
	"net"
	"os"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"github.com/sony/sonyflake"
)

func init() {
	hive.Prepare(func(initiator hive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *SonyflakerImpl {
			return &SonyflakerImpl{}
		})
	})
}

var (
	sf *sonyflake.Sonyflake
)

type Sonyflaker interface {
	SetPodIP(ip string)
	NextID() (uint64, error)
}

type SonyflakerImpl struct {
	hive.Infra
	podIP    string
	settings *sonyflake.Settings
}

func (sfi *SonyflakerImpl) BeginRequest(worker hive.Worker) {
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = "127.0.0.1"
	}
	sfi.podIP = podIP
	sfi.settings = &sonyflake.Settings{}
	sfi.Infra.BeginRequest(worker)
}

func (sfi *SonyflakerImpl) SetPodIP(ip string) {
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.', ':':
			sfi.podIP = ip
			break
		}
	}
}

// NextID 获取唯一ID
// https://github.com/tinrab/makaroni/tree/master/utilities/unique-id
func (sfi *SonyflakerImpl) NextID() (uint64, error) {
	sfi.settings.MachineID = func() (uint16, error) {
		ip := net.ParseIP(sfi.podIP)
		ip = ip.To16()
		if ip == nil || len(ip) < 4 {
			return 0, errors.New("invalid IP")
		}
		return uint16(ip[14])<<8 + uint16(ip[15]), nil
	}
	sf = sonyflake.NewSonyflake(*sfi.settings)
	if sf == nil {
		return 0, errors.New("sonyflake create error")
	}
	return sf.NextID()
}
