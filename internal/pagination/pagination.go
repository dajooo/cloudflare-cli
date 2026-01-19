package pagination

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	LimitFlag = "limit"
	PageFlag  = "page"
)

// Options holds pagination parameters
type Options struct {
	Limit int
	Page  int
}

// PageInfo provides pagination metadata for display
type PageInfo struct {
	Page    int
	Limit   int
	Total   int
	Showing int
	HasMore bool
}

// TotalPages returns the total number of pages
func (p PageInfo) TotalPages() int {
	if p.Limit <= 0 || p.Total == 0 {
		return 1
	}
	pages := p.Total / p.Limit
	if p.Total%p.Limit > 0 {
		pages++
	}
	return pages
}

// FooterMessage returns a formatted pagination footer message
func (p PageInfo) FooterMessage(itemName string) string {
	if p.Limit <= 0 {
		// No pagination - show all
		return fmt.Sprintf("Showing %d %s", p.Total, itemName)
	}
	// Paginated - show page info
	totalPages := p.TotalPages()
	return fmt.Sprintf("Showing %d of %d %s (page %d/%d)", p.Showing, p.Total, itemName, p.Page, totalPages)
}

// RegisterFlags adds --limit and --page flags to a command
func RegisterFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().IntP(LimitFlag, "l", 0, "Maximum items to show (0 = all)")
	cmd.PersistentFlags().IntP(PageFlag, "p", 1, "Page number when limit is set")
}

// GetOptions extracts pagination options from command flags
func GetOptions(cmd *cobra.Command) Options {
	limit, _ := cmd.Flags().GetInt(LimitFlag)
	page, _ := cmd.Flags().GetInt(PageFlag)
	if page < 1 {
		page = 1
	}
	return Options{Limit: limit, Page: page}
}

// Paginate applies pagination to a slice and returns the paginated items with metadata
func Paginate[T any](items []T, opts Options) ([]T, PageInfo) {
	total := len(items)
	info := PageInfo{
		Page:  opts.Page,
		Limit: opts.Limit,
		Total: total,
	}

	// No limit = return all
	if opts.Limit <= 0 {
		info.Showing = total
		info.HasMore = false
		return items, info
	}

	start := (opts.Page - 1) * opts.Limit
	if start >= total {
		info.Showing = 0
		info.HasMore = false
		return []T{}, info
	}

	end := start + opts.Limit
	if end > total {
		end = total
	}

	info.Showing = end - start
	info.HasMore = end < total

	return items[start:end], info
}
