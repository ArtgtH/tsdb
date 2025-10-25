package benchmark

//
//import (
//	"time"
//
//	"github.com/golang/protobuf/proto"
//)
//
//type PointProto struct {
//	Timestamp int64   `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
//	Value     float64 `protobuf:"fixed64,2,opt,name=value,proto3" json:"value,omitempty"`
//}
//
//type SeriesDataProto struct {
//	Metric string            `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
//	Tags   map[string]string `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
//	Points []*PointProto     `protobuf:"bytes,3,rep,name=points,proto3" json:"points,omitempty"`
//}
//
//func (m *PointProto) Reset()         { *m = PointProto{} }
//func (m *PointProto) String() string { return proto.CompactTextString(m) }
//func (*PointProto) ProtoMessage()    {}
//
//func (m *SeriesDataProto) Reset()         { *m = SeriesDataProto{} }
//func (m *SeriesDataProto) String() string { return proto.CompactTextString(m) }
//func (*SeriesDataProto) ProtoMessage()    {}
//
//func BenchmarkProtobuf(dataset *SeriesData, datasetName string) BenchmarkResult {
//	if dataset == nil || len(dataset.Points) == 0 {
//		return BenchmarkResult{}
//	}
//
//	totalPoints := len(dataset.Points)
//	protoData := convertToProtoData(dataset)
//
//	encodeStart := time.Now()
//	encodedData, err := proto.Marshal(protoData)
//	encodeTime := time.Since(encodeStart)
//
//	if err != nil {
//		panic("Protobuf encoding failed: " + err.Error())
//	}
//
//	decodeStart := time.Now()
//	decodedProto := &SeriesDataProto{}
//	err = proto.Unmarshal(encodedData, decodedProto)
//	decodeTime := time.Since(decodeStart)
//
//	if err != nil {
//		panic("Protobuf decoding failed: " + err.Error())
//	}
//
//	decodedData := convertFromProtoData(decodedProto)
//	if len(decodedData) == 0 || len(decodedData[0].Points) != totalPoints {
//		panic("Protobuf data corruption")
//	}
//
//	encodeSpeed := float64(totalPoints) / encodeTime.Seconds()
//	decodeSpeed := float64(totalPoints) / decodeTime.Seconds()
//
//	return BenchmarkResult{
//		Format:      "PROTOBUF",
//		Dataset:     datasetName,
//		DataSize:    totalPoints,
//		EncodedSize: len(encodedData),
//		EncodeTime:  encodeTime,
//		DecodeTime:  decodeTime,
//		EncodeSpeed: encodeSpeed,
//		DecodeSpeed: decodeSpeed,
//	}
//}
//
//func convertToProtoData(data *SeriesData) *SeriesDataProto {
//	if data == nil {
//		return &SeriesDataProto{}
//	}
//
//	protoPoints := make([]*PointProto, len(data.Points))
//
//	for i, p := range data.Points {
//		protoPoints[i] = &PointProto{
//			Timestamp: p.Timestamp,
//			Value:     p.Value,
//		}
//	}
//
//	return &SeriesDataProto{
//		Metric: data.Metric,
//		Tags:   data.Tags,
//		Points: protoPoints,
//	}
//}
//
//func convertFromProtoData(protoData *SeriesDataProto) []*SeriesData {
//	if protoData == nil {
//		return nil
//	}
//
//	points := make([]Point, len(protoData.Points))
//
//	for i, p := range protoData.Points {
//		points[i] = Point{
//			Timestamp: p.Timestamp,
//			Value:     p.Value,
//		}
//	}
//
//	return []*SeriesData{
//		{
//			Metric: protoData.Metric,
//			Tags:   protoData.Tags,
//			Points: points,
//		},
//	}
//}
