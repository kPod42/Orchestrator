package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type CommandKind string

const (
	KindEmpty  CommandKind = "empty"
	KindHelp   CommandKind = "help"
	KindExit   CommandKind = "exit"
	KindNodes  CommandKind = "nodes"
	KindAction CommandKind = "action"
)

type ParsedCommand struct {
	Kind       CommandKind
	ActionName string
	Targets    []string
	Args       map[string]string
	Raw        string
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseCommand(line string) (ParsedCommand, error) {
	raw := line
	line = strings.TrimSpace(line)

	if line == "" {
		return ParsedCommand{
			Kind: KindEmpty,
			Raw:  raw,
		}, nil
	}
	parts, err := splitCommandLine(line)
	if err != nil {
		return ParsedCommand{}, err
	}
	if len(parts) == 0 {
		return ParsedCommand{
			Kind: KindEmpty,
			Raw:  raw,
		}, nil
	}
	name := strings.ToLower(parts[0])
	switch name {
	case "help":
		return ParsedCommand{
			Kind: KindHelp,
			Raw:  raw,
		}, nil
	case "exit", "quit":
		return ParsedCommand{
			Kind: KindExit,
			Raw:  raw,
		}, nil
	case "nodes":
		return ParsedCommand{
			Kind: KindNodes,
			Raw:  raw,
		}, nil
	case "action":
		return parseAction(parts[1:], raw)
	default:
		return ParsedCommand{}, fmt.Errorf("unknown command: %s", parts[0])
	}
}

func parseAction(parts []string, raw string) (ParsedCommand, error) {
	if len(parts) < 3 {
		return ParsedCommand{}, fmt.Errorf("usage: action <name> on <target> [key=value...]")
	}

	actionName := strings.TrimSpace(parts[0])
	if actionName == "" {
		return ParsedCommand{}, fmt.Errorf("action name is empty")
	}

	if strings.ToLower(parts[1]) != "on" {
		return ParsedCommand{}, fmt.Errorf("expected keyword: on")
	}

	targets, err := parseTargets(parts[2])
	if err != nil {
		return ParsedCommand{}, err
	}

	args, err := parseKeyValueArgs(parts[3:])
	if err != nil {
		return ParsedCommand{}, err
	}

	return ParsedCommand{
		Kind:       KindAction,
		ActionName: actionName,
		Targets:    targets,
		Args:       args,
		Raw:        raw,
	}, nil
}

func parseTargets(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("target is empty")
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("targets are empty")
	}

	return result, nil
}

func parseKeyValueArgs(parts []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid argument %q, expected key=value", part)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("argument key is empty")
		}

		result[key] = value
	}

	return result, nil
}

func splitCommandLine(input string) ([]string, error) {
	var result []string
	var current strings.Builder

	var quote rune
	escaped := false

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}

			current.WriteRune(r)
			continue
		}

		if r == '"' || r == '\'' {
			quote = r
			continue
		}

		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}

			continue
		}

		current.WriteRune(r)
	}

	if escaped {
		current.WriteRune('\\')
	}

	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote")
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result, nil
}
