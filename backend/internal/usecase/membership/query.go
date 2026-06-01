package membership

import "errors"

type GetMembershipQuery struct {
	ID string
}

type ListMembershipQuery struct {
	IdolID   *string `form:"idol_id"`
	GroupID  *string `form:"group_id"`
	IsActive *bool   `form:"is_active"`
	Role     *string `form:"role"`
	Sort     *string `form:"sort"`
	Order    *string `form:"order"`
	Page     *int    `form:"page"`
	Limit    *int    `form:"limit"`
}

func (q *ListMembershipQuery) Normalize() {
	if q.Page == nil || *q.Page < 1 {
		p := 1
		q.Page = &p
	}
	if q.Limit == nil || *q.Limit < 1 {
		l := 20
		q.Limit = &l
	}
	if *q.Limit > 100 {
		l := 100
		q.Limit = &l
	}
	if q.Sort == nil {
		s := "created_at"
		q.Sort = &s
	}
	if q.Order == nil {
		o := "desc"
		q.Order = &o
	}
}

func (q *ListMembershipQuery) Validate() error {
	if q.Sort != nil {
		allowed := []string{"joined_at", "left_at", "created_at"}
		if !containsStr(allowed, *q.Sort) {
			return errors.New("無効なソート項目です")
		}
	}
	if q.Order != nil {
		if *q.Order != "asc" && *q.Order != "desc" {
			return errors.New("無効なソート順です")
		}
	}
	return nil
}

func containsStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type MembershipDTO struct {
	ID        string  `json:"id"`
	IdolID    string  `json:"idol_id"`
	GroupID   string  `json:"group_id"`
	Role      string  `json:"role"`
	JoinedAt  *string `json:"joined_at,omitempty"`
	LeftAt    *string `json:"left_at,omitempty"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type MembershipSearchResult struct {
	Data []*MembershipDTO `json:"data"`
	Meta *PaginationMeta  `json:"meta"`
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
