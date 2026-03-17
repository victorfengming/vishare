package status

// Msg carries connection state updates between components.
type Msg struct {
	Connected  bool
	ClientName string
}
