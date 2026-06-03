package dispatcher

import (
	"context"
	"fmt"
	"strings"

	"Orch/internal/coordinator/model"
	"Orch/internal/coordinator/registry"
	"Orch/internal/coordinator/session"
)

type Dispatcher struct {
	registry registry.Registry
	sessions *session.Manager
}

func New(
	registry registry.Registry,
	sessions *session.Manager,
) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		sessions: sessions,
	}
}

func (d *Dispatcher) ExecuteAction(
	ctx context.Context,
	request model.ExecuteRequest,
) (model.ExecuteResponse, error) {
	if strings.TrimSpace(request.Action) == "" {
		return model.ExecuteResponse{}, fmt.Errorf("action is empty")
	}

	if len(request.Targets) == 0 {
		return model.ExecuteResponse{}, fmt.Errorf("targets are empty")
	}

	if request.Args == nil {
		request.Args = make(map[string]string)
	}

	activeNodes := d.registry.GetActive()
	selectedNodes := selectTargetNodes(activeNodes, request.Targets)

	results := make([]model.NodeExecutionResult, 0, len(selectedNodes))

	for _, selected := range selectedNodes {
		if selected.missing {
			results = append(results, model.NodeExecutionResult{
				NodeID:   selected.targetID,
				Status:   "error",
				ExitCode: 1,
				Message:  "node is not active",
			})
			continue
		}

		node := selected.node

		if node.Busy {
			results = append(results, model.NodeExecutionResult{
				NodeID:   node.ID,
				Status:   "error",
				ExitCode: 1,
				Message:  "node is busy",
			})
			continue
		}

		runResult, err := d.sessions.RunAction(
			ctx,
			node.ID,
			request.Action,
			request.Args,
		)
		if err != nil {
			results = append(results, model.NodeExecutionResult{
				NodeID:   node.ID,
				Status:   "error",
				ExitCode: 1,
				Message:  err.Error(),
			})
			continue
		}

		results = append(results, model.NodeExecutionResult{
			NodeID:   node.ID,
			Status:   runResult.Status,
			ExitCode: runResult.ExitCode,
			Message:  runResult.Message,
			Output:   runResult.Output,
		})
	}

	return model.ExecuteResponse{
		Results: results,
	}, nil
}

type selectedNode struct {
	targetID string
	node     model.Node
	missing  bool
}

func selectTargetNodes(nodes []model.Node, targets []string) []selectedNode {
	if isAllTarget(targets) {
		result := make([]selectedNode, 0, len(nodes))

		for _, node := range nodes {
			result = append(result, selectedNode{
				targetID: node.ID,
				node:     node,
				missing:  false,
			})
		}

		return result
	}

	result := make([]selectedNode, 0, len(targets))

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		node, ok := findNode(nodes, target)
		if !ok {
			result = append(result, selectedNode{
				targetID: target,
				missing:  true,
			})
			continue
		}

		result = append(result, selectedNode{
			targetID: target,
			node:     node,
			missing:  false,
		})
	}

	return result
}

func isAllTarget(targets []string) bool {
	return len(targets) == 1 && strings.EqualFold(strings.TrimSpace(targets[0]), "all")
}

func findNode(nodes []model.Node, nodeID string) (model.Node, bool) {
	for _, node := range nodes {
		if node.ID == nodeID {
			return node, true
		}
	}

	return model.Node{}, false
}
