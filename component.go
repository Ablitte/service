package service

import (
	"github.com/greywords/peer"
	"time"
)

type Component interface {
	Init()
	OnSessionClose(*peer.Session) bool
	OnRequestFinished(*peer.Session, string, interface{}, string, time.Duration)
}

type ComponentBase struct{}

func (c *ComponentBase) Init() {}

func (c *ComponentBase) OnSessionClose(session *peer.Session) bool { return false }

func (c *ComponentBase) OnRequestFinished(session *peer.Session, router string, req interface{}, errmsg string, delta time.Duration) {
}
