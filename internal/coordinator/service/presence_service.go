package service

import (
	"errors"
	"io"

	"Orch/gen/go/presencepb"
	"Orch/internal/coordinator/model"
	"Orch/internal/coordinator/registry"
	"Orch/internal/coordinator/session"
	"Orch/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PresenceService struct {
	presencepb.UnimplementedPresenceServiceServer

	reg      registry.Registry
	sessions *session.Manager
}

func NewPresenceService(
	reg registry.Registry,
	sessions *session.Manager,
) *PresenceService {
	return &PresenceService{
		reg:      reg,
		sessions: sessions,
	}
}

func (s *PresenceService) Connect(stream presencepb.PresenceService_ConnectServer) error {
	logger.Log("INFO", "PRESENCE", "stream started")

	firstMsg, err := stream.Recv()
	if err != nil {
		logger.Log("ERROR", "PRESENCE", "failed to receive first message: %v", err)
		return status.Error(codes.InvalidArgument, "failed to receive first message")
	}

	connect := firstMsg.GetConnect()
	if connect == nil {
		logger.Log("ERROR", "PRESENCE", "first message must be connect")
		return status.Error(codes.InvalidArgument, "first message must be connect")
	}

	if connect.NodeId == "" || connect.SessionId == "" {
		logger.Log("ERROR", "PRESENCE", "nodeId and sessionId are required")
		return status.Error(codes.InvalidArgument, "nodeId and sessionId are required")
	}

	logger.Log(
		"INFO",
		"PRESENCE",
		"connect first message: nodeID = %s sessionID = %s",
		connect.NodeId,
		connect.SessionId,
	)

	if err := s.reg.Attach(connect.NodeId, connect.SessionId); err != nil {
		logger.Log(
			"ERROR",
			"PRESENCE",
			"failed to attach node: nodeID = %s sessionID = %s error = %v",
			connect.NodeId,
			connect.SessionId,
			err,
		)

		return status.Error(codes.PermissionDenied, err.Error())
	}

	controlSession := s.sessions.Attach(connect.NodeId, connect.SessionId)

	defer func() {
		s.sessions.Detach(connect.NodeId, connect.SessionId)
		s.reg.Detach(connect.NodeId, connect.SessionId)

		logger.Log(
			"INFO",
			"PRESENCE",
			"stream detached: nodeID = %s sessionID = %s",
			connect.NodeId,
			connect.SessionId,
		)
	}()

	if err := stream.Send(&presencepb.CoordinatorPresenceMessage{
		Payload: &presencepb.CoordinatorPresenceMessage_ConnectAck{
			ConnectAck: &presencepb.ConnectAck{
				NodeId: connect.NodeId,
			},
		},
	}); err != nil {
		logger.Log(
			"ERROR",
			"PRESENCE",
			"failed to send connect ack: nodeID = %s sessionID = %s error = %v",
			connect.NodeId,
			connect.SessionId,
			err,
		)

		return err
	}

	errCh := make(chan error, 2)

	go func() {
		errCh <- s.sendLoop(stream, controlSession)
	}()

	go func() {
		errCh <- s.receiveLoop(stream, connect.NodeId, connect.SessionId)
	}()

	err = <-errCh
	if err != nil {
		logger.Log(
			"ERROR",
			"PRESENCE",
			"control stream stopped with error: nodeID = %s sessionID = %s error = %v",
			connect.NodeId,
			connect.SessionId,
			err,
		)

		return err
	}

	return nil
}

func (s *PresenceService) sendLoop(
	stream presencepb.PresenceService_ConnectServer,
	controlSession *session.Session,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil

		case <-controlSession.Done():
			return nil

		case msg := <-controlSession.SendChannel():
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (s *PresenceService) receiveLoop(
	stream presencepb.PresenceService_ConnectServer,
	nodeID string,
	sessionID string,
) error {
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Log(
				"INFO",
				"PRESENCE",
				"received EOF: nodeID = %s sessionID = %s",
				nodeID,
				sessionID,
			)

			return nil
		}

		if err != nil {
			return err
		}

		if err := s.handleAgentMessage(nodeID, sessionID, msg); err != nil {
			return err
		}
	}
}

func (s *PresenceService) handleAgentMessage(
	nodeID string,
	sessionID string,
	msg *presencepb.AgentPresenceMessage,
) error {
	switch payload := msg.Payload.(type) {
	case *presencepb.AgentPresenceMessage_Status:
		if err := s.reg.UpdateStatus(nodeID, sessionID, payload.Status.Busy); err != nil {
			logger.Log(
				"ERROR",
				"PRESENCE",
				"failed to update status: nodeID = %s sessionID = %s",
				nodeID,
				sessionID,
			)

			return status.Error(codes.FailedPrecondition, err.Error())
		}

	case *presencepb.AgentPresenceMessage_EndpointUpdate:
		endpoints := make([]model.Endpoint, 0, len(payload.EndpointUpdate.Endpoints))

		for _, ep := range payload.EndpointUpdate.Endpoints {
			endpoints = append(endpoints, model.Endpoint{
				Type:    ep.Kind,
				Address: ep.Address,
			})
		}

		if err := s.reg.UpdateEndpoints(nodeID, sessionID, endpoints); err != nil {
			logger.Log(
				"ERROR",
				"PRESENCE",
				"failed to update endpoints: nodeID = %s sessionID = %s",
				nodeID,
				sessionID,
			)

			return status.Error(codes.FailedPrecondition, err.Error())
		}

	case *presencepb.AgentPresenceMessage_TaskOutput:
		if err := s.sessions.HandleAgentMessage(nodeID, sessionID, msg); err != nil {
			return err
		}

	case *presencepb.AgentPresenceMessage_TaskResult:
		if err := s.sessions.HandleAgentMessage(nodeID, sessionID, msg); err != nil {
			return err
		}
	}

	return nil
}
