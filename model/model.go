package model

import (
	"log"
	"strings"

	"github.com/izzanzahrial/tui/message"
	"github.com/izzanzahrial/tui/style"
	"github.com/izzanzahrial/tui/url"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// How much vertical space the title and menubar occupy.
	mainHeaderHeight = 2
	// Any additional vertical padding for the main content area.
	mainVerticalPadding = 1
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 1)

type Main struct {
	// General
	width        int
	height       int
	contentWidth int
	err          error

	// Menubar
	menubar []string
	cursor  int

	// Rank Page
	rank *Rank

	// Detail Page
	detail *Detail

	client *url.Client
}

func New() Main {
	menubar := []string{"Rank", "Detail", "Search"}

	c := url.NewClient()
	r := NewRank(c)
	d := NewDetail(c)

	return Main{
		rank:    r,
		menubar: menubar,
		cursor:  0,
		detail:  d,
		client:  c,
	}
}

func (m Main) Init() tea.Cmd {
	return m.rank.initialRequest
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If we're in an error state, the only thing we care about is the key press
	// to dismiss the error.
	if m.err != nil {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter", "esc", "q":
				m.err = nil // Clear the error
			}
		}
		return m, nil
	}

	// tea.Cmd this is used if you want to set new value to the current UI
	// e.g you have a textInput.Model that you want to update with a new user inputed value
	var (
		// cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	// Get the current window size
	// You modify of your UI based on the window size
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate the final content width
		// by subtracting the horizontal space taken by the baseStyle's border and padding.
		horizontalMargin := baseStyle.GetHorizontalFrameSize()
		m.contentWidth = m.width - horizontalMargin

		// Calculate the height available for child models.
		borderHeight := baseStyle.GetVerticalFrameSize()
		contentHeight := m.height - mainHeaderHeight - mainVerticalPadding - borderHeight

		// Create the message for child models with the correct dimensions.
		childMsg := tea.WindowSizeMsg{Width: m.contentWidth, Height: contentHeight}

		// Update the child models with the new dimensions *immediately* and collect their commands
		rank, cmd := m.rank.Update(childMsg)
		if r, ok := rank.(*Rank); ok {
			m.rank = r
		}
		cmds = append(cmds, cmd)

		detail, cmd := m.detail.Update(childMsg)
		if d, ok := detail.(*Detail); ok {
			m.detail = d
		}
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case message.ErrMsg:
		m.err = msg.Err                              // Set the error
		log.Printf("An error occurred: %v", msg.Err) // Log the technical details
		return m, nil

	// if key press
	case tea.KeyMsg:
		switch msg.String() {
		case "right", "l":
			if m.cursor < len(m.menubar)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.menubar) - 1
			}
		case "ctrcl+c", "q":
			return m, tea.Quit
		}

	// The main update switch at the bottom will handle passing subsequent messages.
	case message.RankMsg:
		m.cursor = 0
		m.rank.Focus()

	case message.DetailMsg:
		m.cursor = 1
		m.detail.Focus()
	}

	// Delegate messages down to the active child model.
	var cmd tea.Cmd
	switch m.menubar[m.cursor] {
	case "Rank":
		m.rank.Focus()
		m.detail.Blur()
		rank, newCmd := m.rank.Update(msg)
		if r, ok := rank.(*Rank); ok {
			m.rank = r
		}
		cmd = newCmd
	case "Detail":
		m.rank.Blur()
		m.detail.Focus()
		detail, newCmd := m.detail.Update(msg)
		if d, ok := detail.(*Detail); ok {
			m.detail = d
		}
		cmd = newCmd
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Main) generateMenubar() string {
	var menu []string

	for i, v := range m.menubar {
		if i == m.cursor {
			menu = append(menu, style.ActiveTab.Render(v))
		} else {
			menu = append(menu, style.Tab.Render(v))
		}
	}

	menubar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menu...,
	)
	gap := style.TabGap.Render(strings.Repeat(" ", max(0, m.contentWidth-lipgloss.Width(menubar))))
	return lipgloss.JoinHorizontal(lipgloss.Bottom, menubar, gap)
}

func (m Main) View() string {
	if m.err != nil {
		return m.errorView(m.width, m.height)
	}

	// The title should be rendered inside a container that can be constrained.
	// We'll set the width of the title's container to the contentWidth.
	title := lipgloss.NewStyle().Width(m.contentWidth).Render(style.Title.Render())

	var body string
	switch m.menubar[m.cursor] {
	case "Rank":
		body = m.rank.View()
	case "Detail":
		body = m.detail.View()
	}

	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.generateMenubar(),
		body,
	)

	return baseStyle.Render(mainContent)
}

func (m Main) errorView(width, height int) string {
	errorHeader := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F1F1")).
		Background(lipgloss.Color("#FF5F87")).
		Bold(true).
		Padding(0, 1).
		Render(" Oh No! An Error Occurred ")

	errorBody := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(1).
		Render(m.err.Error())

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press Enter or Esc to continue...")

	errorBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF5F87")).
		Render(lipgloss.JoinVertical(lipgloss.Center, errorHeader, errorBody, helpText))

	// Place the error box in the center of the screen.
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, errorBox)
}
