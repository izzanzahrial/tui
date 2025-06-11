package model

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/izzanzahrial/tui/entity"
	"github.com/izzanzahrial/tui/message"
	"github.com/izzanzahrial/tui/style"
	"github.com/izzanzahrial/tui/url"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))

type Rank struct {
	// General
	width  int
	height int

	// Menubar
	menubar []string
	cursor  int
	focus   bool

	// Rank Page
	Anime     *entity.Data
	AnimeMap  map[int]*entity.AnimeRank
	rankTable *table.Model

	// Detail Page
	detail *Detail

	client *url.Client
}

func NewRank() Rank {
	columns := []table.Column{
		{Title: "Rank", Width: 4},
		{Title: "Title", Width: 40},
		{Title: "Japanese Title", Width: 40},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	menubar := []string{"Rank", "Detail", "Search"}

	c := url.NewClient()

	d := NewDetail(c)

	return Rank{
		Anime:     &entity.Data{},
		AnimeMap:  make(map[int]*entity.AnimeRank),
		rankTable: &t,
		menubar:   menubar,
		cursor:    0,
		focus:     true,
		detail:    d,
		client:    c,
	}
}

// Initial request wrapper for AnimeRank function
func (r Rank) initialRequest() tea.Msg {
	data, ok := r.client.AnimeRank(0, nil, nil).(*entity.Data)
	if ok {
		*r.Anime = *data
		for _, v := range r.Anime.AnimeRank {
			r.AnimeMap[v.Rank.Rank] = &v
		}
	}

	return data
}

func (r Rank) Init() tea.Cmd {
	return r.initialRequest
}

func (r Rank) Focused() bool {
	return r.focus
}

func (r *Rank) Focus() {
	r.focus = true
}

func (r *Rank) Blur() {
	r.focus = false
}

func (r Rank) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// tea.Cmd this is used if you want to set new value to the current UI
	// e.g you have a textInput.Model that you want to update with a new user inputed value
	var (
		cmd   tea.Cmd
		cmds  []tea.Cmd
		table table.Model
	)

	switch msg := msg.(type) {
	// to make it more accurate to the current window size
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
		// TODO: move title height and menubar height to constants
		r.rankTable.SetHeight(r.height - lipgloss.Height(style.Title.Render()) - lipgloss.Height(r.generateMenubar()) - 5)
	case message.ErrMsg:
		log.Println("Error: ", msg.Error())
		return r, tea.Quit
	// if key press
	case tea.KeyMsg:
		// if focused is on menubar
		if r.focus {
			// what key was pressed
			switch msg.String() {
			case "right", "l":
				if r.cursor < len(r.menubar)-1 {
					r.cursor++
				} else {
					r.cursor = 0
				}
			case "left", "h":
				if r.cursor > 0 {
					r.cursor--
				} else {
					r.cursor = len(r.menubar) - 1
				}
			case "down", "j":
				r.focus = false
				switch r.menubar[r.cursor] {
				case "Rank":
					r.rankTable.Focus()
				}

			case "ctrcl+c", "q":
				return r, tea.Quit

			}
		} else {
			// for now only for focused on rank table
			// what key was pressed
			switch msg.String() {

			// move focus back to menubar
			case "esc":
				if r.rankTable.Focused() {
					r.rankTable.Blur()
					r.Focus()
				}

			case "ctrcl+c", "q":
				return r, tea.Quit

			// move menubar cursor to the detail page
			case "enter", " ":
				r.cursor = 1
				r.Focus()
				r.rankTable.Blur()
				rankInt, err := strconv.Atoi(r.rankTable.SelectedRow()[0])
				if err != nil {
					return r, func() tea.Msg {
						return message.ErrMsg{
							Err: err,
						}
					}
				}

				anime, ok := r.AnimeMap[rankInt]
				if !ok {
					return r, func() tea.Msg {
						return message.ErrMsg{
							Err: errors.New("anime not found"),
						}
					}
				}

				return r, func() tea.Msg { return message.DetailMsg{ID: anime.Anime.ID} }

			case "up", "k":
				r.rankTable.MoveUp(0)
				// if we go up less than the start of the table
				// meaning we go back to the menubar
				if r.rankTable.SelectedRow()[0] == "1" {
					// return the focus to the menu and blur the table
					r.Focus()
					r.rankTable.Blur()
				}
			}
		}

	//pass the message into the detail model
	case message.DetailMsg:
		detail, cmd := r.detail.Update(msg)
		newDetail, ok := detail.(*Detail)
		if !ok {
			panic("detail is not of type *Detail")
		}
		r.detail = newDetail
		cmds = append(cmds, cmd)

		return r, tea.Batch(cmds...)
	case message.BackToMenubarMsg:
		r.Focus()
		r.rankTable.Blur()
		return r, nil
	}

	// // --- FIX PART 2: Proactively set the child's size ---
	// menubarHeight := lipgloss.Height(r.generateMenubar())
	// bodyWidth := r.width - 2
	// bodyHeight := r.height - menubarHeight - 2
	// msg = tea.WindowSizeMsg{Width: bodyWidth, Height: bodyHeight}

	table, cmd = r.rankTable.Update(msg)
	cmds = append(cmds, cmd)
	r.rankTable = &table

	detail, cmd := r.detail.Update(msg)
	newDetail, ok := detail.(*Detail)
	if !ok {
		panic("detail is not of type *Detail")
	}
	r.detail = newDetail
	cmds = append(cmds, cmd)
	return r, tea.Batch(cmds...)
}

func (r Rank) generateMenubar() string {
	var menu []string

	for i, v := range r.menubar {
		if i == r.cursor {
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

func (r Rank) View() string {
	var body string

	// Menubar
	// Page that being displayed is depend on the largest width and tallest height of content
	// in this case is the gap, which is from the end menubar (search tab) to the end of the window
	menubar := r.generateMenubar()
	// TODO: change the 20 into something constant
	gap := style.TabGap.Render(strings.Repeat(" ", max(0, r.width-lipgloss.Width(menubar)-20)))
	menubar = lipgloss.JoinHorizontal(lipgloss.Bottom, menubar, gap)

	switch r.menubar[r.cursor] {
	case "Rank":
		var rows []table.Row
		for _, anime := range r.Anime.AnimeRank {
			var title string
			if anime.Anime.AlternativeTitle.EngTitle == "" {
				title = anime.Anime.Title
			} else {
				title = anime.Anime.AlternativeTitle.EngTitle
			}

			rows = append(rows, table.Row{
				strconv.Itoa(anime.Rank.Rank),
				title,
				anime.Anime.AlternativeTitle.JpnTitle,
			})
		}
		r.rankTable.SetRows(rows)
		body = baseStyle.Align(lipgloss.Center).Render(r.rankTable.View())
	case "Detail":
		body = r.detail.View()
	}

	// menubar = lipgloss.JoinVertical(lipgloss.Top, style.Title.Render(), menubar)

	return lipgloss.Place(
		r.width,
		r.height,
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
