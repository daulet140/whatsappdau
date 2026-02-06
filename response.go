package whatsappdau

type MessageResponse struct {
	MessagingProduct string     `json:"messaging_product"`
	Contacts         []Contacts `json:"contacts"`
	Messages         []Messages `json:"messages"`
}

type Contacts struct {
	Input string `json:"input"`
	WaId  string `json:"wa_id"`
}

type Messages struct {
	Id string `json:"id"`
}
