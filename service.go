package service

import (
	"errors"
	"github.com/greywords/peer"
	"reflect"
	"strings"
)

type Handler struct {
	Receiver reflect.Value  // 值
	Method   reflect.Method // 方法
	Type     reflect.Type   // 类型
	IsRawArg bool           // 数据是否需要序列化
}

type Service struct {
	Name      string              // 服务名
	Type      reflect.Type        // 服务类型
	Receiver  reflect.Value       // 服务值
	Handlers  map[string]*Handler // 注册的方法列表
	component Component
}

func NewService(comp Component) *Service {
	s := &Service{
		Type:      reflect.TypeOf(comp),
		Receiver:  reflect.ValueOf(comp),
		component: comp,
	}
	s.Name = strings.ToLower(reflect.Indirect(s.Receiver).Type().Name())
	//调用初始化方法
	s.component.Init()
	return s
}

// 遍历取出满足条件的函数
func (s *Service) suitableHandlerMethods(typ reflect.Type) map[string]*Handler {
	methods := make(map[string]*Handler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mt := method.Type
		mn := method.Name
		if isHandlerMethod(method) {
			raw := false
			if mt.In(2) == typeOfBytes {
				raw = true
			}
			mn = strings.ToLower(mn)
			methods[mn] = &Handler{Method: method, Type: mt.In(2), IsRawArg: raw}
		}
	}
	return methods
}

// 取结构体中的函数
func (s *Service) ExtractHandler() error {
	typeName := reflect.Indirect(s.Receiver).Type().Name()
	if typeName == "" {
		return errors.New("no service name for type " + s.Type.String())
	}
	if !isExported(typeName) {
		return errors.New("type " + typeName + " is not exported")
	}
	s.Handlers = s.suitableHandlerMethods(s.Type)
	for i := range s.Handlers {
		s.Handlers[i].Receiver = s.Receiver
	}
	if reflect.Indirect(s.Receiver).NumField() > 0 {
		filedNum := reflect.Indirect(s.Receiver).NumField()
		for i := 0; i < filedNum; i++ {
			ty := reflect.Indirect(s.Receiver).Field(i).Type().Name()
			if ty == ChildName {
				h := s.suitableHandlerMethods(reflect.Indirect(s.Receiver).Field(i).Elem().Type())
				for ih, v := range h {
					s.Handlers[ih] = v
					s.Handlers[ih].Receiver = reflect.Indirect(s.Receiver).Field(i).Elem()
				}
			}
		}
	}
	if len(s.Handlers) == 0 {
		str := "service: "
		method := s.suitableHandlerMethods(reflect.PtrTo(s.Type))
		if len(method) != 0 {
			str = "type " + s.Name + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "type " + s.Name + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}

	return nil
}

func (s *Service) OnSessionClose(session *peer.Session) bool {
	return s.component.OnSessionClose(session)
}
