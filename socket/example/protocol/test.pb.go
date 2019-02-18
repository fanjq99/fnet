// Code generated by protoc-gen-go. DO NOT EDIT.
// source: test.proto

package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TestData struct {
	Name             *string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Id               *uint32 `protobuf:"varint,2,opt,name=id" json:"id,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *TestData) Reset()                    { *m = TestData{} }
func (m *TestData) String() string            { return proto.CompactTextString(m) }
func (*TestData) ProtoMessage()               {}
func (*TestData) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *TestData) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *TestData) GetId() uint32 {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return 0
}

type C2S_TestRequest struct {
	Data             *TestData `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte    `json:"-"`
}

func (m *C2S_TestRequest) Reset()                    { *m = C2S_TestRequest{} }
func (m *C2S_TestRequest) String() string            { return proto.CompactTextString(m) }
func (*C2S_TestRequest) ProtoMessage()               {}
func (*C2S_TestRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *C2S_TestRequest) GetData() *TestData {
	if m != nil {
		return m.Data
	}
	return nil
}

type S2C_TestResponse struct {
	Res              *int32 `protobuf:"varint,2,opt,name=res" json:"res,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *S2C_TestResponse) Reset()                    { *m = S2C_TestResponse{} }
func (m *S2C_TestResponse) String() string            { return proto.CompactTextString(m) }
func (*S2C_TestResponse) ProtoMessage()               {}
func (*S2C_TestResponse) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *S2C_TestResponse) GetRes() int32 {
	if m != nil && m.Res != nil {
		return *m.Res
	}
	return 0
}

func init() {
	proto.RegisterType((*TestData)(nil), "protocol.TestData")
	proto.RegisterType((*C2S_TestRequest)(nil), "protocol.C2S_TestRequest")
	proto.RegisterType((*S2C_TestResponse)(nil), "protocol.S2C_TestResponse")
}

func init() { proto.RegisterFile("test.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 142 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0x49, 0x2d, 0x2e,
	0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x00, 0x53, 0xc9, 0xf9, 0x39, 0x4a, 0x2a, 0x5c,
	0x1c, 0x21, 0xa9, 0xc5, 0x25, 0x2e, 0x89, 0x25, 0x89, 0x42, 0x3c, 0x5c, 0x2c, 0x79, 0x89, 0xb9,
	0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x42, 0x5c, 0x5c, 0x4c, 0x99, 0x29, 0x12, 0x4c, 0x0a,
	0x8c, 0x1a, 0xbc, 0x4a, 0xc6, 0x5c, 0xfc, 0xce, 0x46, 0xc1, 0xf1, 0x20, 0x95, 0x41, 0xa9, 0x85,
	0xa5, 0xa9, 0xc5, 0x25, 0x42, 0x0a, 0x5c, 0x2c, 0x29, 0x89, 0x25, 0x89, 0x60, 0xc5, 0xdc, 0x46,
	0x42, 0x7a, 0x30, 0x13, 0xf5, 0x60, 0xc6, 0x29, 0xc9, 0x73, 0x09, 0x04, 0x1b, 0x39, 0x43, 0x35,
	0x15, 0x17, 0xe4, 0xe7, 0x15, 0xa7, 0x0a, 0x71, 0x73, 0x31, 0x17, 0xa5, 0x16, 0x83, 0x4d, 0x65,
	0x05, 0x04, 0x00, 0x00, 0xff, 0xff, 0x17, 0x31, 0x94, 0x25, 0x92, 0x00, 0x00, 0x00,
}
