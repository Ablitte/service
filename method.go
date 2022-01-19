package service

import (
	"github.com/greywords/peer"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfBytes   = reflect.TypeOf(([]byte)(nil))
	typeOfSession = reflect.TypeOf(peer.NewSession(nil))
)

func isExported(name string) bool {
	w, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(w)
}

// 方法检测
func isHandlerMethod(method reflect.Method) bool {
	mt := method.Type
	if method.PkgPath != "" {
		return false
	}
	//两个入参
	if mt.NumIn() != 3 {
		return false
	}
	//一个返回值
	if mt.NumOut() != 1 {
		return false
	}
	if t1 := mt.In(1); t1.Kind() != reflect.Ptr || t1 != typeOfSession {
		return false
	}
	if (mt.In(2).Kind() != reflect.Ptr && mt.In(2) != typeOfBytes) || mt.Out(0) != typeOfError {
		return false
	}
	return true
}
