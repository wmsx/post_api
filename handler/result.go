package handler

type PostInfo struct {
	Id          int64
	Type        int32
	Title       string
	Description string
	MengerInfo  *MengerInfo
	Item        []*PostItem
}

type MengerInfo struct {
	MengerId int64  `json:"menger_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

type PostItem struct {
	Url string `json:"url"`
}
