package handler

type CategoryParam struct {
	Name string `json:"name"`
}

type CategoryListParam struct {
	CategoryId int64 `json:"category_id"`
	LastId     int64 `json:"last_id"`
}

type CreatePostItemParam struct {
	ObjectId int64 `json:"object_id"`
	Index    int32 `json:"index"`
}

type CreatePostParam struct {
	CategoryId  int64                  `json:"category_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	PostItems   []*CreatePostItemParam `json:"post_items"`
}
