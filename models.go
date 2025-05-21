package whatsappdau

type AudioMessage struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Audio            struct {
		ID string `json:"id"` // Media ID from /media upload
	} `json:"audio"`
}

type ImageMessage struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Image            struct {
		ID string `json:"id"`
	} `json:"image"`
}

type LocationMessage struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Location         struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name,omitempty"`
		Address   string  `json:"address,omitempty"`
	} `json:"location"`
}

type WhatsAppMessage struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Interactive      interface{} `json:"interactive"` // Can be ListInteractive or ButtonsInteractive
}

type ListInteractive struct {
	Type   string     `json:"type"`
	Body   BodyText   `json:"body"`
	Action ListAction `json:"action"`
}

type ListAction struct {
	Button   string        `json:"button"`
	Sections []ListSection `json:"sections"`
}

type ListSection struct {
	Title string     `json:"title,omitempty"`
	Rows  []ListItem `json:"rows"`
}

type BodyText struct {
	Text string `json:"text"`
}

type ButtonItem struct {
	Text string `json:"text"`
	ID   string `json:"id"`
	Link string `json:"link,omitempty"`
	Type string `json:"type,omitempty"`
}

type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ButtonAction struct {
	Name       string        `json:"name,omitempty"`
	Parameters *Parameters   `json:"parameters,omitempty"`
	Buttons    []interface{} `json:"buttons,omitempty"`
}

type URLButton struct {
	Name       string     `json:"name"`
	Parameters Parameters `json:"parameters"`
}
type Parameters struct {
	DisplayText string `json:"display_text"`
	Url         string `json:"url"`
}

type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Url   string `json:"url,omitempty"`
	Text  string `json:"text,omitempty"`
}

type ButtonsInteractive struct {
	Type   string       `json:"type"`
	Body   BodyText     `json:"body"`
	Action ButtonAction `json:"action,omitempty"`
}
