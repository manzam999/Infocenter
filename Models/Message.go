package Models

type Message struct {
	Id      int64  `json: "id"`
	Message string `json: "message"`
	Event   string `json: "event"`
}
