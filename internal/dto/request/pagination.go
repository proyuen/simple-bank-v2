package request

// PaginationRequest 分页请求参数
// 可嵌入到其他请求结构体中使用
type PaginationRequest struct {
	// PageID 页码 (从1开始)
	PageID int `form:"page_id" binding:"required,min=1"`

	// PageSize 每页条数
	PageSize int `form:"page_size" binding:"required,min=5,max=100"`
}

// Offset 计算数据库查询的偏移量
// 例如: PageID=2, PageSize=10 → Offset=10
func (p *PaginationRequest) Offset() int {
	return (p.PageID - 1) * p.PageSize
}

// Limit 返回每页条数 (与 PageSize 相同，但命名更符合数据库习惯)
func (p *PaginationRequest) Limit() int {
	return p.PageSize
}
