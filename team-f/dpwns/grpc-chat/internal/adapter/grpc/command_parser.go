package grpc

import "strings"

func ParseCommand(input string) (Command, bool) {
	input = strings.TrimSpace(input)

	if input == "list" {
		return Command{Type: "list"}, true
	}
	if strings.HasPrefix(input, "connect ") {
		arg := strings.TrimPrefix(input, "connect ")
		return Command{Type: "connect", Argument: strings.TrimSpace(arg)}, true
	}
	return Command{}, false
}

