package service

import (
	"fmt"
	"github.com/greywords/codec"
	_ "github.com/greywords/codec/json"
	_ "github.com/greywords/codec/protobuf"
	log "github.com/greywords/logger"
	"github.com/greywords/peer"
	"os"
)

var serviceList map[string]*Service // all registered service
var routerCodec codec.Codec

const ChildName = "Hbase"

func RegisterService(comp ...Component) {
	for _, v := range comp {
		s := NewService(v)
		if _, ok := serviceList[s.Name]; ok {
			log.Error("service: service already defined: %s", s.Name)
		}
		if err := s.ExtractHandler(); err != nil {
			log.Error("service: extract handler function failed: %v", err)
		}
		serviceList[s.Name] = s
		for name, handler := range s.Handlers {
			router := fmt.Sprintf("%s.%s", s.Name, name)
			//注册消息 用于解码
			codec.RegisterMessage(router, handler.Type)
			log.Debug("service: router %s param %s registed", router, handler.Type)
		}
	}
}

func SetCodec(name string) error {
	routerCodec = codec.GetCodec(name)
	if routerCodec == nil {
		return fmt.Errorf("service: codec %s not registered", name)
	}
	return nil
}

func Send(session *peer.Session, router string, data interface{}) error {
	rb, err := routerCodec.Marshal(router, data, nil)
	if err != nil {
		return fmt.Errorf("service: %v", err)
	}
	return session.Conn.Send(rb)
}

func SendBytes(session *peer.Session, data []byte) error {
	return session.Conn.Send(data)
}

func init() {
	serviceList = make(map[string]*Service)
	// handlerList = make(map[interface{}]*Handler)
	routerCodec = codec.GetCodec("json_codec")
	if routerCodec == nil {
		fmt.Println("service: codec json_codec not registered")
		os.Exit(1)
	}
}
