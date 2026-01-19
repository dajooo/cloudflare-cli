package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Colors struct {
	Gray50  lipgloss.AdaptiveColor
	Gray100 lipgloss.AdaptiveColor
	Gray200 lipgloss.AdaptiveColor
	Gray300 lipgloss.AdaptiveColor
	Gray400 lipgloss.AdaptiveColor
	Gray500 lipgloss.AdaptiveColor
	Gray600 lipgloss.AdaptiveColor
	Gray700 lipgloss.AdaptiveColor
	Gray800 lipgloss.AdaptiveColor
	Gray900 lipgloss.AdaptiveColor

	Primary50  lipgloss.AdaptiveColor
	Primary100 lipgloss.AdaptiveColor
	Primary200 lipgloss.AdaptiveColor
	Primary300 lipgloss.AdaptiveColor
	Primary400 lipgloss.AdaptiveColor
	Primary500 lipgloss.AdaptiveColor
	Primary600 lipgloss.AdaptiveColor
	Primary700 lipgloss.AdaptiveColor

	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Error   lipgloss.AdaptiveColor
}

var C = Colors{
	Gray50:  lipgloss.AdaptiveColor{Light: "#fafbfc", Dark: "#0d1117"},
	Gray100: lipgloss.AdaptiveColor{Light: "#f4f6f8", Dark: "#161b22"},
	Gray200: lipgloss.AdaptiveColor{Light: "#e1e8ed", Dark: "#21262d"},
	Gray300: lipgloss.AdaptiveColor{Light: "#c1ccd7", Dark: "#30363d"},
	Gray400: lipgloss.AdaptiveColor{Light: "#8896a6", Dark: "#656d76"},
	Gray500: lipgloss.AdaptiveColor{Light: "#6b7785", Dark: "#8b949e"},
	Gray600: lipgloss.AdaptiveColor{Light: "#4a5663", Dark: "#c9d1d9"},
	Gray700: lipgloss.AdaptiveColor{Light: "#2d3843", Dark: "#f0f6fc"},
	Gray800: lipgloss.AdaptiveColor{Light: "#1a2027", Dark: "#f4f6f8"},
	Gray900: lipgloss.AdaptiveColor{Light: "#0d1117", Dark: "#fafbfc"},

	Primary50:  lipgloss.AdaptiveColor{Light: "#f0f7ff", Dark: "#0c1821"},
	Primary100: lipgloss.AdaptiveColor{Light: "#d9eaff", Dark: "#1a2332"},
	Primary200: lipgloss.AdaptiveColor{Light: "#b3d9ff", Dark: "#2d3748"},
	Primary300: lipgloss.AdaptiveColor{Light: "#80c5ff", Dark: "#4299e1"},
	Primary400: lipgloss.AdaptiveColor{Light: "#4da6ff", Dark: "#63b3ed"},
	Primary500: lipgloss.AdaptiveColor{Light: "#1a85ff", Dark: "#3182ce"},
	Primary600: lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#2c5282"},
	Primary700: lipgloss.AdaptiveColor{Light: "#004d99", Dark: "#2a4365"},

	Success: lipgloss.AdaptiveColor{Light: "#16a34a", Dark: "#22c55e"},
	Warning: lipgloss.AdaptiveColor{Light: "#d97706", Dark: "#f59e0b"},
	Error:   lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#ef4444"},
}

type Symbols struct {
	ArrowRight string
	ArrowDown  string
	ArrowUp    string
	Check      string
	Cross      string
	Dot        string
	Circle     string
	Dash       string

	CornerTL string
	CornerTR string
	CornerBL string
	CornerBR string
	Line     string
	Pipe     string
	Branch   string
	End      string
	Spinner  []string
}

var S = Symbols{
	ArrowRight: "→",
	ArrowDown:  "↓",
	ArrowUp:    "↑",
	Check:      "✓",
	Cross:      "✗",
	Dot:        "•",
	Circle:     "○",
	Dash:       "–",

	CornerTL: "╭",
	CornerTR: "╮",
	CornerBL: "╰",
	CornerBR: "╯",
	Line:     "─",
	Pipe:     "│",
	Branch:   "├",
	End:      "└",
	Spinner:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
}

var (
	H1 = lipgloss.NewStyle().
		Foreground(C.Gray800).
		Bold(true).
		MarginBottom(1)

	H2 = lipgloss.NewStyle().
		Foreground(C.Gray700).
		Bold(true).
		MarginBottom(1)

	H3 = lipgloss.NewStyle().
		Foreground(C.Gray700).
		Bold(false).
		MarginBottom(1)

	Body = lipgloss.NewStyle().
		Foreground(C.Gray700)

	BodyMuted = lipgloss.NewStyle().
			Foreground(C.Gray600)

	BodySmall = lipgloss.NewStyle().
			Foreground(C.Gray500)

	Link = lipgloss.NewStyle().
		Foreground(C.Primary500).
		Underline(true)

	Code = lipgloss.NewStyle().
		Foreground(C.Gray700).
		Background(C.Gray100).
		Padding(0, 1)

	CodeBlock = lipgloss.NewStyle().
			Foreground(C.Gray700).
			Background(C.Gray100).
			Padding(1, 2).
			MarginBottom(1)
)

var (
	ButtonPrimary = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"}).
			Background(C.Primary500).
			Padding(0, 3).
			Margin(0, 1)

	ButtonSecondary = lipgloss.NewStyle().
			Foreground(C.Primary500).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(C.Primary500).
			Padding(0, 3).
			Margin(0, 1)

	ButtonGhost = lipgloss.NewStyle().
			Foreground(C.Gray600).
			Padding(0, 1).
			Margin(0, 1)

	Card = lipgloss.NewStyle().
		Background(C.Gray50).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(C.Gray200).
		Padding(2, 3).
		MarginBottom(1)

	Panel = lipgloss.NewStyle().
		Background(C.Gray100).
		Padding(1, 2).
		MarginBottom(1)

	StatusSuccess = lipgloss.NewStyle().
			Foreground(C.Success).
			Bold(true)

	StatusWarning = lipgloss.NewStyle().
			Foreground(C.Warning).
			Bold(true)

	StatusError = lipgloss.NewStyle().
			Foreground(C.Error).
			Bold(true)

	Badge = lipgloss.NewStyle().
		Background(C.Gray200).
		Foreground(C.Gray700).
		Padding(0, 1)

	BadgeSuccess = lipgloss.NewStyle().
			Background(C.Success).
			Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"}).
			Padding(0, 1)

	BadgePrimary = lipgloss.NewStyle().
			Background(C.Primary500).
			Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#000000"}).
			Padding(0, 1)
)

var (
	Container = lipgloss.NewStyle().
			Padding(1, 2)

	Stack = lipgloss.NewStyle().
		MarginBottom(1)

	Flex = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Left)

	Center = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center)

	Right = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Right)
)

func Title(text string) string {
	return H1.Render(text)
}

func Subtitle(text string) string {
	return H2.Render(text)
}

func Text(text string) string {
	return Body.Render(text)
}

func Muted(text string) string {
	return BodyMuted.Render(text)
}

func Small(text string) string {
	return BodySmall.Render(text)
}

func Success(text string) string {
	return StatusSuccess.Render(S.Check + " " + text)
}

func Warning(text string) string {
	return StatusWarning.Render("⚠ " + text)
}

func Error(text string) string {
	return StatusError.Render(S.Cross + " " + text)
}

func Info(text string) string {
	return lipgloss.NewStyle().
		Foreground(C.Primary500).
		Render("ⓘ " + text)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ErrorMessage(title string, err ...error) string {
	var b strings.Builder

	b.WriteString(StatusError.Render(S.Cross + " " + title))

	if len(err) > 0 && err[0] != nil {
		errorMsg := cleanErrorMessage(err[0].Error())
		if errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(BodyMuted.Render(errorMsg))
		}
	}

	return b.String()
}

func WarningMessage(title string, err ...error) string {
	var b strings.Builder

	b.WriteString(StatusWarning.Render("⚠ " + title))

	if len(err) > 0 && err[0] != nil {
		errorMsg := cleanErrorMessage(err[0].Error())
		if errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(BodyMuted.Render(errorMsg))
		}
	}

	return b.String()
}

func ErrorBox(title string, err ...error) string {
	content := ErrorMessage(title, err...)
	return Box(content)
}

func cleanErrorMessage(errStr string) string {
	if strings.Contains(errStr, `"message":`) {
		if start := strings.Index(errStr, `"message":"`); start != -1 {
			start += 11
			if end := strings.Index(errStr[start:], `"`); end != -1 {
				return errStr[start : start+end]
			}
		}
	}

	if strings.Contains(errStr, "400 Bad Request") {
		return "Invalid request - please check your credentials"
	}
	if strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") {
		return "Authentication failed - invalid credentials"
	}
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "Forbidden") {
		return "Access denied - insufficient permissions"
	}
	if strings.Contains(errStr, "404") || strings.Contains(errStr, "Not Found") {
		return "Resource not found"
	}
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "Too Many Requests") {
		return "Rate limit exceeded - please try again later"
	}
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "Internal Server Error") {
		return "Server error - please try again later"
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "Timeout") {
		return "Request timed out - please check your connection"
	}
	if strings.Contains(errStr, "connection") && strings.Contains(errStr, "refused") {
		return "Cannot connect to server - please check your network"
	}

	if strings.Contains(strings.ToLower(errStr), "invalid") &&
		(strings.Contains(strings.ToLower(errStr), "token") ||
			strings.Contains(strings.ToLower(errStr), "key") ||
			strings.Contains(strings.ToLower(errStr), "credential")) {
		return "Invalid credentials - please check your API token or key"
	}

	return errStr
}

func BulletList(items []string) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(BodyMuted.Render(S.Dot+" ") + Body.Render(item) + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func NumberedList(items []string) string {
	var b strings.Builder
	for i, item := range items {
		b.WriteString(BodyMuted.Render(fmt.Sprintf("%d. ", i+1)) + Body.Render(item) + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func ProgressBar(current, total int, width int) string {
	if width <= 0 {
		width = 20
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return lipgloss.NewStyle().
		Foreground(C.Primary500).
		Render(bar) + " " +
		BodyMuted.Render(fmt.Sprintf("%d/%d", current, total))
}

func Divider(width int) string {
	if width <= 0 {
		width = 40
	}
	return lipgloss.NewStyle().
		Foreground(C.Gray200).
		Render(strings.Repeat(S.Line, width))
}

func DividerWithText(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	textLen := len(text) + 2
	lineLen := (width - textLen) / 2

	if lineLen < 0 {
		lineLen = 0
	}

	leftLine := strings.Repeat(S.Line, lineLen)
	rightLine := strings.Repeat(S.Line, width-lineLen-textLen)

	return lipgloss.NewStyle().Foreground(C.Gray200).Render(leftLine) +
		" " + BodyMuted.Render(text) + " " +
		lipgloss.NewStyle().Foreground(C.Gray200).Render(rightLine)
}

func HuhTheme() *huh.Theme {
	theme := huh.ThemeBase()

	theme.Focused.Title = H2
	theme.Focused.Description = BodyMuted
	theme.Focused.ErrorMessage = StatusError
	theme.Focused.Directory = Body
	theme.Focused.File = Body
	theme.Focused.Option = Body
	theme.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(C.Primary500)
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(C.Primary500).Bold(true)
	theme.Focused.UnselectedOption = Body
	theme.Focused.FocusedButton = ButtonPrimary
	theme.Focused.BlurredButton = ButtonSecondary
	theme.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(C.Primary500)
	theme.Focused.TextInput.Placeholder = BodyMuted
	theme.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(C.Primary500)
	theme.Focused.TextInput.Text = Body

	theme.Blurred.Title = BodyMuted
	theme.Blurred.Description = BodyMuted.Faint(true)
	theme.Blurred.ErrorMessage = StatusError.Faint(true)
	theme.Blurred.Directory = BodyMuted
	theme.Blurred.File = BodyMuted
	theme.Blurred.Option = BodyMuted
	theme.Blurred.MultiSelectSelector = BodyMuted
	theme.Blurred.SelectedOption = BodyMuted
	theme.Blurred.UnselectedOption = BodyMuted
	theme.Blurred.FocusedButton = ButtonGhost
	theme.Blurred.BlurredButton = ButtonGhost
	theme.Blurred.TextInput.Cursor = BodyMuted
	theme.Blurred.TextInput.Placeholder = BodyMuted.Faint(true)
	theme.Blurred.TextInput.Prompt = BodyMuted
	theme.Blurred.TextInput.Text = BodyMuted

	return theme
}

func StyledTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(C.Primary500)
	ti.PlaceholderStyle = BodyMuted
	ti.TextStyle = Body
	return ti
}

func StyledSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: S.Spinner,
		FPS:    10,
	}
	s.Style = lipgloss.NewStyle().Foreground(C.Primary500)
	return s
}

func Box(content string, title ...string) string {
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 80
	}

	originalLines := strings.Split(content, "\n")
	longestLineLen := 0
	for _, line := range originalLines {
		if width := lipgloss.Width(line); width > longestLineLen {
			longestLineLen = width
		}
	}

	const boxOverhead = 4
	const terminalMargin = 2
	maxAllowedContentWidth := termWidth - boxOverhead - terminalMargin
	if maxAllowedContentWidth < 1 {
		maxAllowedContentWidth = 1
	}

	finalContentWidth := longestLineLen
	if finalContentWidth > maxAllowedContentWidth {
		finalContentWidth = maxAllowedContentWidth
	}

	contentWrapper := lipgloss.NewStyle().Width(finalContentWidth)
	wrappedContent := contentWrapper.Render(content)
	lines := strings.Split(wrappedContent, "\n")

	totalBoxWidth := finalContentWidth + 2

	var titleStr string
	var titleLen int
	const leftTitleDashes = 2
	const rightTitleDashes = 2

	if len(title) > 0 && title[0] != "" {
		titleStr = " " + title[0] + " "
		titleLen = lipgloss.Width(titleStr)

		minTitleBarWidth := titleLen + leftTitleDashes + rightTitleDashes

		if minTitleBarWidth > totalBoxWidth {
			totalBoxWidth = minTitleBarWidth
		}
	}

	var b strings.Builder
	borderStyle := lipgloss.NewStyle().Foreground(C.Gray500)
	titleStyle := lipgloss.NewStyle().Foreground(C.Primary500)

	if titleLen > 0 {
		rightLen := totalBoxWidth - titleLen - leftTitleDashes
		if rightLen < 0 {
			rightLen = 0
		}

		b.WriteString(borderStyle.Render(S.CornerTL + strings.Repeat(S.Line, leftTitleDashes)))
		b.WriteString(titleStyle.Render(titleStr))
		b.WriteString(borderStyle.Render(strings.Repeat(S.Line, rightLen) + S.CornerTR))
	} else {
		b.WriteString(borderStyle.Render(S.CornerTL + strings.Repeat(S.Line, totalBoxWidth) + S.CornerTR))
	}
	b.WriteString("\n")

	for _, line := range lines {
		padding := totalBoxWidth - lipgloss.Width(line) - 2
		if padding < 0 {
			padding = 0
		}
		b.WriteString(borderStyle.Render(S.Pipe))
		b.WriteString(" " + line + strings.Repeat(" ", padding) + " ")
		b.WriteString(borderStyle.Render(S.Pipe))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(S.CornerBL + strings.Repeat(S.Line, totalBoxWidth) + S.CornerBR))

	return b.String()
}

func Confirm(prompt string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()
	return confirmed, err
}

func FangTheme() fang.ColorScheme {
	errorFg := lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#1a2027"}

	return fang.ColorScheme{
		Base:           C.Gray700,
		Title:          C.Primary500,
		Description:    C.Gray600,
		Codeblock:      C.Gray100,
		Program:        C.Primary400,
		DimmedArgument: C.Gray400,
		Comment:        C.Gray500,
		Flag:           C.Warning,
		FlagDefault:    C.Gray500,
		Command:        C.Success,
		QuotedString:   C.Success,
		Argument:       C.Gray700,
		Help:           C.Gray600,
		Dash:           C.Gray400,
		ErrorHeader:    [2]color.Color{errorFg, C.Error},
		ErrorDetails:   C.Error,
	}
}
