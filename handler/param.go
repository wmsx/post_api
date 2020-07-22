package handler

type CategoryParam struct {
	Name     string `json:"name" binding:"required"`
	ShowName string `json:"show_name" binding:"required"`
}

type GetPostListParam struct {
	CategoryId int64 `json:"category_id" binding:"required"`
	LastId     int64 `json:"last_id"`
}

type CreatePostItemParam struct {
	ObjectId int64  `json:"object_id" binding:"required"`
	Index    int32  `json:"index" binding:"required"`
	Filename string `json:"filename" binding:"required"`
}

type CreatePostParam struct {
	CategoryId  int64                  `json:"category_id" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	PostItems   []*CreatePostItemParam `json:"post_items" binding:"required,dive"`
}
