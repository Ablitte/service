package service

import (
	"fmt"
	_ "github.com/greywords/codec/json"
	log "github.com/greywords/logger"
	"github.com/greywords/peer"
	_ "github.com/greywords/peer/tcp"
	_ "github.com/greywords/peer/ws"
	"github.com/greywords/utils/shared"
	"reflect"
	"strings"
	"time"
)

type callBackEntity struct{}

func (cb *callBackEntity) OnClosed(session *peer.Session) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("OnClosed: session:%d err:%v", session.Conn.ID(), err))
			fmt.Println(utils.CallStack())
		}
	}()
	for _, v := range serviceList {
		if ok := v.OnSessionClose(session); ok {
			return
		}
	}
}

//调用注册的函数
func callHandlerFunc(foo reflect.Method, args []reflect.Value) (retValue interface{}, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("callHandlerFunc: %v", err))
			fmt.Println(utils.CallStack())
			retValue = nil
			retErr = fmt.Errorf("callHandlerFunc: call method pkg:%s method:%s err:%v", foo.PkgPath, foo.Name, err)
		}
	}()
	if ret := foo.Func.Call(args); len(ret) > 0 {
		var err error = nil
		if r1 := ret[1].Interface(); r1 != nil {
			err = r1.(error)
		}
		return ret[0].Interface(), err
	}
	return nil, fmt.Errorf("callHandlerFunc: call method pkg:%s method:%s", foo.PkgPath, foo.Name)
}

//接收到消息后处理
func (cb *callBackEntity) OnReceive(session *peer.Session, msg []byte) error {
	_, msgPack, err := routerCodec.Unmarshal(msg)
	if err != nil {
		return fmt.Errorf("onreceive: %v", err)
	}
	router, ok := msgPack.Router.(string)
	if !ok {
		return fmt.Errorf("onreceive: invalid router:%v", msgPack.Router)
	}
	routerArr := strings.Split(router, ".")
	if len(routerArr) != 2 {
		return fmt.Errorf("onreceive: invalid router:%s", msgPack.Router)
	}
	s, ok := serviceList[routerArr[0]]
	if !ok {
		return fmt.Errorf("onreceive: function not registed router:%s err:%v", msgPack.Router, err)
	}
	h, ok := s.Handlers[routerArr[1]]
	if !ok {
		return fmt.Errorf("onreceive: function not registed router:%s err:%v", msgPack.Router, err)
	}
	t1 := time.Now()
	var args = []reflect.Value{h.Receiver, reflect.ValueOf(session), reflect.ValueOf(msgPack.DataPtr)}
	ack, err := callHandlerFunc(h.Method, args)
	if ack != nil && !reflect.ValueOf(ack).IsNil() {
		rb, err := routerCodec.Marshal(router, ack, nil)
		if err != nil {
			return fmt.Errorf("service: %v", err)
		}
		err = session.Conn.Send(rb)
		if err != nil {
			log.Warning("warn! service send msg failed router:%s err:%v", router, err)
		}
	} else {
		rb, err := routerCodec.Marshal(router, nil, err)
		if err != nil {
			return fmt.Errorf("service: %v", err)
		}
		err = session.Conn.Send(rb)
		if err != nil {
			log.Warning("warn! service send msg failed router:%s err:%v", router, err)
		}
	}
	var errmsg string
	if err != nil {
		errmsg = err.Error()
	}
	dt := time.Since(t1)
	go s.component.OnRequestFinished(session, router, routerCodec.ToString(msgPack.DataPtr), errmsg, dt)
	return nil
}

func GetSessionManager() *peer.SessionManager {
	return peer.NewSessionMgr(&callBackEntity{})
}
