package protocol

import "strings"

func ParseCommand(line string) (*Command, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, nil
	}

	cmd := &Command{
		Name: strings.ToUpper(parts[0]),
		Args: parts[1:],
	}

	return cmd, nil
}
