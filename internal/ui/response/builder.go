package response

import (
	"fmt"
	"strings"

	"dario.lol/cf/internal/ui"
)

type Builder struct {
	title          string
	summary        map[string]any
	items          []string
	footerSuccess  string
	err            error
	errTitle       string
	noItemsMessage string
}

func New() *Builder {
	return &Builder{
		summary: make(map[string]any),
		items:   []string{},
	}
}

func (b *Builder) Title(title string) *Builder {
	b.title = title
	return b
}

func (b *Builder) Summary(key string, value any) *Builder {
	b.summary[key] = value
	return b
}

func (b *Builder) AddItem(title, content string) *Builder {
	b.items = append(b.items, ui.Box(content, title))
	return b
}

func (b *Builder) NoItemsMessage(message string) *Builder {
	b.noItemsMessage = message
	return b
}

func (b *Builder) FooterSuccess(format string, a ...any) *Builder {
	b.footerSuccess = fmt.Sprintf(format, a...)
	return b
}

func (b *Builder) Error(title string, err error) *Builder {
	b.errTitle = title
	b.err = err
	return b
}

func (b *Builder) Display() {
	if b.err != nil {
		fmt.Println(ui.ErrorMessage(b.errTitle, b.err))
		return
	}

	if b.title != "" {
		fmt.Println(ui.Title(b.title))
		fmt.Println()
	}

	if len(b.summary) > 0 {
		var summaryContent strings.Builder
		for key, val := range b.summary {
			summaryContent.WriteString(fmt.Sprintf("%-12s %v\n", key, val))
		}
		fmt.Println(ui.Box(strings.TrimSpace(summaryContent.String()), "Summary"))
		fmt.Println()
	}

	if len(b.items) == 0 {
		if b.noItemsMessage != "" {
			fmt.Println(ui.Warning(b.noItemsMessage))
		}
	} else {
		var itemsContent strings.Builder
		for _, item := range b.items {
			itemsContent.WriteString(item)
			itemsContent.WriteString("\n\n")
		}
		fmt.Print(itemsContent.String())
	}

	if b.footerSuccess != "" {
		fmt.Println(ui.Success(b.footerSuccess))
	}
}

type ItemContentBuilder struct {
	sb strings.Builder
}

func NewItemContent() *ItemContentBuilder {
	return &ItemContentBuilder{}
}

func (ic *ItemContentBuilder) Add(key, value string) *ItemContentBuilder {
	ic.sb.WriteString(fmt.Sprintf("%-12s %s\n", key, value))
	return ic
}

func (ic *ItemContentBuilder) AddRaw(content string) *ItemContentBuilder {
	ic.sb.WriteString(content)
	ic.sb.WriteString("\n")
	return ic
}

func (ic *ItemContentBuilder) String() string {
	return strings.TrimSpace(ic.sb.String())
}
