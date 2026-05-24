package presence

import (
	"context"
	"fmt"
	"time"

	"Orch/gen/go/presencepb"
	"Orch/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (c *Client) Start(ctx context.Context) error {
	logger.Log("INFO", "PRESENCE", "starting presence client: nodeID = %s", c.cfg.Agent.NodeID)

	regResp, err := c.register(ctx)
	if err != nil {
		return err
	}

	endpoints := coordinatorGRPCEndpoints(c.cfg, regResp)
	if len(endpoints) == 0 {
		return fmt.Errorf("presence grpc endpoints are empty")
	}

	var lastErr error

	for _, ep := range endpoints {
		if ctx.Err() != nil {
			return nil
		}

		logger.Log(
			"INFO",
			"PRESENCE",
			"trying coordinator endpoint: %s",
			endpointLabel(ep),
		)

		err := c.runSession(ctx, ep.Address, regResp)
		if err == nil {
			return nil
		}

		lastErr = err

		logger.Log(
			"WARNING",
			"PRESENCE",
			"coordinator endpoint failed: %s error = %v",
			endpointLabel(ep),
			err,
		)
	}

	return fmt.Errorf("all coordinator grpc endpoints failed: %w", lastErr)
}

func (c *Client) runSession(ctx context.Context, grpcAddr string, regResp *registerResponse) error {
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("dial coordinator presence grpc %s: %w", grpcAddr, err)
	}
	defer conn.Close()

	client := presencepb.NewPresenceServiceClient(conn)

	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("open presence stream: %w", err)
	}

	if err := stream.Send(connectMessage(regResp.NodeID, regResp.SessionID)); err != nil {
		return fmt.Errorf("send connect message: %w", err)
	}

	ack, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive connect ack: %w", err)
	}

	switch payload := ack.Payload.(type) {
	case *presencepb.CoordinatorPresenceMessage_ConnectAck:
		logger.Log("INFO", "PRESENCE", "connected to coordinator: nodeID = %s", payload.ConnectAck.NodeId)
	case *presencepb.CoordinatorPresenceMessage_Error:
		return fmt.Errorf("coordinator returned error: %s", payload.Error.Message)
	default:
		return fmt.Errorf("unexpected ack payload")
	}

	if err := c.flushDirty(stream); err != nil {
		return fmt.Errorf("send initial presence snapshot: %w", err)
	}

	c.markReady()
	logger.Log("INFO", "PRESENCE", "presence session is ready: nodeID = %s", regResp.NodeID)

	for {
		select {
		case <-ctx.Done():
			logger.Log("INFO", "PRESENCE", "presence client shutting down: nodeID = %s", c.cfg.Agent.NodeID)
			_ = stream.CloseSend()
			return nil

		case <-c.flushCh:
			if err := c.flushDirty(stream); err != nil {
				return fmt.Errorf("send presence snapshot: %w", err)
			}
		}
	}
}

func (c *Client) flushDirty(stream presencepb.PresenceService_ConnectClient) error {
	snap := c.takeDirtySnapshot()

	if snap.dirtyEndpoints {
		if err := stream.Send(endpointMessage(snap.endpoints)); err != nil {
			return fmt.Errorf("send endpoints update: %w", err)
		}

		logger.Log(
			"INFO",
			"PRESENCE",
			"presence endpoints sent: nodeID = %s endpoints = %d",
			c.cfg.Agent.NodeID,
			len(snap.endpoints),
		)
	}

	if snap.dirtyBusy {
		if err := stream.Send(statusMessage(snap.busy)); err != nil {
			return fmt.Errorf("send busy update: %w", err)
		}

		logger.Log(
			"INFO",
			"PRESENCE",
			"presence busy sent: nodeID = %s busy = %v",
			c.cfg.Agent.NodeID,
			snap.busy,
		)
	}

	return nil
}
