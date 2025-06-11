package style

import "github.com/charmbracelet/lipgloss"

var (
	// general
	Normal    = lipgloss.Color("#EEEEEE")
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	Base = lipgloss.NewStyle().Foreground(Normal)

	// tabs
	ActiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	TabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	Tab = lipgloss.NewStyle().
		Border(TabBorder, true).
		BorderForeground(Highlight).
		Padding(0, 1)

	ActiveTab = Tab.Border(ActiveTabBorder, true)

	TabGap = Tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	Title = lipgloss.NewStyle().
		MarginLeft(1).
		MarginRight(5).
		Padding(0, 1).
		Foreground(lipgloss.Color("#874BFD")).SetString("ANIME TUI")

	Desc = Base.MarginTop(1)

	Info = Base.
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(Subtle)
)
