package service

import (
	"Coordinator/internal/model"
	"Coordinator/internal/registry"
	pb "Coordinator/internal/transport/grpc/pb"
	"errors"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PresenceService struct {
	pb.UnimplementedPresenceServiceServer
	reg registry.Registry
}

func NewPresenceService(reg registry.Registry) *PresenceService {
	return &PresenceService{reg: reg}
}

func (s *PresenceService) Connect(stream pb.PresenceService_ConnectServer) error {
	firstMsg, err := stream.Recv()
	if err != nil {
		return status.Error(codes.InvalidArgument, "failed to receive first message")
	}

	connect := firstMsg.GetConnect()
	if connect == nil {
		return status.Error(codes.InvalidArgument, "first message must be connect")
	}

	if connect.NodeId == "" || connect.SessionId == "" {
		return status.Error(codes.InvalidArgument, "nodeId and sessionId are required")
	}

	if err := s.reg.Attach(connect.NodeId, connect.SessionId); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}
	defer s.reg.Detach(connect.NodeId, connect.SessionId)

	if err := stream.Send(&pb.CoordinatorPresenceMessage{
		Payload: &pb.CoordinatorPresenceMessage_ConnectAck{
			ConnectAck: &pb.ConnectAck{
				NodeId: connect.NodeId,
			},
		},
	}); err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return nil
		}

		switch payload := msg.Payload.(type) {
		case *pb.AgentPresenceMessage_Status:
			if err := s.reg.UpdateStatus(connect.NodeId, connect.SessionId, payload.Status.Busy); err != nil {
				return status.Error(codes.FailedPrecondition, err.Error())
			}

		case *pb.AgentPresenceMessage_EndpointUpdate:
			endpoints := make([]model.Endpoint, 0, len(payload.EndpointUpdate.Endpoints))
			for _, ep := range payload.EndpointUpdate.Endpoints {
				endpoints = append(endpoints, model.Endpoint{
					Type:    ep.Kind,
					Address: ep.Address,
				})
			}

			if err := s.reg.UpdateEndpoints(connect.NodeId, connect.SessionId, endpoints); err != nil {
				return status.Error(codes.FailedPrecondition, err.Error())
			}
		}
	}
}
