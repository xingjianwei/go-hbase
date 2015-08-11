package hbase

import (
	"github.com/c4pt0r/go-hbase/proto"
	pb "github.com/golang/protobuf/proto"
)

type call struct {
	id             uint32
	methodName     string
	request        pb.Message
	responseBuffer pb.Message
	responseCh     chan pb.Message
}

type exception struct {
	msg string
}

func (m *exception) Reset()         { *m = exception{} }
func (m *exception) String() string { return m.msg }
func (*exception) ProtoMessage()    {}

func newCall(request pb.Message) *call {
	var responseBuffer pb.Message
	var methodName string

	switch request.(type) {
	case *proto.GetRequest:
		responseBuffer = &proto.GetResponse{}
		methodName = "Get"
	case *proto.MutateRequest:
		responseBuffer = &proto.MutateResponse{}
		methodName = "Mutate"
	case *proto.ScanRequest:
		responseBuffer = &proto.ScanResponse{}
		methodName = "Scan"
	case *proto.GetTableDescriptorsRequest:
		responseBuffer = &proto.GetTableDescriptorsResponse{}
		methodName = "GetTableDescriptors"
	case *proto.CoprocessorServiceRequest:
		responseBuffer = &proto.CoprocessorServiceResponse{}
		methodName = "ExecService"
	}

	return &call{
		methodName:     methodName,
		request:        request,
		responseBuffer: responseBuffer,
		responseCh:     make(chan pb.Message, 1),
	}
}

func (c *call) complete(err error, response []byte) {
	defer close(c.responseCh)

	if err != nil {
		c.responseCh <- &exception{
			msg: err.Error(),
		}
		return
	}

	err2 := pb.Unmarshal(response, c.responseBuffer)
	if err2 != nil {
		c.responseCh <- &exception{
			msg: err2.Error(),
		}
		return
	}

	c.responseCh <- c.responseBuffer
}
