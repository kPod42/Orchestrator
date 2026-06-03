package presence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"Orch/gen/go/presencepb"
	agentmodel "Orch/internal/agent/model"
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

func (c *Client) runSession(
	ctx context.Context,
	grpcAddr string,
	regResp *registerResponse,
) error {
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

	var sendMu sync.Mutex

	send := func(msg *presencepb.AgentPresenceMessage) error {
		sendMu.Lock()
		defer sendMu.Unlock()

		return stream.Send(msg)
	}

	if err := send(connectMessage(regResp.NodeID, regResp.SessionID)); err != nil {
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

	if err := c.flushDirty(send); err != nil {
		return fmt.Errorf("send initial presence snapshot: %w", err)
	}

	c.markReady()
	logger.Log("INFO", "PRESENCE", "presence session is ready: nodeID = %s", regResp.NodeID)

	recvErrCh := make(chan error, 1)

	go func() {
		recvErrCh <- c.receiveCoordinatorMessages(ctx, stream, send)
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Log("INFO", "PRESENCE", "presence client shutting down: nodeID = %s", c.cfg.Agent.NodeID)
			_ = stream.CloseSend()
			return nil

		case <-c.flushCh:
			if err := c.flushDirty(send); err != nil {
				return fmt.Errorf("send presence snapshot: %w", err)
			}

		case err := <-recvErrCh:
			return err
		}
	}
}

func (c *Client) receiveCoordinatorMessages(
	ctx context.Context,
	stream presencepb.PresenceService_ConnectClient,
	send sendPresenceMessageFunc,
) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("receive coordinator message: %w", err)
		}

		if err := c.handleCoordinatorMessage(ctx, msg, send); err != nil {
			return err
		}
	}
}

func (c *Client) handleCoordinatorMessage(
	ctx context.Context,
	msg *presencepb.CoordinatorPresenceMessage,
	send sendPresenceMessageFunc,
) error {
	switch payload := msg.Payload.(type) {
	case *presencepb.CoordinatorPresenceMessage_TaskRequest:
		go c.runTask(ctx, payload.TaskRequest, send)
		return nil

	case *presencepb.CoordinatorPresenceMessage_Error:
		logger.Log("ERROR", "PRESENCE", "coordinator error: %s", payload.Error.Message)
		return nil

	default:
		return nil
	}
}

func (c *Client) runTask(
	ctx context.Context,
	task *presencepb.TaskRequest,
	send sendPresenceMessageFunc,
) {
	taskID := task.TaskId

	if taskID == "" {
		logger.Log("ERROR", "PRESENCE", "received task without taskID")
		return
	}

	if !c.busy.Acquire() {
		_ = send(taskResultMessage(taskID, false, 1, "agent is busy"))
		return
	}
	defer c.busy.Release()

	if err := c.policy.CheckAction(task.Action); err != nil {
		_ = send(taskResultMessage(taskID, false, 1, err.Error()))
		return
	}

	c.SetBusy(true)
	defer c.SetBusy(false)

	events, err := c.executor.RunAction(ctx, task.Action, task.Args)
	if err != nil {
		_ = send(taskResultMessage(taskID, false, 1, err.Error()))
		return
	}

	resultSent := false

	for ev := range events {
		if ev.Output != nil {
			if err := sendTaskOutput(send, taskID, ev.Output); err != nil {
				logger.Log("ERROR", "PRESENCE", "failed to send task output: %v", err)
				return
			}
		}

		if ev.Result != nil {
			resultSent = true

			if err := sendTaskResult(send, taskID, ev.Result); err != nil {
				logger.Log("ERROR", "PRESENCE", "failed to send task result: %v", err)
				return
			}
		}
	}

	if !resultSent {
		_ = send(taskResultMessage(taskID, false, 1, "task finished without result"))
	}
}

func sendTaskOutput(
	send sendPresenceMessageFunc,
	taskID string,
	output *agentmodel.Output,
) error {
	return send(taskOutputMessage(
		taskID,
		output.Stream,
		output.Chunk,
	))
}

func sendTaskResult(
	send sendPresenceMessageFunc,
	taskID string,
	result *agentmodel.Result,
) error {
	return send(taskResultMessage(
		taskID,
		result.Success,
		result.ExitCode,
		result.Message,
	))
}

func (c *Client) flushDirty(send sendPresenceMessageFunc) error {
	snap := c.takeDirtySnapshot()

	if snap.dirtyEndpoints {
		if err := send(endpointMessage(snap.endpoints)); err != nil {
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
		if err := send(statusMessage(snap.busy)); err != nil {
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
