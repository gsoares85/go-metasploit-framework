package payloads

import "fmt"

func CreatePayload(args []string) {
	if len(args) == 0 {
		fmt.Println("Please provide a payload name example, payload/windows/meterpreter/reverse_tcp")
		return
	}

	payloadType := args[0]

	switch payloadType {
	case "payload/windows/meterpreter/reverse_tcp":
		fmt.Println("Creating payload for windows/meterpreter/reverse_tcp")
	default:
		fmt.Println("Payload type not supported")
	}
}
