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
	sp := spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))))

	columns := []table.Column{
		{Title: "Rank", Width: 4},
		{Title: "Title", Width: 40},
		{Title: "Japanese Title", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true), // Start focused by default
		table.WithHeight(10),    // Initial height, will be resized
	)

	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
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
	data, err := r.client.AnimeRank(0, nil, nil)
	if err != nil {
		return message.ErrMsg{Err: fmt.Errorf("failed to fetch anime ranks: %w", err)}
	}

	return data
}

func (r Rank) Init() tea.Cmd {
	return r.spinner.Tick
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
		r.table.SetHeight(msg.Height)
		return r, nil

	case *entity.Data:
		r.anime = msg
		for _, v := range r.anime.AnimeRank {
			r.animeMap[v.Rank.Rank] = &v
		}

		rows := make([]table.Row, len(r.anime.AnimeRank))
		for i, anime := range r.anime.AnimeRank {
			title := anime.Anime.AlternativeTitle.EngTitle
			if title == "" {
				title = anime.Anime.Title
			}
			rows[i] = table.Row{strconv.Itoa(anime.Rank.Rank), title, anime.Anime.AlternativeTitle.JpnTitle}
		}
		r.table.SetRows(rows)
		r.isLoading = false
		return r, nil // No further command needed

	case tea.KeyMsg:
		// Don't handle keys if we're not focused.
		if !r.table.Focused() {
			return r, nil
		}

		switch msg.String() {
		case "enter", " ":
			if len(r.table.SelectedRow()) == 0 {
				return r, nil
			}

			rankStr := r.table.SelectedRow()[0]
			rankInt, err := strconv.Atoi(rankStr)
			if err != nil {
				return r, func() tea.Msg {
					return message.ErrMsg{Err: fmt.Errorf("error parsing rank [%s]: %w", rankStr, err)}
				}
			}

			anime, ok := r.animeMap[rankInt]
			if !ok {
				return r, func() tea.Msg {
					return message.ErrMsg{Err: errors.New("anime not found")}
				}
			}

			// Send message to switch to the detail view
			return r, func() tea.Msg { return message.DetailMsg{ID: anime.Anime.ID} }
		}
	}

	if r.isLoading {
		r.spinner, cmd = r.spinner.Update(msg)
		return r, cmd
	}

	*r.table, cmd = r.table.Update(msg)
	return r, cmd
}

func (r Rank) View() string {
	if r.isLoading {
		loadingStyle := lipgloss.NewStyle().Width(r.table.Width()).Height(r.table.Height()).Align(lipgloss.Center, lipgloss.Center)
		return loadingStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, r.spinner.View(), " Loading..."))
	}
	return baseStyle.Align(lipgloss.Left).Render(r.table.View())
}
