package hooks

import (
	"encoding/json"
	"fmt"
	"maps"
)

// hookEntry is a command handler in Claude and Codex's nested hook format.
// Unknown fields are retained so installing Loom cannot erase another tool's
// timeout, status message, or other options.
type hookEntry struct {
	Type    string
	Command string
	extra   map[string]json.RawMessage
}

func (h hookEntry) MarshalJSON() ([]byte, error) {
	fields := make(map[string]json.RawMessage, len(h.extra)+2)
	maps.Copy(fields, h.extra)

	typeJSON, err := json.Marshal(h.Type)
	if err != nil {
		return nil, err
	}
	fields["type"] = typeJSON

	commandJSON, err := json.Marshal(h.Command)
	if err != nil {
		return nil, err
	}
	fields["command"] = commandJSON

	return json.Marshal(fields)
}

func (h *hookEntry) UnmarshalJSON(data []byte) error {
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	if raw, ok := fields["type"]; ok {
		if err := json.Unmarshal(raw, &h.Type); err != nil {
			return fmt.Errorf("parsing hook type: %w", err)
		}
		delete(fields, "type")
	}
	if raw, ok := fields["command"]; ok {
		if err := json.Unmarshal(raw, &h.Command); err != nil {
			return fmt.Errorf("parsing hook command: %w", err)
		}
		delete(fields, "command")
	}

	if len(fields) > 0 {
		h.extra = fields
	}
	return nil
}

type hookMatcher struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

// cursorHookHandler is a handler in Cursor's flat per-event format. Unknown
// fields are retained for the same reason as hookEntry.extra.
type cursorHookHandler struct {
	Command string
	Matcher string
	extra   map[string]json.RawMessage
}

func (h cursorHookHandler) MarshalJSON() ([]byte, error) {
	fields := make(map[string]json.RawMessage, len(h.extra)+2)
	maps.Copy(fields, h.extra)

	commandJSON, err := json.Marshal(h.Command)
	if err != nil {
		return nil, err
	}
	fields["command"] = commandJSON

	if h.Matcher != "" {
		matcherJSON, err := json.Marshal(h.Matcher)
		if err != nil {
			return nil, err
		}
		fields["matcher"] = matcherJSON
	}

	return json.Marshal(fields)
}

func (h *cursorHookHandler) UnmarshalJSON(data []byte) error {
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	if raw, ok := fields["command"]; ok {
		if err := json.Unmarshal(raw, &h.Command); err != nil {
			return fmt.Errorf("parsing hook command: %w", err)
		}
		delete(fields, "command")
	}
	if raw, ok := fields["matcher"]; ok {
		if err := json.Unmarshal(raw, &h.Matcher); err != nil {
			return fmt.Errorf("parsing hook matcher: %w", err)
		}
		delete(fields, "matcher")
	}

	if len(fields) > 0 {
		h.extra = fields
	}
	return nil
}

type hookWiring struct {
	event   string
	matcher string
	command string
}
