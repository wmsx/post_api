package handler

type PostInfo struct {
	Id          int64       `json:"id"`
	Type        int32       `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	MengerInfo  *MengerInfo `json:"menger_info"`
	Item        []*PostItem `json:"item"`
}

type MengerInfo struct {
	MengerId int64  `json:"menger_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

type PostItem struct {
	Url string `json:"url"`
}
