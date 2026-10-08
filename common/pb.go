package common

import (
	pb "github.com/therealtunge/backgammon/pb/autogen"
	"google.golang.org/protobuf/proto"
)

func CreateWriteReq(data []byte, fd uint32) *pb.WriteRequest {
	builder := pb.WriteRequest_builder{
		Data: data,
		Fd:   &fd,
	}

	return builder.Build()
}

func CreateConnReq(target string) *pb.ConnectRequest {
	builder := pb.ConnectRequest_builder{
		Dest: &target,
	}

	return builder.Build()
}

func CreateReadReq(fd uint32) *pb.ReadRequest {
	builder := pb.ReadRequest_builder{
		Fd: &fd,
	}

	return builder.Build()
}

func CreateCloseReq(fd uint32) *pb.CloseRequest {
	builder := pb.CloseRequest_builder{
		Fd: &fd,
	}

	return builder.Build()
}

func CreateResponse(fd uint32, read []byte, error string) ([]byte, error) {
	var builder pb.Response_builder
	errorRef := &error

	if error == "" {
		errorRef = nil
	}
	if read != nil {
		data := string(read)

		builder = pb.Response_builder{
			Fd:    &fd,
			Data:  &data,
			Error: errorRef,
		}
	} else {
		builder = pb.Response_builder{
			Fd:    &fd,
			Error: errorRef,
		}
	}

	return proto.Marshal(builder.Build())
}
