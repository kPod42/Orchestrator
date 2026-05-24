package presence

import (
	"fmt"
	"sort"
	"strings"

	"Orch/internal/agent/config"
)

func coordinatorGRPCEndpoints(cfg *config.Config, resp *registerResponse) []config.Endpoint {
	var endpoints []config.Endpoint

	// 1. Основной источник — ответ координатора после регистрации.
	for _, ep := range resp.CoordinatorEndpoints {
		if !isGRPCEndpoint(ep) {
			continue
		}

		endpoints = append(endpoints, ep)
	}
	// 2. Fallback из конфига агента.
	for _, ep := range cfg.Coordinator.Endpoints {
		if !isGRPCEndpoint(ep) {
			continue
		}
		endpoints = append(endpoints, ep)
	}
	// 3. Legacy fallback.
	if strings.TrimSpace(resp.GRPCAddress) != "" {
		endpoints = append(endpoints, config.Endpoint{
			Name:     "legacy-response",
			Kind:     "grpc",
			Address:  resp.GRPCAddress,
			Scope:    "legacy",
			Priority: 1000,
		})
	}

	endpoints = dedupeEndpoints(endpoints)

	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Priority < endpoints[j].Priority
	})
	return endpoints
}

func isGRPCEndpoint(endpoint config.Endpoint) bool {
	return strings.EqualFold(strings.TrimSpace(endpoint.Kind), "grpc") &&
		strings.TrimSpace(endpoint.Address) != ""
}

func dedupeEndpoints(endpoints []config.Endpoint) []config.Endpoint {
	seen := make(map[string]struct{})
	result := make([]config.Endpoint, 0, len(endpoints))

	for _, ep := range endpoints {
		key := strings.ToLower(strings.TrimSpace(ep.Kind)) + "|" + strings.TrimSpace(ep.Address)

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, ep)
	}

	return result
}

func endpointLabel(ep config.Endpoint) string {
	if ep.Name != "" {
		return fmt.Sprintf("%s/%s/%s", ep.Name, ep.Kind, ep.Address)
	}

	return fmt.Sprintf("%s/%s", ep.Kind, ep.Address)
}
