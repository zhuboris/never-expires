package mailmsg

type Message struct {
	Recipient string `json:"recipient"`
	Email     []byte `json:"email"`
}
