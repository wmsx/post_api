package handler

type PostInfo struct {
	Id          int64       `json:"id"`
	Type        int32       `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	MengerInfo  *MengerInfo `json:"menger"`
	Items       []*PostItem `json:"items"`
	CreateAt    int64
}

type MengerInfo struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

type PostItem struct {
	Url string `json:"url"`
}
