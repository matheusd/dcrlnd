// Code generated by protoc-gen-go. DO NOT EDIT.
// source: invoicesrpc/invoices.proto

package invoicesrpc

import (
	context "context"
	fmt "fmt"
	lnrpc "github.com/decred/dcrlnd/lnrpc"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
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

type CancelInvoiceMsg struct {
	/// Hash corresponding to the (hold) invoice to cancel.
	PaymentHash          []byte   `protobuf:"bytes,1,opt,name=payment_hash,json=paymentHash,proto3" json:"payment_hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CancelInvoiceMsg) Reset()         { *m = CancelInvoiceMsg{} }
func (m *CancelInvoiceMsg) String() string { return proto.CompactTextString(m) }
func (*CancelInvoiceMsg) ProtoMessage()    {}
func (*CancelInvoiceMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{0}
}

func (m *CancelInvoiceMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelInvoiceMsg.Unmarshal(m, b)
}
func (m *CancelInvoiceMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelInvoiceMsg.Marshal(b, m, deterministic)
}
func (m *CancelInvoiceMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelInvoiceMsg.Merge(m, src)
}
func (m *CancelInvoiceMsg) XXX_Size() int {
	return xxx_messageInfo_CancelInvoiceMsg.Size(m)
}
func (m *CancelInvoiceMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelInvoiceMsg.DiscardUnknown(m)
}

var xxx_messageInfo_CancelInvoiceMsg proto.InternalMessageInfo

func (m *CancelInvoiceMsg) GetPaymentHash() []byte {
	if m != nil {
		return m.PaymentHash
	}
	return nil
}

type CancelInvoiceResp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CancelInvoiceResp) Reset()         { *m = CancelInvoiceResp{} }
func (m *CancelInvoiceResp) String() string { return proto.CompactTextString(m) }
func (*CancelInvoiceResp) ProtoMessage()    {}
func (*CancelInvoiceResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{1}
}

func (m *CancelInvoiceResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CancelInvoiceResp.Unmarshal(m, b)
}
func (m *CancelInvoiceResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CancelInvoiceResp.Marshal(b, m, deterministic)
}
func (m *CancelInvoiceResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CancelInvoiceResp.Merge(m, src)
}
func (m *CancelInvoiceResp) XXX_Size() int {
	return xxx_messageInfo_CancelInvoiceResp.Size(m)
}
func (m *CancelInvoiceResp) XXX_DiscardUnknown() {
	xxx_messageInfo_CancelInvoiceResp.DiscardUnknown(m)
}

var xxx_messageInfo_CancelInvoiceResp proto.InternalMessageInfo

type AddHoldInvoiceRequest struct {
	//*
	//An optional memo to attach along with the invoice. Used for record keeping
	//purposes for the invoice's creator, and will also be set in the description
	//field of the encoded payment request if the description_hash field is not
	//being used.
	Memo string `protobuf:"bytes,1,opt,name=memo,proto3" json:"memo,omitempty"`
	/// The hash of the preimage
	Hash []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	//*
	//The value of this invoice in atoms.
	//
	//The fields value and value_m_atoms are mutually exclusive.
	Value int64 `protobuf:"varint,3,opt,name=value,proto3" json:"value,omitempty"`
	//*
	//The value of this invoice in milliatoms.
	//
	//The fields value and value_m_atoms are mutually exclusive.
	ValueMAtoms int64 `protobuf:"varint,10,opt,name=value_m_atoms,proto3" json:"value_m_atoms,omitempty"`
	//*
	//Hash (SHA-256) of a description of the payment. Used if the description of
	//payment (memo) is too long to naturally fit within the description field
	//of an encoded payment request.
	DescriptionHash []byte `protobuf:"bytes,4,opt,name=description_hash,proto3" json:"description_hash,omitempty"`
	/// Payment request expiry time in seconds. Default is 3600 (1 hour).
	Expiry int64 `protobuf:"varint,5,opt,name=expiry,proto3" json:"expiry,omitempty"`
	/// Fallback on-chain address.
	FallbackAddr string `protobuf:"bytes,6,opt,name=fallback_addr,proto3" json:"fallback_addr,omitempty"`
	/// Delta to use for the time-lock of the CLTV extended to the final hop.
	CltvExpiry uint64 `protobuf:"varint,7,opt,name=cltv_expiry,proto3" json:"cltv_expiry,omitempty"`
	//*
	//Route hints that can each be individually used to assist in reaching the
	//invoice's destination.
	RouteHints []*lnrpc.RouteHint `protobuf:"bytes,8,rep,name=route_hints,proto3" json:"route_hints,omitempty"`
	/// Whether this invoice should include routing hints for private channels.
	Private              bool     `protobuf:"varint,9,opt,name=private,proto3" json:"private,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddHoldInvoiceRequest) Reset()         { *m = AddHoldInvoiceRequest{} }
func (m *AddHoldInvoiceRequest) String() string { return proto.CompactTextString(m) }
func (*AddHoldInvoiceRequest) ProtoMessage()    {}
func (*AddHoldInvoiceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{2}
}

func (m *AddHoldInvoiceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddHoldInvoiceRequest.Unmarshal(m, b)
}
func (m *AddHoldInvoiceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddHoldInvoiceRequest.Marshal(b, m, deterministic)
}
func (m *AddHoldInvoiceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddHoldInvoiceRequest.Merge(m, src)
}
func (m *AddHoldInvoiceRequest) XXX_Size() int {
	return xxx_messageInfo_AddHoldInvoiceRequest.Size(m)
}
func (m *AddHoldInvoiceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AddHoldInvoiceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AddHoldInvoiceRequest proto.InternalMessageInfo

func (m *AddHoldInvoiceRequest) GetMemo() string {
	if m != nil {
		return m.Memo
	}
	return ""
}

func (m *AddHoldInvoiceRequest) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *AddHoldInvoiceRequest) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *AddHoldInvoiceRequest) GetValueMAtoms() int64 {
	if m != nil {
		return m.ValueMAtoms
	}
	return 0
}

func (m *AddHoldInvoiceRequest) GetDescriptionHash() []byte {
	if m != nil {
		return m.DescriptionHash
	}
	return nil
}

func (m *AddHoldInvoiceRequest) GetExpiry() int64 {
	if m != nil {
		return m.Expiry
	}
	return 0
}

func (m *AddHoldInvoiceRequest) GetFallbackAddr() string {
	if m != nil {
		return m.FallbackAddr
	}
	return ""
}

func (m *AddHoldInvoiceRequest) GetCltvExpiry() uint64 {
	if m != nil {
		return m.CltvExpiry
	}
	return 0
}

func (m *AddHoldInvoiceRequest) GetRouteHints() []*lnrpc.RouteHint {
	if m != nil {
		return m.RouteHints
	}
	return nil
}

func (m *AddHoldInvoiceRequest) GetPrivate() bool {
	if m != nil {
		return m.Private
	}
	return false
}

type AddHoldInvoiceResp struct {
	//*
	//A bare-bones invoice for a payment within the Lightning Network.  With the
	//details of the invoice, the sender has all the data necessary to send a
	//payment to the recipient.
	PaymentRequest       string   `protobuf:"bytes,1,opt,name=payment_request,proto3" json:"payment_request,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddHoldInvoiceResp) Reset()         { *m = AddHoldInvoiceResp{} }
func (m *AddHoldInvoiceResp) String() string { return proto.CompactTextString(m) }
func (*AddHoldInvoiceResp) ProtoMessage()    {}
func (*AddHoldInvoiceResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{3}
}

func (m *AddHoldInvoiceResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddHoldInvoiceResp.Unmarshal(m, b)
}
func (m *AddHoldInvoiceResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddHoldInvoiceResp.Marshal(b, m, deterministic)
}
func (m *AddHoldInvoiceResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddHoldInvoiceResp.Merge(m, src)
}
func (m *AddHoldInvoiceResp) XXX_Size() int {
	return xxx_messageInfo_AddHoldInvoiceResp.Size(m)
}
func (m *AddHoldInvoiceResp) XXX_DiscardUnknown() {
	xxx_messageInfo_AddHoldInvoiceResp.DiscardUnknown(m)
}

var xxx_messageInfo_AddHoldInvoiceResp proto.InternalMessageInfo

func (m *AddHoldInvoiceResp) GetPaymentRequest() string {
	if m != nil {
		return m.PaymentRequest
	}
	return ""
}

type SettleInvoiceMsg struct {
	/// Externally discovered pre-image that should be used to settle the hold invoice.
	Preimage             []byte   `protobuf:"bytes,1,opt,name=preimage,proto3" json:"preimage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SettleInvoiceMsg) Reset()         { *m = SettleInvoiceMsg{} }
func (m *SettleInvoiceMsg) String() string { return proto.CompactTextString(m) }
func (*SettleInvoiceMsg) ProtoMessage()    {}
func (*SettleInvoiceMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{4}
}

func (m *SettleInvoiceMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SettleInvoiceMsg.Unmarshal(m, b)
}
func (m *SettleInvoiceMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SettleInvoiceMsg.Marshal(b, m, deterministic)
}
func (m *SettleInvoiceMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SettleInvoiceMsg.Merge(m, src)
}
func (m *SettleInvoiceMsg) XXX_Size() int {
	return xxx_messageInfo_SettleInvoiceMsg.Size(m)
}
func (m *SettleInvoiceMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SettleInvoiceMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SettleInvoiceMsg proto.InternalMessageInfo

func (m *SettleInvoiceMsg) GetPreimage() []byte {
	if m != nil {
		return m.Preimage
	}
	return nil
}

type SettleInvoiceResp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SettleInvoiceResp) Reset()         { *m = SettleInvoiceResp{} }
func (m *SettleInvoiceResp) String() string { return proto.CompactTextString(m) }
func (*SettleInvoiceResp) ProtoMessage()    {}
func (*SettleInvoiceResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{5}
}

func (m *SettleInvoiceResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SettleInvoiceResp.Unmarshal(m, b)
}
func (m *SettleInvoiceResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SettleInvoiceResp.Marshal(b, m, deterministic)
}
func (m *SettleInvoiceResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SettleInvoiceResp.Merge(m, src)
}
func (m *SettleInvoiceResp) XXX_Size() int {
	return xxx_messageInfo_SettleInvoiceResp.Size(m)
}
func (m *SettleInvoiceResp) XXX_DiscardUnknown() {
	xxx_messageInfo_SettleInvoiceResp.DiscardUnknown(m)
}

var xxx_messageInfo_SettleInvoiceResp proto.InternalMessageInfo

type SubscribeSingleInvoiceRequest struct {
	/// Hash corresponding to the (hold) invoice to subscribe to.
	RHash                []byte   `protobuf:"bytes,2,opt,name=r_hash,proto3" json:"r_hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubscribeSingleInvoiceRequest) Reset()         { *m = SubscribeSingleInvoiceRequest{} }
func (m *SubscribeSingleInvoiceRequest) String() string { return proto.CompactTextString(m) }
func (*SubscribeSingleInvoiceRequest) ProtoMessage()    {}
func (*SubscribeSingleInvoiceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_090ab9c4958b987d, []int{6}
}

func (m *SubscribeSingleInvoiceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubscribeSingleInvoiceRequest.Unmarshal(m, b)
}
func (m *SubscribeSingleInvoiceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubscribeSingleInvoiceRequest.Marshal(b, m, deterministic)
}
func (m *SubscribeSingleInvoiceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubscribeSingleInvoiceRequest.Merge(m, src)
}
func (m *SubscribeSingleInvoiceRequest) XXX_Size() int {
	return xxx_messageInfo_SubscribeSingleInvoiceRequest.Size(m)
}
func (m *SubscribeSingleInvoiceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SubscribeSingleInvoiceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SubscribeSingleInvoiceRequest proto.InternalMessageInfo

func (m *SubscribeSingleInvoiceRequest) GetRHash() []byte {
	if m != nil {
		return m.RHash
	}
	return nil
}

func init() {
	proto.RegisterType((*CancelInvoiceMsg)(nil), "invoicesrpc.CancelInvoiceMsg")
	proto.RegisterType((*CancelInvoiceResp)(nil), "invoicesrpc.CancelInvoiceResp")
	proto.RegisterType((*AddHoldInvoiceRequest)(nil), "invoicesrpc.AddHoldInvoiceRequest")
	proto.RegisterType((*AddHoldInvoiceResp)(nil), "invoicesrpc.AddHoldInvoiceResp")
	proto.RegisterType((*SettleInvoiceMsg)(nil), "invoicesrpc.SettleInvoiceMsg")
	proto.RegisterType((*SettleInvoiceResp)(nil), "invoicesrpc.SettleInvoiceResp")
	proto.RegisterType((*SubscribeSingleInvoiceRequest)(nil), "invoicesrpc.SubscribeSingleInvoiceRequest")
}

func init() { proto.RegisterFile("invoicesrpc/invoices.proto", fileDescriptor_090ab9c4958b987d) }

var fileDescriptor_090ab9c4958b987d = []byte{
	// 515 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x55, 0xda, 0x6e, 0xb7, 0x9d, 0xee, 0x2e, 0xc5, 0x40, 0x15, 0x45, 0x2c, 0x84, 0x88, 0x43,
	0x54, 0xa1, 0x04, 0x15, 0x71, 0x04, 0x09, 0xb8, 0x14, 0x24, 0x38, 0xa4, 0x82, 0x03, 0x97, 0xc8,
	0x4d, 0x4c, 0x6a, 0x91, 0xd8, 0xc6, 0x76, 0x2b, 0xf6, 0x43, 0x39, 0xf0, 0x37, 0x28, 0x8e, 0xbb,
	0x4a, 0xb2, 0xcb, 0xde, 0x66, 0xde, 0xcc, 0x3c, 0x4f, 0xde, 0x9b, 0x80, 0x47, 0xd9, 0x81, 0xd3,
	0x8c, 0x28, 0x29, 0xb2, 0xf8, 0x18, 0x47, 0x42, 0x72, 0xcd, 0xd1, 0xac, 0x55, 0xf3, 0x1e, 0x17,
	0x9c, 0x17, 0x25, 0x89, 0xb1, 0xa0, 0x31, 0x66, 0x8c, 0x6b, 0xac, 0x29, 0x67, 0xb6, 0xd5, 0x9b,
	0x4a, 0x91, 0x35, 0x61, 0xf0, 0x1a, 0xe6, 0x1f, 0x30, 0xcb, 0x48, 0xf9, 0xb1, 0x99, 0xfe, 0xac,
	0x0a, 0xf4, 0x0c, 0xce, 0x04, 0xbe, 0xaa, 0x08, 0xd3, 0xe9, 0x0e, 0xab, 0x9d, 0xeb, 0xf8, 0x4e,
	0x78, 0x96, 0xcc, 0x2c, 0xb6, 0xc6, 0x6a, 0x17, 0x3c, 0x80, 0xfb, 0x9d, 0xb1, 0x84, 0x28, 0x11,
	0xfc, 0x19, 0xc0, 0xa3, 0x77, 0x79, 0xbe, 0xe6, 0x65, 0x7e, 0x0d, 0xff, 0xda, 0x13, 0xa5, 0x11,
	0x82, 0x51, 0x45, 0x2a, 0x6e, 0x98, 0xa6, 0x89, 0x89, 0x6b, 0xcc, 0xb0, 0x0f, 0x0c, 0xbb, 0x89,
	0xd1, 0x43, 0x38, 0x39, 0xe0, 0x72, 0x4f, 0xdc, 0xa1, 0xef, 0x84, 0xc3, 0xa4, 0x49, 0xd0, 0x73,
	0x38, 0x37, 0x41, 0x5a, 0xa5, 0x58, 0xf3, 0x4a, 0xb9, 0x60, 0xaa, 0x5d, 0x10, 0x2d, 0x61, 0x9e,
	0x13, 0x95, 0x49, 0x2a, 0xea, 0x4f, 0x6d, 0x36, 0x1f, 0x19, 0xee, 0x1b, 0x38, 0x5a, 0xc0, 0x98,
	0xfc, 0x16, 0x54, 0x5e, 0xb9, 0x27, 0x86, 0xca, 0x66, 0xf5, 0x4b, 0x3f, 0x70, 0x59, 0x6e, 0x71,
	0xf6, 0x33, 0xc5, 0x79, 0x2e, 0xdd, 0xb1, 0x59, 0xb8, 0x0b, 0x22, 0x1f, 0x66, 0x59, 0xa9, 0x0f,
	0xa9, 0xa5, 0x38, 0xf5, 0x9d, 0x70, 0x94, 0xb4, 0x21, 0xb4, 0x82, 0x99, 0xe4, 0x7b, 0x4d, 0xd2,
	0x1d, 0x65, 0x5a, 0xb9, 0x13, 0x7f, 0x18, 0xce, 0x56, 0xf3, 0xa8, 0x64, 0xb5, 0xf0, 0x49, 0x5d,
	0x59, 0x53, 0xa6, 0x93, 0x76, 0x13, 0x72, 0xe1, 0x54, 0x48, 0x7a, 0xc0, 0x9a, 0xb8, 0x53, 0xdf,
	0x09, 0x27, 0xc9, 0x31, 0x0d, 0xde, 0x02, 0xea, 0xcb, 0xaa, 0x04, 0x0a, 0xe1, 0xde, 0xd1, 0x25,
	0xd9, 0xc8, 0x6c, 0xe5, 0xed, 0xc3, 0x41, 0x04, 0xf3, 0x0d, 0xd1, 0xba, 0x24, 0x2d, 0x8f, 0x3d,
	0x98, 0x08, 0x49, 0x68, 0x85, 0x0b, 0x62, 0xfd, 0xbd, 0xce, 0x6b, 0x73, 0x3b, 0xfd, 0xc6, 0xdc,
	0x37, 0x70, 0xb9, 0xd9, 0x6f, 0x6b, 0x1d, 0xb7, 0x64, 0x43, 0x59, 0xd1, 0xaa, 0x36, 0x1e, 0x2f,
	0x60, 0x2c, 0xd3, 0x96, 0xa3, 0x36, 0xfb, 0x34, 0x9a, 0x38, 0xf3, 0xc1, 0xea, 0xef, 0x00, 0x26,
	0x76, 0x40, 0xa1, 0x6f, 0xb0, 0xb8, 0x9d, 0x0b, 0x2d, 0xa3, 0xd6, 0x15, 0x47, 0x77, 0x3e, 0xe8,
	0x5d, 0x58, 0x3d, 0x2d, 0xfc, 0xd2, 0x41, 0x5f, 0xe0, 0xbc, 0x73, 0x95, 0xe8, 0xb2, 0x43, 0xd7,
	0x3f, 0x74, 0xef, 0xc9, 0xff, 0xcb, 0x46, 0xe2, 0xaf, 0x70, 0xd1, 0x15, 0x1e, 0x05, 0x9d, 0x89,
	0x5b, 0x8f, 0xdd, 0x7b, 0x7a, 0x67, 0x8f, 0x12, 0xf5, 0x9a, 0x1d, 0x7d, 0x7b, 0x6b, 0xf6, 0xbd,
	0xea, 0xad, 0x79, 0xc3, 0x9a, 0xf7, 0x2f, 0xbe, 0x2f, 0x0b, 0xaa, 0x77, 0xfb, 0x6d, 0x94, 0xf1,
	0x2a, 0xce, 0x49, 0x26, 0x49, 0x1e, 0xe7, 0x99, 0x2c, 0x59, 0x1e, 0x1b, 0x89, 0xe2, 0xd6, 0xfc,
	0x76, 0x6c, 0x7e, 0xfc, 0x57, 0xff, 0x02, 0x00, 0x00, 0xff, 0xff, 0x51, 0xf8, 0x61, 0x27, 0x4c,
	0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// InvoicesClient is the client API for Invoices service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type InvoicesClient interface {
	//*
	//SubscribeSingleInvoice returns a uni-directional stream (server -> client)
	//to notify the client of state transitions of the specified invoice.
	//Initially the current invoice state is always sent out.
	SubscribeSingleInvoice(ctx context.Context, in *SubscribeSingleInvoiceRequest, opts ...grpc.CallOption) (Invoices_SubscribeSingleInvoiceClient, error)
	//*
	//CancelInvoice cancels a currently open invoice. If the invoice is already
	//canceled, this call will succeed. If the invoice is already settled, it will
	//fail.
	CancelInvoice(ctx context.Context, in *CancelInvoiceMsg, opts ...grpc.CallOption) (*CancelInvoiceResp, error)
	//*
	//AddHoldInvoice creates a hold invoice. It ties the invoice to the hash
	//supplied in the request.
	AddHoldInvoice(ctx context.Context, in *AddHoldInvoiceRequest, opts ...grpc.CallOption) (*AddHoldInvoiceResp, error)
	//*
	//SettleInvoice settles an accepted invoice. If the invoice is already
	//settled, this call will succeed.
	SettleInvoice(ctx context.Context, in *SettleInvoiceMsg, opts ...grpc.CallOption) (*SettleInvoiceResp, error)
}

type invoicesClient struct {
	cc *grpc.ClientConn
}

func NewInvoicesClient(cc *grpc.ClientConn) InvoicesClient {
	return &invoicesClient{cc}
}

func (c *invoicesClient) SubscribeSingleInvoice(ctx context.Context, in *SubscribeSingleInvoiceRequest, opts ...grpc.CallOption) (Invoices_SubscribeSingleInvoiceClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Invoices_serviceDesc.Streams[0], "/invoicesrpc.Invoices/SubscribeSingleInvoice", opts...)
	if err != nil {
		return nil, err
	}
	x := &invoicesSubscribeSingleInvoiceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Invoices_SubscribeSingleInvoiceClient interface {
	Recv() (*lnrpc.Invoice, error)
	grpc.ClientStream
}

type invoicesSubscribeSingleInvoiceClient struct {
	grpc.ClientStream
}

func (x *invoicesSubscribeSingleInvoiceClient) Recv() (*lnrpc.Invoice, error) {
	m := new(lnrpc.Invoice)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *invoicesClient) CancelInvoice(ctx context.Context, in *CancelInvoiceMsg, opts ...grpc.CallOption) (*CancelInvoiceResp, error) {
	out := new(CancelInvoiceResp)
	err := c.cc.Invoke(ctx, "/invoicesrpc.Invoices/CancelInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *invoicesClient) AddHoldInvoice(ctx context.Context, in *AddHoldInvoiceRequest, opts ...grpc.CallOption) (*AddHoldInvoiceResp, error) {
	out := new(AddHoldInvoiceResp)
	err := c.cc.Invoke(ctx, "/invoicesrpc.Invoices/AddHoldInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *invoicesClient) SettleInvoice(ctx context.Context, in *SettleInvoiceMsg, opts ...grpc.CallOption) (*SettleInvoiceResp, error) {
	out := new(SettleInvoiceResp)
	err := c.cc.Invoke(ctx, "/invoicesrpc.Invoices/SettleInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InvoicesServer is the server API for Invoices service.
type InvoicesServer interface {
	//*
	//SubscribeSingleInvoice returns a uni-directional stream (server -> client)
	//to notify the client of state transitions of the specified invoice.
	//Initially the current invoice state is always sent out.
	SubscribeSingleInvoice(*SubscribeSingleInvoiceRequest, Invoices_SubscribeSingleInvoiceServer) error
	//*
	//CancelInvoice cancels a currently open invoice. If the invoice is already
	//canceled, this call will succeed. If the invoice is already settled, it will
	//fail.
	CancelInvoice(context.Context, *CancelInvoiceMsg) (*CancelInvoiceResp, error)
	//*
	//AddHoldInvoice creates a hold invoice. It ties the invoice to the hash
	//supplied in the request.
	AddHoldInvoice(context.Context, *AddHoldInvoiceRequest) (*AddHoldInvoiceResp, error)
	//*
	//SettleInvoice settles an accepted invoice. If the invoice is already
	//settled, this call will succeed.
	SettleInvoice(context.Context, *SettleInvoiceMsg) (*SettleInvoiceResp, error)
}

func RegisterInvoicesServer(s *grpc.Server, srv InvoicesServer) {
	s.RegisterService(&_Invoices_serviceDesc, srv)
}

func _Invoices_SubscribeSingleInvoice_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeSingleInvoiceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(InvoicesServer).SubscribeSingleInvoice(m, &invoicesSubscribeSingleInvoiceServer{stream})
}

type Invoices_SubscribeSingleInvoiceServer interface {
	Send(*lnrpc.Invoice) error
	grpc.ServerStream
}

type invoicesSubscribeSingleInvoiceServer struct {
	grpc.ServerStream
}

func (x *invoicesSubscribeSingleInvoiceServer) Send(m *lnrpc.Invoice) error {
	return x.ServerStream.SendMsg(m)
}

func _Invoices_CancelInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelInvoiceMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InvoicesServer).CancelInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicesrpc.Invoices/CancelInvoice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InvoicesServer).CancelInvoice(ctx, req.(*CancelInvoiceMsg))
	}
	return interceptor(ctx, in, info, handler)
}

func _Invoices_AddHoldInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddHoldInvoiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InvoicesServer).AddHoldInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicesrpc.Invoices/AddHoldInvoice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InvoicesServer).AddHoldInvoice(ctx, req.(*AddHoldInvoiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Invoices_SettleInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SettleInvoiceMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InvoicesServer).SettleInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicesrpc.Invoices/SettleInvoice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InvoicesServer).SettleInvoice(ctx, req.(*SettleInvoiceMsg))
	}
	return interceptor(ctx, in, info, handler)
}

var _Invoices_serviceDesc = grpc.ServiceDesc{
	ServiceName: "invoicesrpc.Invoices",
	HandlerType: (*InvoicesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CancelInvoice",
			Handler:    _Invoices_CancelInvoice_Handler,
		},
		{
			MethodName: "AddHoldInvoice",
			Handler:    _Invoices_AddHoldInvoice_Handler,
		},
		{
			MethodName: "SettleInvoice",
			Handler:    _Invoices_SettleInvoice_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeSingleInvoice",
			Handler:       _Invoices_SubscribeSingleInvoice_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "invoicesrpc/invoices.proto",
}
