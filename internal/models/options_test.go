package models

import (
	"testing"
)

func TestListOptionsValidation(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListOptions
		expected *ListOptions
	}{
		{
			name: "default values",
			opts: &ListOptions{},
			expected: &ListOptions{
				Page:      1,
				PerPage:   30,
				SortBy:    "created_at",
				SortOrder: "desc",
				Filters:   make(map[string]interface{}),
			},
		},
		{
			name: "custom values within limits",
			opts: &ListOptions{
				Page:      2,
				PerPage:   50,
				SortBy:    "name",
				SortOrder: "asc",
			},
			expected: &ListOptions{
				Page:      2,
				PerPage:   50,
				SortBy:    "name",
				SortOrder: "asc",
				Filters:   make(map[string]interface{}),
			},
		},
		{
			name: "per_page exceeds limit",
			opts: &ListOptions{
				PerPage: 150,
			},
			expected: &ListOptions{
				Page:      1,
				PerPage:   100, // Should be capped at 100
				SortBy:    "created_at",
				SortOrder: "desc",
				Filters:   make(map[string]interface{}),
			},
		},
		{
			name: "invalid sort order",
			opts: &ListOptions{
				SortOrder: "invalid",
			},
			expected: &ListOptions{
				Page:      1,
				PerPage:   30,
				SortBy:    "created_at",
				SortOrder: "desc", // Should default to desc
				Filters:   make(map[string]interface{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v", err)
				return
			}

			if tt.opts.Page != tt.expected.Page {
				t.Errorf("Page = %v, want %v", tt.opts.Page, tt.expected.Page)
			}
			if tt.opts.PerPage != tt.expected.PerPage {
				t.Errorf("PerPage = %v, want %v", tt.opts.PerPage, tt.expected.PerPage)
			}
			if tt.opts.SortBy != tt.expected.SortBy {
				t.Errorf("SortBy = %v, want %v", tt.opts.SortBy, tt.expected.SortBy)
			}
			if tt.opts.SortOrder != tt.expected.SortOrder {
				t.Errorf("SortOrder = %v, want %v", tt.opts.SortOrder, tt.expected.SortOrder)
			}
		})
	}
}

func TestListOptionsPagination(t *testing.T) {
	opts := &ListOptions{
		Page:    3,
		PerPage: 20,
	}

	if opts.GetPage() != 3 {
		t.Errorf("GetPage() = %v, want 3", opts.GetPage())
	}
	if opts.GetPerPage() != 20 {
		t.Errorf("GetPerPage() = %v, want 20", opts.GetPerPage())
	}
	if opts.GetLimit() != 20 {
		t.Errorf("GetLimit() = %v, want 20", opts.GetLimit())
	}
	if opts.GetOffset() != 40 { // (3-1) * 20 = 40
		t.Errorf("GetOffset() = %v, want 40", opts.GetOffset())
	}
}

func TestListResponse(t *testing.T) {
	// Mock data
	type MockItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	items := []MockItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
	}

	opts := &ListOptions{
		Page:    2,
		PerPage: 2,
	}

	total := int64(10) // Total items in database
	response := NewListResponse(items, total, opts)

	// Check response structure
	if len(response.Items) != 3 {
		t.Errorf("Items length = %v, want 3", len(response.Items))
	}
	if response.Total != 10 {
		t.Errorf("Total = %v, want 10", response.Total)
	}
	if response.Page != 2 {
		t.Errorf("Page = %v, want 2", response.Page)
	}
	if response.PerPage != 2 {
		t.Errorf("PerPage = %v, want 2", response.PerPage)
	}
	if response.LastPage != 5 { // 10 items / 2 per page = 5 pages
		t.Errorf("LastPage = %v, want 5", response.LastPage)
	}
	if !response.HasMore {
		t.Errorf("HasMore = %v, want true", response.HasMore)
	}
	if response.NextPage == nil || *response.NextPage != 3 {
		t.Errorf("NextPage = %v, want 3", response.NextPage)
	}
	if response.PrevPage == nil || *response.PrevPage != 1 {
		t.Errorf("PrevPage = %v, want 1", response.PrevPage)
	}
}

func TestListResponseLastPage(t *testing.T) {
	type MockItem struct {
		ID int `json:"id"`
	}

	items := []MockItem{{ID: 1}, {ID: 2}}
	opts := &ListOptions{
		Page:    2,
		PerPage: 2,
	}
	total := int64(4) // Exactly 2 pages

	response := NewListResponse(items, total, opts)

	if response.HasMore {
		t.Errorf("HasMore = %v, want false (last page)", response.HasMore)
	}
	if response.NextPage != nil {
		t.Errorf("NextPage = %v, want nil (last page)", response.NextPage)
	}
	if response.PrevPage == nil || *response.PrevPage != 1 {
		t.Errorf("PrevPage = %v, want 1", response.PrevPage)
	}
}
