// Code generated by protoc-gen-go.
// source: google/bigtable/admin/v2/instance.proto
// DO NOT EDIT!

package admin

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Possible states of an instance.
type Instance_State int32

const (
	// The state of the instance could not be determined.
	Instance_STATE_NOT_KNOWN Instance_State = 0
	// The instance has been successfully created and can serve requests
	// to its tables.
	Instance_READY Instance_State = 1
	// The instance is currently being created, and may be destroyed
	// if the creation process encounters an error.
	Instance_CREATING Instance_State = 2
)

var Instance_State_name = map[int32]string{
	0: "STATE_NOT_KNOWN",
	1: "READY",
	2: "CREATING",
}
var Instance_State_value = map[string]int32{
	"STATE_NOT_KNOWN": 0,
	"READY":           1,
	"CREATING":        2,
}

func (x Instance_State) String() string {
	return proto.EnumName(Instance_State_name, int32(x))
}
func (Instance_State) EnumDescriptor() ([]byte, []int) { return fileDescriptor3, []int{0, 0} }

// Possible states of a cluster.
type Cluster_State int32

const (
	// The state of the cluster could not be determined.
	Cluster_STATE_NOT_KNOWN Cluster_State = 0
	// The cluster has been successfully created and is ready to serve requests.
	Cluster_READY Cluster_State = 1
	// The cluster is currently being created, and may be destroyed
	// if the creation process encounters an error.
	// A cluster may not be able to serve requests while being created.
	Cluster_CREATING Cluster_State = 2
	// The cluster is currently being resized, and may revert to its previous
	// node count if the process encounters an error.
	// A cluster is still capable of serving requests while being resized,
	// but may exhibit performance as if its number of allocated nodes is
	// between the starting and requested states.
	Cluster_RESIZING Cluster_State = 3
	// The cluster has no backing nodes. The data (tables) still
	// exist, but no operations can be performed on the cluster.
	Cluster_DISABLED Cluster_State = 4
)

var Cluster_State_name = map[int32]string{
	0: "STATE_NOT_KNOWN",
	1: "READY",
	2: "CREATING",
	3: "RESIZING",
	4: "DISABLED",
}
var Cluster_State_value = map[string]int32{
	"STATE_NOT_KNOWN": 0,
	"READY":           1,
	"CREATING":        2,
	"RESIZING":        3,
	"DISABLED":        4,
}

func (x Cluster_State) String() string {
	return proto.EnumName(Cluster_State_name, int32(x))
}
func (Cluster_State) EnumDescriptor() ([]byte, []int) { return fileDescriptor3, []int{1, 0} }

// A collection of Bigtable [Tables][google.bigtable.admin.v2.Table] and
// the resources that serve them.
// All tables in an instance are served from a single
// [Cluster][google.bigtable.admin.v2.Cluster].
type Instance struct {
	// @OutputOnly
	// The unique name of the instance. Values are of the form
	// projects/<project>/instances/[a-z][a-z0-9\\-]+[a-z0-9]
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// The descriptive name for this instance as it appears in UIs.
	// Can be changed at any time, but should be kept globally unique
	// to avoid confusion.
	DisplayName string `protobuf:"bytes,2,opt,name=display_name,json=displayName" json:"display_name,omitempty"`
	//
	// The current state of the instance.
	State Instance_State `protobuf:"varint,3,opt,name=state,enum=google.bigtable.admin.v2.Instance_State" json:"state,omitempty"`
}

func (m *Instance) Reset()                    { *m = Instance{} }
func (m *Instance) String() string            { return proto.CompactTextString(m) }
func (*Instance) ProtoMessage()               {}
func (*Instance) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *Instance) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Instance) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

func (m *Instance) GetState() Instance_State {
	if m != nil {
		return m.State
	}
	return Instance_STATE_NOT_KNOWN
}

// A resizable group of nodes in a particular cloud location, capable
// of serving all [Tables][google.bigtable.admin.v2.Table] in the parent
// [Instance][google.bigtable.admin.v2.Instance].
type Cluster struct {
	// @OutputOnly
	// The unique name of the cluster. Values are of the form
	// projects/<project>/instances/<instance>/clusters/[a-z][-a-z0-9]*
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// @CreationOnly
	// The location where this cluster's nodes and storage reside. For best
	// performance, clients should be located as close as possible to this cluster.
	// Currently only zones are supported, e.g. projects/*/locations/us-central1-b
	Location string `protobuf:"bytes,2,opt,name=location" json:"location,omitempty"`
	// @OutputOnly
	// The current state of the cluster.
	State Cluster_State `protobuf:"varint,3,opt,name=state,enum=google.bigtable.admin.v2.Cluster_State" json:"state,omitempty"`
	// The number of nodes allocated to this cluster. More nodes enable higher
	// throughput and more consistent performance.
	ServeNodes int32 `protobuf:"varint,4,opt,name=serve_nodes,json=serveNodes" json:"serve_nodes,omitempty"`
	// @CreationOnly
	// The type of storage used by this cluster to serve its
	// parent instance's tables, unless explicitly overridden.
	DefaultStorageType StorageType `protobuf:"varint,5,opt,name=default_storage_type,json=defaultStorageType,enum=google.bigtable.admin.v2.StorageType" json:"default_storage_type,omitempty"`
}

func (m *Cluster) Reset()                    { *m = Cluster{} }
func (m *Cluster) String() string            { return proto.CompactTextString(m) }
func (*Cluster) ProtoMessage()               {}
func (*Cluster) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{1} }

func (m *Cluster) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Cluster) GetLocation() string {
	if m != nil {
		return m.Location
	}
	return ""
}

func (m *Cluster) GetState() Cluster_State {
	if m != nil {
		return m.State
	}
	return Cluster_STATE_NOT_KNOWN
}

func (m *Cluster) GetServeNodes() int32 {
	if m != nil {
		return m.ServeNodes
	}
	return 0
}

func (m *Cluster) GetDefaultStorageType() StorageType {
	if m != nil {
		return m.DefaultStorageType
	}
	return StorageType_STORAGE_TYPE_UNSPECIFIED
}

func init() {
	proto.RegisterType((*Instance)(nil), "google.bigtable.admin.v2.Instance")
	proto.RegisterType((*Cluster)(nil), "google.bigtable.admin.v2.Cluster")
	proto.RegisterEnum("google.bigtable.admin.v2.Instance_State", Instance_State_name, Instance_State_value)
	proto.RegisterEnum("google.bigtable.admin.v2.Cluster_State", Cluster_State_name, Cluster_State_value)
}

func init() { proto.RegisterFile("google/bigtable/admin/v2/instance.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 412 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0x51, 0xab, 0xd3, 0x30,
	0x14, 0xc7, 0xed, 0xee, 0xaa, 0xbb, 0xb9, 0x57, 0x2d, 0xd1, 0x87, 0x32, 0x06, 0xce, 0xc2, 0x58,
	0x9f, 0x52, 0x98, 0xf8, 0x24, 0x13, 0xba, 0xad, 0x48, 0x51, 0xba, 0xd9, 0x16, 0x86, 0x7b, 0x29,
	0x59, 0x1b, 0x43, 0xa1, 0x4d, 0x4a, 0x93, 0x0d, 0xf6, 0xcd, 0xfc, 0x02, 0x7e, 0x2f, 0x69, 0xda,
	0x89, 0x43, 0xfb, 0xe2, 0x5b, 0xfe, 0xe7, 0xfc, 0x92, 0xff, 0x9f, 0x93, 0x03, 0xe6, 0x94, 0x73,
	0x5a, 0x10, 0xe7, 0x98, 0x53, 0x89, 0x8f, 0x05, 0x71, 0x70, 0x56, 0xe6, 0xcc, 0x39, 0x2f, 0x9c,
	0x9c, 0x09, 0x89, 0x59, 0x4a, 0x50, 0x55, 0x73, 0xc9, 0xa1, 0xd9, 0x82, 0xe8, 0x0a, 0x22, 0x05,
	0xa2, 0xf3, 0x62, 0x3c, 0xe9, 0x9e, 0xc0, 0x55, 0xee, 0x60, 0xc6, 0xb8, 0xc4, 0x32, 0xe7, 0x4c,
	0xb4, 0xf7, 0xc6, 0xb3, 0x5e, 0x83, 0x94, 0x97, 0x25, 0x67, 0x2d, 0x66, 0xfd, 0xd0, 0xc0, 0xc8,
	0xef, 0x1c, 0x21, 0x04, 0x43, 0x86, 0x4b, 0x62, 0x6a, 0x53, 0xcd, 0xbe, 0x0f, 0xd5, 0x19, 0xbe,
	0x05, 0x8f, 0x59, 0x2e, 0xaa, 0x02, 0x5f, 0x12, 0xd5, 0x1b, 0xa8, 0xde, 0x43, 0x57, 0x0b, 0x1a,
	0xe4, 0x23, 0xd0, 0x85, 0xc4, 0x92, 0x98, 0x77, 0x53, 0xcd, 0x7e, 0xb1, 0xb0, 0x51, 0x5f, 0x64,
	0x74, 0x75, 0x42, 0x51, 0xc3, 0x87, 0xed, 0x35, 0xeb, 0x3d, 0xd0, 0x95, 0x86, 0xaf, 0xc0, 0xcb,
	0x28, 0x76, 0x63, 0x2f, 0x09, 0xb6, 0x71, 0xf2, 0x39, 0xd8, 0xee, 0x03, 0xe3, 0x09, 0xbc, 0x07,
	0x7a, 0xe8, 0xb9, 0x9b, 0x6f, 0x86, 0x06, 0x1f, 0xc1, 0x68, 0x1d, 0x7a, 0x6e, 0xec, 0x07, 0x9f,
	0x8c, 0x81, 0xf5, 0x73, 0x00, 0x9e, 0xad, 0x8b, 0x93, 0x90, 0xa4, 0xfe, 0x67, 0xf2, 0x31, 0x18,
	0x15, 0x3c, 0x55, 0x43, 0xe9, 0x52, 0xff, 0xd6, 0x70, 0x79, 0x1b, 0x79, 0xde, 0x1f, 0xb9, 0x73,
	0xb8, 0x49, 0x0c, 0xdf, 0x80, 0x07, 0x41, 0xea, 0x33, 0x49, 0x18, 0xcf, 0x88, 0x30, 0x87, 0x53,
	0xcd, 0xd6, 0x43, 0xa0, 0x4a, 0x41, 0x53, 0x81, 0x7b, 0xf0, 0x3a, 0x23, 0xdf, 0xf1, 0xa9, 0x90,
	0x89, 0x90, 0xbc, 0xc6, 0x94, 0x24, 0xf2, 0x52, 0x11, 0x53, 0x57, 0x76, 0xb3, 0x7e, 0xbb, 0xa8,
	0xa5, 0xe3, 0x4b, 0x45, 0x42, 0xd8, 0x3d, 0xf1, 0x47, 0xcd, 0xfa, 0xfa, 0x5f, 0xb3, 0x6a, 0x54,
	0xe8, 0x45, 0xfe, 0xa1, 0x51, 0x77, 0x8d, 0xda, 0xf8, 0x91, 0xbb, 0xfa, 0xe2, 0x6d, 0x8c, 0xe1,
	0x8a, 0x81, 0x49, 0xca, 0xcb, 0xde, 0x48, 0xab, 0xe7, 0xd7, 0x5f, 0xdb, 0x35, 0x1b, 0xb3, 0xd3,
	0x0e, 0xcb, 0x0e, 0xa5, 0xbc, 0xc0, 0x8c, 0x22, 0x5e, 0x53, 0x87, 0x12, 0xa6, 0xf6, 0xc9, 0x69,
	0x5b, 0xb8, 0xca, 0xc5, 0xdf, 0x9b, 0xf7, 0x41, 0x1d, 0x8e, 0x4f, 0x15, 0xf9, 0xee, 0x57, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x33, 0x2a, 0xdc, 0x7a, 0x03, 0x03, 0x00, 0x00,
}
