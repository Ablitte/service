package service_test

import (
	"encoding/json"
	"fmt"
	_ "github.com/greywords/codec/json"
	"github.com/greywords/peer"
	_ "github.com/greywords/peer/tcp"
	_ "github.com/greywords/peer/ws"
	"github.com/greywords/service"
	"log"
	"testing"
	"time"
)

type TestMessage struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}
type TestService struct {
	service.ComponentBase
}

func (s *TestService) Init() {
	log.Println("call init function.")
}
func (s *TestService) OnSessionClose(session *peer.Session) bool {
	log.Printf("session closed : %d", session.Conn.ID())
	return true
}
func (s *TestService) TestHandler(session *peer.Session, msg []byte) error {
	req := &TestMessage{}
	json.Unmarshal(msg, req)
	log.Println(req.Id, req.Content) //w
	return session.Conn.Send([]byte(time.Now().String() + " response content:" + req.Content))
}
func (s *TestService) TestFunc(session *peer.Session, msg []byte) error {
	req := &TestMessage{}
	json.Unmarshal(msg, req)
	log.Println(req.Id, req.Content) //w
	return session.Conn.Send([]byte(fmt.Sprintf(time.Now().String()+" func response content: %d,%s", req.Id, req.Content)))
}

func TestNewService(t *testing.T) {
	s := new(TestService)
	service.RegisterService(s)
	service.SetCodec("json_codec")
	svr := peer.GetAcceptor("ws")
	err := svr.Start("ws://:8080/echo", service.GetSessionManager())
	if err != nil {
		log.Fatal(err)
	}
}

func TestTcpService(t *testing.T) {
	s := new(TestService)
	service.RegisterService(s)
	service.SetCodec("protobuf_codec")
	svr := peer.GetAcceptor("tcp")
	err := svr.Start("127.0.0.1:9999", service.GetSessionManager())
	if err != nil {
		log.Fatal(err.Error())
	}
}
