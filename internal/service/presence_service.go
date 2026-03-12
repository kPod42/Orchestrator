package service

import (
	"Coordinator/internal/logger"
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
	logger.Presence("stream opened")
	firstMsg, err := stream.Recv()
	if err != nil {
		logger.Error("presence failed to receive first message: %v", err)
		return status.Error(codes.InvalidArgument, "failed to receive first message")
	}

	connect := firstMsg.GetConnect()
	if connect == nil {
		logger.Error("presence first message is not connect")
		return status.Error(codes.InvalidArgument, "first message must be connect")
	}

	logger.Presence("connect request: nodeId=%s sessionId=%s", connect.NodeId, connect.SessionId)

	if connect.NodeId == "" || connect.SessionId == "" {
		logger.Error("presence invalid connect payload: nodeId=%s sessionId=%s",
			connect.NodeId, connect.SessionId)
		return status.Error(codes.InvalidArgument, "nodeId and sessionId are required")
	}

	if err := s.reg.Attach(connect.NodeId, connect.SessionId); err != nil {
		logger.Error("presence attach failed: nodeId=%s err=%v", connect.NodeId, err)
		return status.Error(codes.PermissionDenied, err.Error())
	}
	defer func() {
		s.reg.Detach(connect.NodeId, connect.SessionId)
		logger.Presence("stream detached: nodeId=%s", connect.NodeId)
	}()

	if err := stream.Send(&pb.CoordinatorPresenceMessage{
		Payload: &pb.CoordinatorPresenceMessage_ConnectAck{
			ConnectAck: &pb.ConnectAck{
				NodeId: connect.NodeId,
			},
		},
	}); err != nil {
		logger.Error("presence ack send failed: nodeId=%s err=%v", connect.NodeId, err)
		return err
	}

	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Presence("stream closed by client: nodeId=%s", connect.NodeId)
			return nil
		}
		if err != nil {
			logger.Error("presence stream recv error: nodeId=%s err=%v", connect.NodeId, err)
			return nil
		}

		switch payload := msg.Payload.(type) {
		case *pb.AgentPresenceMessage_Status:
			if err := s.reg.UpdateStatus(connect.NodeId, connect.SessionId, payload.Status.Busy); err != nil {
				logger.Error("presence status update failed: nodeId=%s err=%v", connect.NodeId, err)
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
				logger.Error("presence endpoint update failed: nodeId=%s err=%v", connect.NodeId, err)
				return status.Error(codes.FailedPrecondition, err.Error())
			}
		}
	}
}
