// Code generated by protoc-gen-go. DO NOT EDIT.
// source: signrpc/signer.proto

package signrpc // import "github.com/decred/dcrlnd/lnrpc/signrpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type KeyLocator struct {
	// / The family of key being identified.
	KeyFamily int32 `protobuf:"varint,1,opt,name=key_family,json=keyFamily,proto3" json:"key_family,omitempty"`
	// / The precise index of the key being identified.
	KeyIndex             int32    `protobuf:"varint,2,opt,name=key_index,json=keyIndex,proto3" json:"key_index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KeyLocator) Reset()         { *m = KeyLocator{} }
func (m *KeyLocator) String() string { return proto.CompactTextString(m) }
func (*KeyLocator) ProtoMessage()    {}
func (*KeyLocator) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{0}
}
func (m *KeyLocator) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyLocator.Unmarshal(m, b)
}
func (m *KeyLocator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyLocator.Marshal(b, m, deterministic)
}
func (dst *KeyLocator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyLocator.Merge(dst, src)
}
func (m *KeyLocator) XXX_Size() int {
	return xxx_messageInfo_KeyLocator.Size(m)
}
func (m *KeyLocator) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyLocator.DiscardUnknown(m)
}

var xxx_messageInfo_KeyLocator proto.InternalMessageInfo

func (m *KeyLocator) GetKeyFamily() int32 {
	if m != nil {
		return m.KeyFamily
	}
	return 0
}

func (m *KeyLocator) GetKeyIndex() int32 {
	if m != nil {
		return m.KeyIndex
	}
	return 0
}

type KeyDescriptor struct {
	// *
	// The raw bytes of the key being identified. Either this or the KeyLocator
	// must be specified.
	RawKeyBytes []byte `protobuf:"bytes,1,opt,name=raw_key_bytes,json=rawKeyBytes,proto3" json:"raw_key_bytes,omitempty"`
	// *
	// The key locator that identifies which key to use for signing. Either this
	// or the raw bytes of the target key must be specified.
	KeyLoc               *KeyLocator `protobuf:"bytes,2,opt,name=key_loc,json=keyLoc,proto3" json:"key_loc,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *KeyDescriptor) Reset()         { *m = KeyDescriptor{} }
func (m *KeyDescriptor) String() string { return proto.CompactTextString(m) }
func (*KeyDescriptor) ProtoMessage()    {}
func (*KeyDescriptor) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{1}
}
func (m *KeyDescriptor) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeyDescriptor.Unmarshal(m, b)
}
func (m *KeyDescriptor) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeyDescriptor.Marshal(b, m, deterministic)
}
func (dst *KeyDescriptor) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeyDescriptor.Merge(dst, src)
}
func (m *KeyDescriptor) XXX_Size() int {
	return xxx_messageInfo_KeyDescriptor.Size(m)
}
func (m *KeyDescriptor) XXX_DiscardUnknown() {
	xxx_messageInfo_KeyDescriptor.DiscardUnknown(m)
}

var xxx_messageInfo_KeyDescriptor proto.InternalMessageInfo

func (m *KeyDescriptor) GetRawKeyBytes() []byte {
	if m != nil {
		return m.RawKeyBytes
	}
	return nil
}

func (m *KeyDescriptor) GetKeyLoc() *KeyLocator {
	if m != nil {
		return m.KeyLoc
	}
	return nil
}

type TxOut struct {
	// / The value of the output being spent.
	Value int64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	// / The script of the output being spent.
	PkScript             []byte   `protobuf:"bytes,2,opt,name=pk_script,json=pkScript,proto3" json:"pk_script,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TxOut) Reset()         { *m = TxOut{} }
func (m *TxOut) String() string { return proto.CompactTextString(m) }
func (*TxOut) ProtoMessage()    {}
func (*TxOut) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{2}
}
func (m *TxOut) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TxOut.Unmarshal(m, b)
}
func (m *TxOut) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TxOut.Marshal(b, m, deterministic)
}
func (dst *TxOut) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TxOut.Merge(dst, src)
}
func (m *TxOut) XXX_Size() int {
	return xxx_messageInfo_TxOut.Size(m)
}
func (m *TxOut) XXX_DiscardUnknown() {
	xxx_messageInfo_TxOut.DiscardUnknown(m)
}

var xxx_messageInfo_TxOut proto.InternalMessageInfo

func (m *TxOut) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *TxOut) GetPkScript() []byte {
	if m != nil {
		return m.PkScript
	}
	return nil
}

type SignDescriptor struct {
	// *
	// A descriptor that precisely describes *which* key to use for signing. This
	// may provide the raw public key directly, or require the Signer to re-derive
	// the key according to the populated derivation path.
	KeyDesc *KeyDescriptor `protobuf:"bytes,1,opt,name=key_desc,json=keyDesc,proto3" json:"key_desc,omitempty"`
	// *
	// A scalar value that will be added to the private key corresponding to the
	// above public key to obtain the private key to be used to sign this input.
	// This value is typically derived via the following computation:
	//
	// derivedKey = privkey + sha256(perCommitmentPoint || pubKey) mod N
	SingleTweak []byte `protobuf:"bytes,2,opt,name=single_tweak,json=singleTweak,proto3" json:"single_tweak,omitempty"`
	// *
	// A private key that will be used in combination with its corresponding
	// private key to derive the private key that is to be used to sign the target
	// input. Within the Lightning protocol, this value is typically the
	// commitment secret from a previously revoked commitment transaction. This
	// value is in combination with two hash values, and the original private key
	// to derive the private key to be used when signing.
	//
	// k = (privKey*sha256(pubKey || tweakPub) +
	// tweakPriv*sha256(tweakPub || pubKey)) mod N
	DoubleTweak []byte `protobuf:"bytes,3,opt,name=double_tweak,json=doubleTweak,proto3" json:"double_tweak,omitempty"`
	// *
	// The full script required to properly redeem the output.  This field will
	// only be populated if a p2wsh or a p2sh output is being signed.
	WitnessScript []byte `protobuf:"bytes,4,opt,name=witness_script,json=witnessScript,proto3" json:"witness_script,omitempty"`
	// *
	// A description of the output being spent. The value and script MUST be provided.
	Output *TxOut `protobuf:"bytes,5,opt,name=output,proto3" json:"output,omitempty"`
	// *
	// The target sighash type that should be used when generating the final
	// sighash, and signature.
	Sighash uint32 `protobuf:"varint,7,opt,name=sighash,proto3" json:"sighash,omitempty"`
	// *
	// The target input within the transaction that should be signed.
	InputIndex           int32    `protobuf:"varint,8,opt,name=input_index,json=inputIndex,proto3" json:"input_index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignDescriptor) Reset()         { *m = SignDescriptor{} }
func (m *SignDescriptor) String() string { return proto.CompactTextString(m) }
func (*SignDescriptor) ProtoMessage()    {}
func (*SignDescriptor) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{3}
}
func (m *SignDescriptor) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignDescriptor.Unmarshal(m, b)
}
func (m *SignDescriptor) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignDescriptor.Marshal(b, m, deterministic)
}
func (dst *SignDescriptor) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignDescriptor.Merge(dst, src)
}
func (m *SignDescriptor) XXX_Size() int {
	return xxx_messageInfo_SignDescriptor.Size(m)
}
func (m *SignDescriptor) XXX_DiscardUnknown() {
	xxx_messageInfo_SignDescriptor.DiscardUnknown(m)
}

var xxx_messageInfo_SignDescriptor proto.InternalMessageInfo

func (m *SignDescriptor) GetKeyDesc() *KeyDescriptor {
	if m != nil {
		return m.KeyDesc
	}
	return nil
}

func (m *SignDescriptor) GetSingleTweak() []byte {
	if m != nil {
		return m.SingleTweak
	}
	return nil
}

func (m *SignDescriptor) GetDoubleTweak() []byte {
	if m != nil {
		return m.DoubleTweak
	}
	return nil
}

func (m *SignDescriptor) GetWitnessScript() []byte {
	if m != nil {
		return m.WitnessScript
	}
	return nil
}

func (m *SignDescriptor) GetOutput() *TxOut {
	if m != nil {
		return m.Output
	}
	return nil
}

func (m *SignDescriptor) GetSighash() uint32 {
	if m != nil {
		return m.Sighash
	}
	return 0
}

func (m *SignDescriptor) GetInputIndex() int32 {
	if m != nil {
		return m.InputIndex
	}
	return 0
}

type SignReq struct {
	// / The raw bytes of the transaction to be signed.
	RawTxBytes []byte `protobuf:"bytes,1,opt,name=raw_tx_bytes,json=rawTxBytes,proto3" json:"raw_tx_bytes,omitempty"`
	// / A set of sign descriptors, for each input to be signed.
	SignDescs            []*SignDescriptor `protobuf:"bytes,2,rep,name=sign_descs,json=signDescs,proto3" json:"sign_descs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *SignReq) Reset()         { *m = SignReq{} }
func (m *SignReq) String() string { return proto.CompactTextString(m) }
func (*SignReq) ProtoMessage()    {}
func (*SignReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{4}
}
func (m *SignReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignReq.Unmarshal(m, b)
}
func (m *SignReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignReq.Marshal(b, m, deterministic)
}
func (dst *SignReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignReq.Merge(dst, src)
}
func (m *SignReq) XXX_Size() int {
	return xxx_messageInfo_SignReq.Size(m)
}
func (m *SignReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SignReq.DiscardUnknown(m)
}

var xxx_messageInfo_SignReq proto.InternalMessageInfo

func (m *SignReq) GetRawTxBytes() []byte {
	if m != nil {
		return m.RawTxBytes
	}
	return nil
}

func (m *SignReq) GetSignDescs() []*SignDescriptor {
	if m != nil {
		return m.SignDescs
	}
	return nil
}

type SignResp struct {
	// *
	// A set of signatures realized in a fixed 64-byte format ordered in ascending
	// input order.
	RawSigs              [][]byte `protobuf:"bytes,1,rep,name=raw_sigs,json=rawSigs,proto3" json:"raw_sigs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignResp) Reset()         { *m = SignResp{} }
func (m *SignResp) String() string { return proto.CompactTextString(m) }
func (*SignResp) ProtoMessage()    {}
func (*SignResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{5}
}
func (m *SignResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignResp.Unmarshal(m, b)
}
func (m *SignResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignResp.Marshal(b, m, deterministic)
}
func (dst *SignResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignResp.Merge(dst, src)
}
func (m *SignResp) XXX_Size() int {
	return xxx_messageInfo_SignResp.Size(m)
}
func (m *SignResp) XXX_DiscardUnknown() {
	xxx_messageInfo_SignResp.DiscardUnknown(m)
}

var xxx_messageInfo_SignResp proto.InternalMessageInfo

func (m *SignResp) GetRawSigs() [][]byte {
	if m != nil {
		return m.RawSigs
	}
	return nil
}

type InputScript struct {
	// / The serializes witness stack for the specified input.
	Witness [][]byte `protobuf:"bytes,1,rep,name=witness,proto3" json:"witness,omitempty"`
	// **
	// The optional sig script for the specified witness that will only be set if
	// the input specified is a nested p2sh witness program.
	SigScript            []byte   `protobuf:"bytes,2,opt,name=sig_script,json=sigScript,proto3" json:"sig_script,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *InputScript) Reset()         { *m = InputScript{} }
func (m *InputScript) String() string { return proto.CompactTextString(m) }
func (*InputScript) ProtoMessage()    {}
func (*InputScript) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{6}
}
func (m *InputScript) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InputScript.Unmarshal(m, b)
}
func (m *InputScript) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InputScript.Marshal(b, m, deterministic)
}
func (dst *InputScript) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InputScript.Merge(dst, src)
}
func (m *InputScript) XXX_Size() int {
	return xxx_messageInfo_InputScript.Size(m)
}
func (m *InputScript) XXX_DiscardUnknown() {
	xxx_messageInfo_InputScript.DiscardUnknown(m)
}

var xxx_messageInfo_InputScript proto.InternalMessageInfo

func (m *InputScript) GetWitness() [][]byte {
	if m != nil {
		return m.Witness
	}
	return nil
}

func (m *InputScript) GetSigScript() []byte {
	if m != nil {
		return m.SigScript
	}
	return nil
}

type InputScriptResp struct {
	// / The set of fully valid input scripts requested.
	InputScripts         []*InputScript `protobuf:"bytes,1,rep,name=input_scripts,json=inputScripts,proto3" json:"input_scripts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *InputScriptResp) Reset()         { *m = InputScriptResp{} }
func (m *InputScriptResp) String() string { return proto.CompactTextString(m) }
func (*InputScriptResp) ProtoMessage()    {}
func (*InputScriptResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_signer_6c6df08ff5e47989, []int{7}
}
func (m *InputScriptResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_InputScriptResp.Unmarshal(m, b)
}
func (m *InputScriptResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_InputScriptResp.Marshal(b, m, deterministic)
}
func (dst *InputScriptResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InputScriptResp.Merge(dst, src)
}
func (m *InputScriptResp) XXX_Size() int {
	return xxx_messageInfo_InputScriptResp.Size(m)
}
func (m *InputScriptResp) XXX_DiscardUnknown() {
	xxx_messageInfo_InputScriptResp.DiscardUnknown(m)
}

var xxx_messageInfo_InputScriptResp proto.InternalMessageInfo

func (m *InputScriptResp) GetInputScripts() []*InputScript {
	if m != nil {
		return m.InputScripts
	}
	return nil
}

func init() {
	proto.RegisterType((*KeyLocator)(nil), "signrpc.KeyLocator")
	proto.RegisterType((*KeyDescriptor)(nil), "signrpc.KeyDescriptor")
	proto.RegisterType((*TxOut)(nil), "signrpc.TxOut")
	proto.RegisterType((*SignDescriptor)(nil), "signrpc.SignDescriptor")
	proto.RegisterType((*SignReq)(nil), "signrpc.SignReq")
	proto.RegisterType((*SignResp)(nil), "signrpc.SignResp")
	proto.RegisterType((*InputScript)(nil), "signrpc.InputScript")
	proto.RegisterType((*InputScriptResp)(nil), "signrpc.InputScriptResp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SignerClient is the client API for Signer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SignerClient interface {
	// *
	// SignOutputRaw is a method that can be used to generated a signature for a
	// set of inputs/outputs to a transaction. Each request specifies details
	// concerning how the outputs should be signed, which keys they should be
	// signed with, and also any optional tweaks. The return value is a fixed
	// 64-byte signature (the same format as we use on the wire in Lightning).
	//
	// If we are  unable to sign using the specified keys, then an error will be
	// returned.
	SignOutputRaw(ctx context.Context, in *SignReq, opts ...grpc.CallOption) (*SignResp, error)
	// *
	// ComputeInputScript generates a complete InputIndex for the passed
	// transaction with the signature as defined within the passed SignDescriptor.
	// This method should be capable of generating the proper input script for
	// both regular p2wkh output and p2wkh outputs nested within a regular p2sh
	// output.
	//
	// Note that when using this method to sign inputs belonging to the wallet,
	// the only items of the SignDescriptor that need to be populated are pkScript
	// in the TxOut field, the value in that same field, and finally the input
	// index.
	ComputeInputScript(ctx context.Context, in *SignReq, opts ...grpc.CallOption) (*InputScriptResp, error)
}

type signerClient struct {
	cc *grpc.ClientConn
}

func NewSignerClient(cc *grpc.ClientConn) SignerClient {
	return &signerClient{cc}
}

func (c *signerClient) SignOutputRaw(ctx context.Context, in *SignReq, opts ...grpc.CallOption) (*SignResp, error) {
	out := new(SignResp)
	err := c.cc.Invoke(ctx, "/signrpc.Signer/SignOutputRaw", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *signerClient) ComputeInputScript(ctx context.Context, in *SignReq, opts ...grpc.CallOption) (*InputScriptResp, error) {
	out := new(InputScriptResp)
	err := c.cc.Invoke(ctx, "/signrpc.Signer/ComputeInputScript", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SignerServer is the server API for Signer service.
type SignerServer interface {
	// *
	// SignOutputRaw is a method that can be used to generated a signature for a
	// set of inputs/outputs to a transaction. Each request specifies details
	// concerning how the outputs should be signed, which keys they should be
	// signed with, and also any optional tweaks. The return value is a fixed
	// 64-byte signature (the same format as we use on the wire in Lightning).
	//
	// If we are  unable to sign using the specified keys, then an error will be
	// returned.
	SignOutputRaw(context.Context, *SignReq) (*SignResp, error)
	// *
	// ComputeInputScript generates a complete InputIndex for the passed
	// transaction with the signature as defined within the passed SignDescriptor.
	// This method should be capable of generating the proper input script for
	// both regular p2wkh output and p2wkh outputs nested within a regular p2sh
	// output.
	//
	// Note that when using this method to sign inputs belonging to the wallet,
	// the only items of the SignDescriptor that need to be populated are pkScript
	// in the TxOut field, the value in that same field, and finally the input
	// index.
	ComputeInputScript(context.Context, *SignReq) (*InputScriptResp, error)
}

func RegisterSignerServer(s *grpc.Server, srv SignerServer) {
	s.RegisterService(&_Signer_serviceDesc, srv)
}

func _Signer_SignOutputRaw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SignerServer).SignOutputRaw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/signrpc.Signer/SignOutputRaw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SignerServer).SignOutputRaw(ctx, req.(*SignReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Signer_ComputeInputScript_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SignerServer).ComputeInputScript(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/signrpc.Signer/ComputeInputScript",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SignerServer).ComputeInputScript(ctx, req.(*SignReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Signer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "signrpc.Signer",
	HandlerType: (*SignerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SignOutputRaw",
			Handler:    _Signer_SignOutputRaw_Handler,
		},
		{
			MethodName: "ComputeInputScript",
			Handler:    _Signer_ComputeInputScript_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "signrpc/signer.proto",
}

func init() { proto.RegisterFile("signrpc/signer.proto", fileDescriptor_signer_6c6df08ff5e47989) }

var fileDescriptor_signer_6c6df08ff5e47989 = []byte{
	// 556 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x53, 0x4d, 0x8f, 0xd3, 0x30,
	0x10, 0x55, 0x5b, 0xda, 0x74, 0x27, 0x49, 0x01, 0x53, 0x41, 0x00, 0x21, 0x4a, 0xa4, 0x5d, 0xf5,
	0x80, 0x5a, 0x51, 0x10, 0x12, 0x9c, 0xd0, 0x82, 0x56, 0xac, 0xba, 0xd2, 0x4a, 0x6e, 0x4f, 0x5c,
	0xa2, 0x34, 0x31, 0xa9, 0x95, 0x34, 0xf1, 0xda, 0x0e, 0x69, 0x6e, 0xfc, 0x07, 0xfe, 0x30, 0xb2,
	0x9d, 0x7e, 0xc1, 0x9e, 0x9a, 0xf7, 0x3c, 0x33, 0xef, 0x79, 0x5e, 0x0d, 0x43, 0x41, 0x93, 0x9c,
	0xb3, 0x68, 0xaa, 0x7e, 0x09, 0x9f, 0x30, 0x5e, 0xc8, 0x02, 0x59, 0x0d, 0xeb, 0x7f, 0x07, 0x98,
	0x93, 0xfa, 0xa6, 0x88, 0x42, 0x59, 0x70, 0xf4, 0x0a, 0x20, 0x25, 0x75, 0xf0, 0x33, 0xdc, 0xd0,
	0xac, 0xf6, 0x5a, 0xa3, 0xd6, 0xb8, 0x8b, 0xcf, 0x52, 0x52, 0x5f, 0x69, 0x02, 0xbd, 0x04, 0x05,
	0x02, 0x9a, 0xc7, 0x64, 0xeb, 0xb5, 0xf5, 0x69, 0x3f, 0x25, 0xf5, 0xb5, 0xc2, 0x7e, 0x08, 0xee,
	0x9c, 0xd4, 0xdf, 0x88, 0x88, 0x38, 0x65, 0x6a, 0x98, 0x0f, 0x2e, 0x0f, 0xab, 0x40, 0x75, 0xac,
	0x6a, 0x49, 0x84, 0x9e, 0xe7, 0x60, 0x9b, 0x87, 0xd5, 0x9c, 0xd4, 0x97, 0x8a, 0x42, 0x6f, 0xc1,
	0x52, 0xe7, 0x59, 0x11, 0xe9, 0x79, 0xf6, 0xec, 0xc9, 0xa4, 0x71, 0x36, 0x39, 0xd8, 0xc2, 0xbd,
	0x54, 0x7f, 0xfb, 0x9f, 0xa1, 0xbb, 0xdc, 0xde, 0x96, 0x12, 0x0d, 0xa1, 0xfb, 0x2b, 0xcc, 0x4a,
	0xa2, 0x47, 0x76, 0xb0, 0x01, 0xca, 0x1e, 0x4b, 0x03, 0xa3, 0xaf, 0xc7, 0x39, 0xb8, 0xcf, 0xd2,
	0x85, 0xc6, 0xfe, 0x9f, 0x36, 0x0c, 0x16, 0x34, 0xc9, 0x8f, 0x0c, 0xbe, 0x03, 0xe5, 0x3e, 0x88,
	0x89, 0x88, 0xf4, 0x20, 0x7b, 0xf6, 0xf4, 0x58, 0xfd, 0x50, 0x89, 0x95, 0x49, 0x05, 0xd1, 0x1b,
	0x70, 0x04, 0xcd, 0x93, 0x8c, 0x04, 0xb2, 0x22, 0x61, 0xda, 0xa8, 0xd8, 0x86, 0x5b, 0x2a, 0x4a,
	0x95, 0xc4, 0x45, 0xb9, 0xda, 0x97, 0x74, 0x4c, 0x89, 0xe1, 0x4c, 0xc9, 0x39, 0x0c, 0x2a, 0x2a,
	0x73, 0x22, 0xc4, 0xce, 0xed, 0x03, 0x5d, 0xe4, 0x36, 0xac, 0xb1, 0x8c, 0x2e, 0xa0, 0x57, 0x94,
	0x92, 0x95, 0xd2, 0xeb, 0x6a, 0x77, 0x83, 0xbd, 0x3b, 0xbd, 0x05, 0xdc, 0x9c, 0x22, 0x0f, 0x54,
	0x9c, 0xeb, 0x50, 0xac, 0x3d, 0x6b, 0xd4, 0x1a, 0xbb, 0x78, 0x07, 0xd1, 0x6b, 0xb0, 0x69, 0xce,
	0x4a, 0xd9, 0x44, 0xd6, 0xd7, 0x91, 0x81, 0xa6, 0x4c, 0x68, 0x11, 0x58, 0x6a, 0x29, 0x98, 0xdc,
	0xa1, 0x11, 0x38, 0x2a, 0x2e, 0xb9, 0x3d, 0x49, 0x0b, 0x78, 0x58, 0x2d, 0xb7, 0x26, 0xac, 0x8f,
	0x00, 0xca, 0x80, 0x5e, 0x98, 0xf0, 0xda, 0xa3, 0xce, 0xd8, 0x9e, 0x3d, 0xdb, 0x7b, 0x3a, 0x5d,
	0x2e, 0x3e, 0x13, 0x0d, 0x16, 0xfe, 0x39, 0xf4, 0x8d, 0x88, 0x60, 0xe8, 0x39, 0xf4, 0x95, 0x8a,
	0xa0, 0x89, 0x52, 0xe8, 0x8c, 0x1d, 0x6c, 0xf1, 0xb0, 0x5a, 0xd0, 0x44, 0xf8, 0x57, 0x60, 0x5f,
	0x2b, 0x67, 0xcd, 0xed, 0x3d, 0xb0, 0x9a, 0x75, 0xec, 0x0a, 0x1b, 0xa8, 0xfe, 0xa5, 0x82, 0x26,
	0xa7, 0x41, 0x2b, 0xb9, 0x26, 0xe9, 0x1b, 0x78, 0x78, 0x34, 0x47, 0xab, 0x7e, 0x02, 0xd7, 0xec,
	0xc1, 0xf4, 0x98, 0x89, 0xf6, 0x6c, 0xb8, 0x37, 0x7f, 0xdc, 0xe0, 0xd0, 0x03, 0x10, 0xb3, 0xdf,
	0x2d, 0xe8, 0x2d, 0xf4, 0xd3, 0x41, 0x1f, 0xc0, 0x55, 0x5f, 0xb7, 0x7a, 0xeb, 0x38, 0xac, 0xd0,
	0xa3, 0x93, 0xcb, 0x63, 0x72, 0xf7, 0xe2, 0xf1, 0x3f, 0x8c, 0x60, 0xe8, 0x0b, 0xa0, 0xaf, 0xc5,
	0x86, 0x95, 0x92, 0x1c, 0xdf, 0xee, 0xff, 0x56, 0xef, 0x5e, 0x33, 0x44, 0xb0, 0xcb, 0xf1, 0x8f,
	0x8b, 0x84, 0xca, 0x75, 0xb9, 0x9a, 0x44, 0xc5, 0x66, 0x1a, 0x93, 0x88, 0x93, 0x78, 0x1a, 0x47,
	0x3c, 0xcb, 0xe3, 0x69, 0xb6, 0x7f, 0xdb, 0x9c, 0x45, 0xab, 0x9e, 0x7e, 0xdd, 0xef, 0xff, 0x06,
	0x00, 0x00, 0xff, 0xff, 0xe8, 0x09, 0x0a, 0x14, 0xf5, 0x03, 0x00, 0x00,
}
