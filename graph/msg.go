package graph

type msgTag struct {
	id     int
	passUp bool
}

func (m msgTag) ID() int      { return m.id }
func (m msgTag) PassUp() bool { return m.passUp }

type Msg interface {
	// ID is the id of the node sending the message.
	ID() int
	// PassUp is true when completed data is being
	// sent backwards from the child to the parent.
	PassUp() bool
	// SetID returns a copied message with
	// the id of the message changed.
	SetID(id int) Msg
}

type ValueMsg struct {
	msgTag

	// Data is the data contained
	// in the value message.
	Data interface{}
}

func (m ValueMsg) SetID(id int) Msg {
	m.id = id
	return m
}

type StreamMsg struct {
	msgTag

	// Idx is the index of the message
	// sent over the stream.
	Idx StreamIndex
	// Len is the total number of messages
	// sent over the stream.
	Len StreamIndex
	// Data is the data contained in the
	// stream message.
	Data interface{}
}

func (m StreamMsg) SetID(id int) Msg {
	m.id = id
	return m
}

type ErrMsg struct {
	msgTag

	// Err is the underlying err behind the message.
	Err error
}

func (m ErrMsg) SetID(id int) Msg {
	m.id = id
	return m
}

func NewValueMsg(id int, passUp bool, data interface{}) ValueMsg {
	return ValueMsg{msgTag{id, passUp}, data}
}

func NewStreamMsg(id int, passUp bool, idx StreamIndex, len StreamIndex, data interface{}) StreamMsg {
	return StreamMsg{msgTag{id, passUp}, idx, len, data}
}

func NewErrMsg(id int, passUp bool, err error) ErrMsg {
	return ErrMsg{msgTag{id, passUp}, err}
}

func BroadcastMsg(msg Msg, parentChans map[int]chan Msg) {
	for _, parentChan := range parentChans {
		parentChan <- msg
	}
}
