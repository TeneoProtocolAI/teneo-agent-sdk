package deploy

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UnmarshalJSON accepts legacy parameter metadata and normalizes it to the
// canonical SDK shape: type "enum" with values in "options".
func (p *CommandParameter) UnmarshalJSON(data []byte) error {
	type commandParameterAlias CommandParameter
	type commandParameterInput struct {
		commandParameterAlias
		LegacyEnum    []string `json:"enum"`
		LegacyChoices []string `json:"choices"`
	}

	var input commandParameterInput
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	*p = CommandParameter(input.commandParameterAlias)
	normalizeCommandParameterInPlace(p)

	if len(p.Options) == 0 {
		switch {
		case len(input.LegacyChoices) > 0:
			p.Options = append([]string(nil), input.LegacyChoices...)
		case len(input.LegacyEnum) > 0:
			p.Options = append([]string(nil), input.LegacyEnum...)
		}
	}

	return nil
}

func normalizeCommandParameterInPlace(param *CommandParameter) {
	if strings.EqualFold(param.Type, "choice") {
		param.Type = ParamTypeEnum
	}
}

func normalizeCommandsInPlace(commands []Command) {
	for i := range commands {
		for j := range commands[i].Parameters {
			normalizeCommandParameterInPlace(&commands[i].Parameters[j])
		}
		for j := range commands[i].Variants {
			for k := range commands[i].Variants[j].Parameters {
				normalizeCommandParameterInPlace(&commands[i].Variants[j].Parameters[k])
			}
		}
	}
}

func normalizeCommandsJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var commands []Command
	if err := json.Unmarshal(raw, &commands); err != nil {
		return nil, fmt.Errorf("invalid commands JSON: %w", err)
	}

	normalizeCommandsInPlace(commands)

	normalized, err := json.Marshal(commands)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize commands JSON: %w", err)
	}

	return normalized, nil
}
