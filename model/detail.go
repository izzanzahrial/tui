package model

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/izzanzahrial/tui/entity"
	"github.com/izzanzahrial/tui/message"
	"github.com/izzanzahrial/tui/style"
	"github.com/izzanzahrial/tui/url"
)

// TODO: handle when certain data is zero
// e.g. data background of 'To Be Hero X' ID '53447'
// even though already using omitzero within the entity.Detail json tag
// within the template still show empty string
const animeTemplate = `{{.Title}}
> Released: {{.StartDate}} | Status: {{.Status}}

## Overview
Rank: {{.Rank}} | Popularity: {{.Popularity}} | Rating: {{.Rating}}
Genres: {{.Genres}}
Studios: {{.Studios}}
{{.Separator}}
{{if .Synopsis}}
## Synopsis

{{.Synopsis}}
{{end}}
{{if .Background}}
{{.Separator}}
## Background

{{.Background}}
{{end}}
{{if .RelatedAnimes}}
{{.Separator}}
## Related Anime
{{range .RelatedAnimes}}
- {{.Node.Title}} ({{.RelationType}})
{{end}}
{{end}}
{{if .Recomendations}}
{{.Separator}}
## Recommendations
{{range .Recomendations}}
- {{.Node.Title}}
{{end}}
{{end}}`

type templateData struct {
	Title          string
	StartDate      string
	Status         string
	Rank           int
	Popularity     int
	Rating         string
	Genres         string
	Studios        string
	Synopsis       string
	Background     string
	RelatedAnimes  []entity.RelatedAnime
	Recomendations []entity.Recommendation
	Separator      string
}

type Detail struct {
	viewport  viewport.Model
	client    *url.Client
	templ     *template.Template
	ready     bool
	isFocused bool
}

func NewDetail(c *url.Client) *Detail {
	templ, err := template.New("anime_detail").Parse(animeTemplate)
	if err != nil {
		panic(err)
	}

	return &Detail{
		viewport: viewport.New(0, 0),
		client:   c,
		templ:    templ,
		ready:    false,
	}
}

func (d Detail) Init() tea.Cmd { return nil }

func (d *Detail) Focus() { d.isFocused = true }
func (d *Detail) Blur()  { d.isFocused = false }

func (d *Detail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(d.headerView())
		footerHeight := lipgloss.Height(d.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !d.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.

			d.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			d.viewport.YPosition = headerHeight
			d.ready = true
		} else {
			d.viewport.Width = msg.Width
			d.viewport.Height = msg.Height - verticalMarginHeight
		}

	case tea.KeyMsg:
		if !d.isFocused {
			return d, nil
		}

	case message.DetailMsg:
		d.viewport.GotoTop()
		detail, err := d.client.AnimeDetail(msg.ID)
		if err != nil {
			return d, func() tea.Msg {
				return message.ErrMsg{Err: fmt.Errorf("failed to get detail for ID %d: %w", msg.ID, err)}
			}
		}

		content, err := d.renderContent(detail)
		if err != nil {
			return d, func() tea.Msg { return message.ErrMsg{Err: err} }
		}
		d.viewport.SetContent(content)
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// It handles text wrapping and templating in one step, entirely in memory.
func (d *Detail) renderContent(data *entity.Detail) (string, error) {
	// This style will handle word wrapping for us automatically.
	// We subtract a little padding to make it look nice.
	contentWidth := d.viewport.Width - 2
	contentStyle := lipgloss.NewStyle().Width(contentWidth)

	title := data.AlternativeTitle.EngTitle
	if title == "" {
		title = data.Title
	}

	// Prepare data for the template
	templatePayload := templateData{
		Title:          style.DetailTitle.Render(title),
		StartDate:      data.StartDate,
		Status:         strings.ReplaceAll(data.Status, "_", " "),
		Rank:           data.Rank,
		Popularity:     data.Popularity,
		Rating:         data.Rating,
		Genres:         joinNames(data.Genres),
		Studios:        joinNames(data.Studios),
		Synopsis:       contentStyle.Render(data.Synopsis),
		Background:     contentStyle.Render(data.Background),
		RelatedAnimes:  data.RelatedAnimes,
		Recomendations: data.Recomendations,
		Separator:      style.Separator.Render(strings.Repeat("─", d.viewport.Width)),
	}

	var buf bytes.Buffer
	if err := d.templ.Execute(&buf, templatePayload); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (d Detail) View() string {
	if !d.ready {
		return lipgloss.NewStyle().
			Width(d.viewport.Width).
			Height(d.viewport.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Select an anime from the Rank page.")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		d.headerView(),
		d.viewport.View(),
		d.footerView(),
	)
}

func (d Detail) headerView() string {
	line := strings.Repeat("─", d.viewport.Width)
	return lipgloss.JoinHorizontal(lipgloss.Center, line)
}

func (d Detail) footerView() string {
	info := style.Info.Render(fmt.Sprintf("%3.f%%", d.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, d.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// Helper function to join struct slices with a .Name field into a string.
func joinNames[T any](items []T) string {
	var names []string
	for _, item := range items {
		// Use reflection to get the 'Name' field. This is a bit advanced but very flexible.
		// A simpler approach would be to use type switches or interfaces if you only have a few types.
		if name, ok := any(item).(interface{ GetName() string }); ok {
			names = append(names, name.GetName())
		}
	}
	return strings.Join(names, ", ")
}
