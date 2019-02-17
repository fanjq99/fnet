package fnet

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
)

type MessageMeta struct {
	Type reflect.Type
	Name string
	ID   uint32
}

var (
	metaByName = make(map[string]*MessageMeta)
	metaByID   = make(map[uint32]*MessageMeta)
	metaByType = make(map[reflect.Type]*MessageMeta)
)

// RegisterMessageMeta 注册消息元信息(代码生成专用)
func RegisterMessageMeta(id uint32, name string, msgType reflect.Type) {
	meta := &MessageMeta{
		Type: msgType,
		Name: name,
		ID:   id,
	}

	if _, ok := metaByName[name]; ok {
		panic("duplicate message meta register by name: " + name)
	}

	if _, ok := metaByID[id]; ok {
		panic(fmt.Sprintf("duplicate message meta register by id: %d", id))
	}

	if _, ok := metaByType[msgType]; ok {
		panic(fmt.Sprintf("duplicate message meta register by type : %d", id))
	}

	metaByName[name] = meta
	metaByID[id] = meta
	metaByType[msgType] = meta
}

// MessageMetaByName 根据名字查找消息元信息
func MessageMetaByName(name string) *MessageMeta {
	if v, ok := metaByName[name]; ok {
		return v
	}
	return nil
}

// MessageMetaByType 根据类型查找消息元消息
func MessageMetaByType(t reflect.Type) *MessageMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if v, ok := metaByType[t]; ok {
		return v
	}
	return nil
}

// MessageMeatByID 根据id查找消息元信息
func MessageMeatByID(id uint32) *MessageMeta {
	if v, ok := metaByID[id]; ok {
		return v
	}
	return nil
}

// MessageFullName 消息全名
func MessageFullName(rtype reflect.Type) string {
	if rtype == nil {
		panic("empty msg type")
	}

	if rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}

	var b bytes.Buffer
	b.WriteString(path.Base(rtype.PkgPath()))
	b.WriteString(".")
	b.WriteString(rtype.Name())

	return b.String()
}

// MessageNameByID 根据id查询消息名称
func MessageNameByID(id uint32) string {
	if meta := MessageMeatByID(id); meta != nil {
		return meta.Name
	}

	return ""
}

// ForEachMessageMeta 遍历消息元信息
func ForEachMessageMeta(callback func(*MessageMeta)) {
	for _, meta := range metaByName {
		callback(meta)
	}
}
