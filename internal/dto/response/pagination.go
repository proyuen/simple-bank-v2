package response

// PaginationResponse 分页响应
// 通用的分页信息结构
type PaginationResponse struct {
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页条数
	TotalCount int64 `json:"total_count"` // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// NewPaginationResponse 创建分页响应
// 参数:
//   - page: 当前页码
//   - pageSize: 每页条数
//   - totalCount: 总记录数
func NewPaginationResponse(page, pageSize int, totalCount int64) PaginationResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

// ListResponse 通用列表响应
// T 可以是任意类型的数据切片
type ListResponse[T any] struct {
	Data       []T                `json:"data"`       // 数据列表
	Pagination PaginationResponse `json:"pagination"` // 分页信息
}

// NewListResponse 创建列表响应
func NewListResponse[T any](data []T, page, pageSize int, totalCount int64) ListResponse[T] {
	return ListResponse[T]{
		Data:       data,
		Pagination: NewPaginationResponse(page, pageSize, totalCount),
	}
}
