// Code generated for syncpb package. DO NOT EDIT.
// source: packages/core/grpc/syncpb/sync.proto

package syncpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var (
	_ = reflect.TypeOf
	_ = sync.Once{}
)

type SyncEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	EntityType  string                 `protobuf:"bytes,2,opt,name=entity_type,json=entityType,proto3" json:"entity_type,omitempty"`
	EntityId    int64                  `protobuf:"varint,3,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	Action      string                 `protobuf:"bytes,4,opt,name=action,proto3" json:"action,omitempty"`
	PayloadJson []byte                 `protobuf:"bytes,5,opt,name=payload_json,json=payloadJson,proto3" json:"payload_json,omitempty"`
	StoreId     int32                  `protobuf:"varint,6,opt,name=store_id,json=storeId,proto3" json:"store_id,omitempty"`
	TenantSlug  string                 `protobuf:"bytes,7,opt,name=tenant_slug,json=tenantSlug,proto3" json:"tenant_slug,omitempty"`
	EventTime   *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=event_time,json=eventTime,proto3" json:"event_time,omitempty"`
	Sha256      string                 `protobuf:"bytes,9,opt,name=sha256,proto3" json:"sha256,omitempty"`
	ChunkOffset uint64                 `protobuf:"varint,10,opt,name=chunk_offset,json=chunkOffset,proto3" json:"chunk_offset,omitempty"`
	IsLastChunk bool                   `protobuf:"varint,11,opt,name=is_last_chunk,json=isLastChunk,proto3" json:"is_last_chunk,omitempty"`
	CorrelationId string               `protobuf:"bytes,12,opt,name=correlation_id,json=correlationId,proto3" json:"correlation_id,omitempty"`
}

func (x *SyncEvent) Reset() {
	*x = SyncEvent{}
}

func (x *SyncEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncEvent) ProtoMessage() {}

func (x *SyncEvent) ProtoReflect() protoreflect.Message {
	mi := &file_sync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *SyncEvent) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SyncEvent) GetEntityType() string {
	if x != nil {
		return x.EntityType
	}
	return ""
}

func (x *SyncEvent) GetEntityId() int64 {
	if x != nil {
		return x.EntityId
	}
	return 0
}

func (x *SyncEvent) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *SyncEvent) GetPayloadJson() []byte {
	if x != nil {
		return x.PayloadJson
	}
	return nil
}

func (x *SyncEvent) GetStoreId() int32 {
	if x != nil {
		return x.StoreId
	}
	return 0
}

func (x *SyncEvent) GetTenantSlug() string {
	if x != nil {
		return x.TenantSlug
	}
	return ""
}

func (x *SyncEvent) GetEventTime() *timestamppb.Timestamp {
	if x != nil {
		return x.EventTime
	}
	return nil
}

func (x *SyncEvent) GetSha256() string {
	if x != nil {
		return x.Sha256
	}
	return ""
}

func (x *SyncEvent) GetChunkOffset() uint64 {
	if x != nil {
		return x.ChunkOffset
	}
	return 0
}

func (x *SyncEvent) GetIsLastChunk() bool {
	if x != nil {
		return x.IsLastChunk
	}
	return false
}

func (x *SyncEvent) GetCorrelationId() string {
	if x != nil {
		return x.CorrelationId
	}
	return ""
}

type SyncAck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	EntityType   string                 `protobuf:"bytes,2,opt,name=entity_type,json=entityType,proto3" json:"entity_type,omitempty"`
	EntityId     int64                  `protobuf:"varint,3,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	Success      bool                   `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
	ErrorMessage string                 `protobuf:"bytes,5,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	Sha256       string                 `protobuf:"bytes,6,opt,name=sha256,proto3" json:"sha256,omitempty"`
	ProcessedAt  *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=processed_at,json=processedAt,proto3" json:"processed_at,omitempty"`
}

func (x *SyncAck) Reset() {
	*x = SyncAck{}
}

func (x *SyncAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncAck) ProtoMessage() {}

func (x *SyncAck) ProtoReflect() protoreflect.Message {
	mi := &file_sync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *SyncAck) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SyncAck) GetEntityType() string {
	if x != nil {
		return x.EntityType
	}
	return ""
}

func (x *SyncAck) GetEntityId() int64 {
	if x != nil {
		return x.EntityId
	}
	return 0
}

func (x *SyncAck) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SyncAck) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *SyncAck) GetSha256() string {
	if x != nil {
		return x.Sha256
	}
	return ""
}

func (x *SyncAck) GetProcessedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.ProcessedAt
	}
	return nil
}

type PullRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TenantSlug  string                 `protobuf:"bytes,1,opt,name=tenant_slug,json=tenantSlug,proto3" json:"tenant_slug,omitempty"`
	StoreId     int32                  `protobuf:"varint,2,opt,name=store_id,json=storeId,proto3" json:"store_id,omitempty"`
	Since       *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=since,proto3" json:"since,omitempty"`
	EntityTypes []string               `protobuf:"bytes,4,rep,name=entity_types,json=entityTypes,proto3" json:"entity_types,omitempty"`
	Limit       int32                  `protobuf:"varint,5,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *PullRequest) Reset() {
	*x = PullRequest{}
}

func (x *PullRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullRequest) ProtoMessage() {}

func (x *PullRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sync_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *PullRequest) GetTenantSlug() string {
	if x != nil {
		return x.TenantSlug
	}
	return ""
}

func (x *PullRequest) GetStoreId() int32 {
	if x != nil {
		return x.StoreId
	}
	return 0
}

func (x *PullRequest) GetSince() *timestamppb.Timestamp {
	if x != nil {
		return x.Since
	}
	return nil
}

func (x *PullRequest) GetEntityTypes() []string {
	if x != nil {
		return x.EntityTypes
	}
	return nil
}

func (x *PullRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

var file_sync_proto_msgTypes = make([]protoimpl.MessageInfo, 3)

var file_sync_proto_rawDesc = []byte{
	0x0a, 0x24, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x67, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x79, 0x6e, 0x63, 0x70, 0x62, 0x2f, 0x73, 0x79, 0x6e, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x73, 0x79, 0x6e, 0x63, 0x1a, 0x1f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
}

func init() {
	file_sync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
		switch v := v.(*SyncEvent); i {
		case 0:
			return &v.state
		case 1:
			return &v.sizeCache
		case 2:
			return &v.unknownFields
		default:
			return nil
		}
	}
	file_sync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
		switch v := v.(*SyncAck); i {
		case 0:
			return &v.state
		case 1:
			return &v.sizeCache
		case 2:
			return &v.unknownFields
		default:
			return nil
		}
	}
	file_sync_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
		switch v := v.(*PullRequest); i {
		case 0:
			return &v.state
		case 1:
			return &v.sizeCache
		case 2:
			return &v.unknownFields
		default:
			return nil
		}
	}
}
