package uniqueid

/**
分布式系统全局唯一ID生成器组件

Created by Dustin.zhu on 2023/04/20.
*/

import (
	"errors"
	"math/rand"
	"net"
	"os"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"github.com/sony/sonyflake"

	"github.com/yitter/idgenerator-go/idgen"
)

//go:generate mockgen -package mock_infra -source sonyflake.go -destination ./mock/sonyflake_mock.go

func init() {
	options := idgen.NewIdGeneratorOptions(uint16(rand.Intn(2048))) // 0~2047(2^11 - 1)
	options.WorkerIdBitLength = 11                                  // WorkerIdBitLength + SeqBitLength <= 22
	options.SeqBitLength = 5                                        // 默认值6，限制每毫秒生成的ID个数。若生成速度超过5万个/秒，建议加大 SeqBitLength 到 10。
	options.BaseTime = 1690777000000                                // Mon Jul 31 2023 12:16:40 GMT+0800 (中国标准时间), 以此为基准时间,约30年后达到js最大数值
	idgen.SetIdGenerator(options)

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
	NextID() (int, error)
	ShortID() int
}

type SonyflakerImpl struct {
	hive.Infra
	podIP string
}

func (sfi *SonyflakerImpl) BeginRequest(worker hive.Worker) {
	sfi.Infra.BeginRequest(worker)
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = "127.0.0.1"
	}
	sfi.podIP = podIP
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

// NextID 获取唯一ID 18位
// https://github.com/tinrab/makaroni/tree/master/utilities/unique-id
func (sfi *SonyflakerImpl) NextID() (int, error) {
	settings := &sonyflake.Settings{}
	settings.MachineID = func() (uint16, error) {
		ip := net.ParseIP(sfi.podIP)
		ip = ip.To16()
		if ip == nil || len(ip) < 4 {
			return 0, errors.New("invalid IP")
		}
		return uint16(ip[14])<<8 + uint16(ip[15]), nil
	}
	sf = sonyflake.NewSonyflake(*settings)
	if sf == nil {
		return 0, errors.New("sonyflake create error")
	}
	nextID, err := sf.NextID()
	return int(nextID), err
}

// ShortID 16位
func (sfi *SonyflakerImpl) ShortID() int {
	return int(idgen.NextId())
}
