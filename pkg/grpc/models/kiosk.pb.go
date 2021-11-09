// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/grpc/proto/models/kiosk.proto

package models

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type EnumScreenAction int32

const (
	EnumScreenAction_START    EnumScreenAction = 0
	EnumScreenAction_UPDATE   EnumScreenAction = 1
	EnumScreenAction_STOP     EnumScreenAction = 2
	EnumScreenAction_POWEROFF EnumScreenAction = 3
	EnumScreenAction_POWERON  EnumScreenAction = 4
)

var EnumScreenAction_name = map[int32]string{
	0: "START",
	1: "UPDATE",
	2: "STOP",
	3: "POWEROFF",
	4: "POWERON",
}

var EnumScreenAction_value = map[string]int32{
	"START":    0,
	"UPDATE":   1,
	"STOP":     2,
	"POWEROFF": 3,
	"POWERON":  4,
}

func (x EnumScreenAction) String() string {
	return proto.EnumName(EnumScreenAction_name, int32(x))
}

func (EnumScreenAction) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_943e55fd7ffbcf07, []int{0}
}

type KioskState struct {
	Content              string           `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Title                string           `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	SizeW                int64            `protobuf:"varint,3,opt,name=size_w,json=sizeW,proto3" json:"size_w,omitempty"`
	SizeH                int64            `protobuf:"varint,4,opt,name=size_h,json=sizeH,proto3" json:"size_h,omitempty"`
	Action               EnumScreenAction `protobuf:"varint,5,opt,name=action,proto3,enum=models.EnumScreenAction" json:"action,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *KioskState) Reset()         { *m = KioskState{} }
func (m *KioskState) String() string { return proto.CompactTextString(m) }
func (*KioskState) ProtoMessage()    {}
func (*KioskState) Descriptor() ([]byte, []int) {
	return fileDescriptor_943e55fd7ffbcf07, []int{0}
}

func (m *KioskState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KioskState.Unmarshal(m, b)
}
func (m *KioskState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KioskState.Marshal(b, m, deterministic)
}
func (m *KioskState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KioskState.Merge(m, src)
}
func (m *KioskState) XXX_Size() int {
	return xxx_messageInfo_KioskState.Size(m)
}
func (m *KioskState) XXX_DiscardUnknown() {
	xxx_messageInfo_KioskState.DiscardUnknown(m)
}

var xxx_messageInfo_KioskState proto.InternalMessageInfo

func (m *KioskState) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func (m *KioskState) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *KioskState) GetSizeW() int64 {
	if m != nil {
		return m.SizeW
	}
	return 0
}

func (m *KioskState) GetSizeH() int64 {
	if m != nil {
		return m.SizeH
	}
	return 0
}

func (m *KioskState) GetAction() EnumScreenAction {
	if m != nil {
		return m.Action
	}
	return EnumScreenAction_START
}

func init() {
	proto.RegisterEnum("models.EnumScreenAction", EnumScreenAction_name, EnumScreenAction_value)
	proto.RegisterType((*KioskState)(nil), "models.KioskState")
}

func init() {
	proto.RegisterFile("pkg/grpc/proto/models/kiosk.proto", fileDescriptor_943e55fd7ffbcf07)
}

var fileDescriptor_943e55fd7ffbcf07 = []byte{
	// 246 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x86, 0xcd, 0xda, 0x66, 0xdb, 0x51, 0x34, 0x1e, 0x14, 0x72, 0x59, 0xbd, 0x2a, 0x5e, 0xb4,
	0xa2, 0x4f, 0x50, 0xb1, 0x43, 0x10, 0xd6, 0x92, 0x56, 0x06, 0xde, 0xc8, 0xac, 0x61, 0x96, 0x6e,
	0x49, 0x69, 0x23, 0x82, 0xcf, 0xe2, 0xc3, 0xca, 0x12, 0x9d, 0xb0, 0xcb, 0xef, 0xfb, 0x0f, 0x9c,
	0x9f, 0x1f, 0x2e, 0xba, 0x76, 0x95, 0xac, 0xfa, 0xae, 0x4e, 0xba, 0x5e, 0x1b, 0x9d, 0x6c, 0xf4,
	0x9b, 0x5c, 0x0f, 0x49, 0xdb, 0xe8, 0xa1, 0x8d, 0xad, 0x42, 0xea, 0xdc, 0xe5, 0x37, 0x01, 0x78,
	0xdc, 0xfa, 0xd2, 0x2c, 0x8d, 0x44, 0x0e, 0xe3, 0x5a, 0x2b, 0x23, 0x95, 0xe1, 0x24, 0x24, 0xd1,
	0x54, 0xfc, 0x21, 0x9e, 0x41, 0x60, 0x1a, 0xb3, 0x96, 0x7c, 0x64, 0xbd, 0x03, 0x3c, 0x07, 0x3a,
	0x34, 0x5f, 0xf2, 0xe5, 0x93, 0x7b, 0x21, 0x89, 0x3c, 0x11, 0x6c, 0x69, 0xb1, 0xd3, 0xef, 0xdc,
	0xff, 0xd7, 0x0f, 0x78, 0x0d, 0x74, 0x59, 0x9b, 0x46, 0x2b, 0x1e, 0x84, 0x24, 0x3a, 0xbe, 0xe1,
	0xb1, 0x6b, 0x11, 0x67, 0xea, 0x63, 0x53, 0xd6, 0xbd, 0x94, 0x2a, 0xb5, 0xb9, 0xf8, 0xbd, 0xbb,
	0x9a, 0x03, 0xdb, 0xcf, 0x70, 0x0a, 0x41, 0x59, 0xa5, 0xa2, 0x62, 0x07, 0x08, 0x40, 0x9f, 0x8a,
	0xfb, 0xb4, 0xca, 0x18, 0xc1, 0x09, 0xf8, 0x65, 0x95, 0x17, 0x6c, 0x84, 0x47, 0x30, 0x29, 0xf2,
	0x45, 0x26, 0xf2, 0xd9, 0x8c, 0x79, 0x78, 0x08, 0x63, 0x47, 0x73, 0xe6, 0xdf, 0x9d, 0x3e, 0x9f,
	0xec, 0xb6, 0x71, 0xbf, 0x5f, 0xa9, 0x1d, 0xe4, 0xf6, 0x27, 0x00, 0x00, 0xff, 0xff, 0xde, 0xf5,
	0x88, 0x2d, 0x35, 0x01, 0x00, 0x00,
}
