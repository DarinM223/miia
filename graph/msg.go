package graph

type msgTag struct {
	id     int
	passUp bool
}

func (m *msgTag) ID() int      { return m.id }
func (m *msgTag) PassUp() bool { return m.passUp }
func (m *msgTag) setID(id int) { m.id = id }

type Msg interface {
	// ID is the id of the node sending the message.
	ID() int
	// PassUp is true when completed data is being
	// sent backwards from the child to the parent.
	PassUp() bool

	// setID sets the id of the message.
	setID(id int)
}

type ValueMsg struct {
	*msgTag

	// Data is the data contained
	// in the value message.
	Data interface{}
}

type StreamMsg struct {
	*msgTag

	// Idx is the index of the message
	// sent over the stream.
	Idx int
	// Len is the total number of messages
	// sent over the stream.
	Len int
	// Data is the data contained in the
	// stream message.
	Data interface{}
}

type ErrMsg struct {
	*msgTag

	// Err is the underlying err behind the message.
	Err error
}

func NewValueMsg(id int, passUp bool, data interface{}) *ValueMsg {
	return &ValueMsg{&msgTag{id, passUp}, data}
}

func NewStreamMsg(id int, passUp bool, idx int, len int, data interface{}) *StreamMsg {
	return &StreamMsg{&msgTag{id, passUp}, idx, len, data}
}

func NewErrMsg(id int, passUp bool, err error) *ErrMsg {
	return &ErrMsg{&msgTag{id, passUp}, err}
}