package api

// 注意：ErrorResponse 和 SuccessResponse 已移至 common.go
// 此处保留注释以便于理解代码历史

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int64       `json:"total_pages"`
}