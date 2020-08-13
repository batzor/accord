package server

type Message struct {
	timestamp uint64
	from      string
	content   string
}

type Channel struct {
	channel_id          uint64
	messages            []Message
	msgc                chan Message
	users               []User
	subscription		[]
	pinned_msg          uint64
	is_public           bool
	rolesWithPermission map[string][]string
}

func NewChannel(uid uint64, users []User, is_public bool) {
	return &Channel{
		channel_id: uid,
		users:      users,
		msgc:       make(chan Message),
		is_public:  is_public,
	}
}

func (ch *Channel) Listen() {
	for {
		select {
		case msg := <-msgc:
			ch.messages.append(msg)
			ch.Broadcast(msg)
		}
	}
}

func (ch *Channel) Broadcast(msg Message) {
	return
}
