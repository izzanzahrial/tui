package model

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/izzanzahrial/tui/entity"
	"github.com/izzanzahrial/tui/message"
	"github.com/izzanzahrial/tui/url"
)

type Rank struct {
	anime     *entity.Data
	animeMap  map[int]*entity.AnimeRank
	isLoading bool
	spinner   spinner.Model
	table     *table.Model
	client    *url.Client
}

func NewRank(c *url.Client) *Rank {
	sp := spinner.New()
	// spinner style
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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

	return &Rank{
		anime:     &entity.Data{},
		animeMap:  make(map[int]*entity.AnimeRank),
		isLoading: true,
		spinner:   sp,
		table:     &t,
		client:    c,
	}
}

// initialRequest fetches the first batch of data needed for the rank view
func (r Rank) initialRequest() tea.Msg {
	data, ok := r.client.AnimeRank(0, nil, nil).(*entity.Data)
	if !ok {
		return message.ErrMsg{Err: fmt.Errorf("failed to fetch initial data")}
	}

	return data
}

func (r Rank) Init() tea.Cmd {
	return tea.Batch(r.spinner.Tick, r.initialRequest)
}

func (r *Rank) Focused() bool {
	return r.table.Focused()
}

func (r *Rank) Focus() {
	r.table.Focus()
}

func (r *Rank) Blur() {
	r.table.Blur()
}

func (r *Rank) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.table.SetWidth(msg.Width)
		r.table.SetHeight(msg.Height - 6)

	case *entity.Data:
		r.anime = msg
		for _, v := range r.anime.AnimeRank {
			r.animeMap[v.Rank.Rank] = &v
		}

		var rows []table.Row
		for _, anime := range r.anime.AnimeRank {
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
		r.table.SetRows(rows)
		r.Focus()
		r.isLoading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrcl+c", "q":
			return r, tea.Quit
		case "enter", " ":
			r.Blur()
			rank := r.table.SelectedRow()[0]
			rankInt, err := strconv.Atoi(rank)
			if err != nil {
				return r, func() tea.Msg {
					return message.ErrMsg{Err: fmt.Errorf("error parsing rank [%s]: %w", rank, err)}
				}
			}

			anime, ok := r.animeMap[rankInt]
			if !ok {
				return r, func() tea.Msg {
					return message.ErrMsg{Err: errors.New("anime not found")}
				}
			}

			return r, func() tea.Msg { return message.DetailMsg{ID: anime.Anime.ID} }
		}

	case message.RankMsg:
		r.Focus()
	}

	if r.isLoading {
		var cmd tea.Cmd
		r.spinner, cmd = r.spinner.Update(msg)
		return r, cmd
	}

	t, cmd := r.table.Update(msg)
	r.table = &t
	return r, cmd
}

func (r Rank) View() string {
	if r.isLoading {
		return lipgloss.JoinHorizontal(lipgloss.Center, r.spinner.View(), "Loading...")
	}
	return baseStyle.Align(lipgloss.Left).Render(r.table.View())
}
