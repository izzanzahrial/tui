package model

import (
	"fmt"
	"log"
	"strings"

	"github.com/izzanzahrial/tui/message"
	"github.com/izzanzahrial/tui/style"
	"github.com/izzanzahrial/tui/url"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))

type Main struct {
	// General
	width  int
	height int

	// Menubar
	menubar []string
	cursor  int
	focus   bool

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
		focus:   true,
		detail:  d,
		client:  c,
	}
}

func (m Main) Init() tea.Cmd {
	return m.rank.Init()
}

func (m Main) Focused() bool {
	return m.focus
}

func (m *Main) Focus() {
	m.focus = true
}

func (m *Main) Blur() {
	m.focus = false
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// TODO: move title height and menubar height to constants
		titleHeight := lipgloss.Height(style.Title.Render())
		menubarHeight := lipgloss.Height(m.generateMenubar())

		// TODO: do something with the number 20 and 2
		msg.Width = msg.Width - 20
		msg.Height = msg.Height - titleHeight - menubarHeight + 2
		fmt.Println("Main ", msg.Height, msg.Width)
		// Update the child models with the new dimensions *immediately* and collect their commands
		rank, cmd := m.rank.Update(msg)
		newRank, ok := rank.(*Rank)
		if !ok {
			panic("rank is not of type *Rank")
		}
		m.rank = newRank
		cmds = append(cmds, cmd)

		detail, cmd := m.detail.Update(msg)
		newDetail, ok := detail.(*Detail)
		if !ok {
			panic("detail is not of type *Detail")
		}
		m.detail = newDetail
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	// TODO: do something with the error
	case message.ErrMsg:
		log.Println("Error: ", msg.Error())
		return m, tea.Quit

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
		case "down", "j":
			m.Blur()
		case "ctrcl+c", "q":
			return m, tea.Quit
		}

	case message.RankMsg:
		m.cursor = 0
		rank, cmd := m.rank.Update(msg)
		newRank, ok := rank.(*Rank)
		if !ok {
			panic("rank is not of type *Rank")
		}
		m.rank = newRank
		cmds = append(cmds, cmd)

		// return m, tea.Batch(cmds...)

	case message.DetailMsg:
		m.cursor = 1
		detail, cmd := m.detail.Update(msg)
		newDetail, ok := detail.(*Detail)
		if !ok {
			panic("detail is not of type *Detail")
		}
		m.detail = newDetail
		cmds = append(cmds, cmd)

		// return m, tea.Batch(cmds...)
	}

	// Pass messages down to the active child model.
	switch m.menubar[m.cursor] {
	case "Rank":
		m.rank.Focus()
		rank, cmd := m.rank.Update(msg)
		newRank, ok := rank.(*Rank)
		if !ok {
			panic("rank is not of type *Rank")
		}
		m.rank = newRank
		cmds = append(cmds, cmd)
	case "Detail":
		detail, cmd := m.detail.Update(msg)
		newDetail, ok := detail.(*Detail)
		if !ok {
			panic("detail is not of type *Detail")
		}
		m.detail = newDetail
		cmds = append(cmds, cmd)
	}

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

	return menubar
}

func (m Main) View() string {
	// Page that being displayed is depend on the largest width and tallest height of content
	// in this case is the gap, which is from the end menubar (search tab) to the end of the window
	menubar := m.generateMenubar()
	// TODO: change the 20 into something constant
	gap := style.TabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(menubar)-20)))
	menubar = lipgloss.JoinHorizontal(lipgloss.Bottom, menubar, gap)

	var body string
	switch m.menubar[m.cursor] {
	case "Rank":
		body = m.rank.View()
	case "Detail":
		body = m.detail.View()
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		baseStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				style.Title.Render(),
				menubar,
				body,
			),
		),
	)
}
