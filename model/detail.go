package model

import (
	"bytes"
	"fmt"
	"os"
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
const animeTemplate = `
#{{.AlternativeTitle.EngTitle}} 
> Relased : {{.StartDate}}

## Overview
All Time Rank : {{.Rank}}
Popularity    : {{.Popularity}}
Status        : {{.Status}}
Genres        : {{range .Genres}}{{.Name}} {{end}}
Rating        : {{.Rating}}
Studios       : {{range .Studios}}{{.Name}} {{end}}

================================================================================================================================================================

Synopsis:

{{.Synopsis}} {{if ne .Background ""}}
================================================================================================================================================================

Background:

{{.Background}} {{end}} {{if .RelatedAnimes}}
================================================================================================================================================================

Related Anime:
	{{- range .RelatedAnimes}}

	Title : {{.Node.Title}}
	Relation : {{.RelationType}}
	{{end}} {{end}} {{if .Recomendations}}
================================================================================================================================================================

Recomendations Anime:
	{{- range .Recomendations}}

	Title : {{.Node.Title}}
	{{end}}
{{end}}
`

type Detail struct {
	RawData     map[string]any
	AnimeDetail *entity.Detail
	viewport    viewport.Model
	ready       bool
	client      *url.Client
	templ       *template.Template
}

func NewDetail(c *url.Client) *Detail {
	templ, err := template.New("anime_detail").Parse(animeTemplate)
	if err != nil {
		panic(err)
	}
	return &Detail{
		RawData:     make(map[string]any),
		AnimeDetail: &entity.Detail{},
		viewport:    viewport.New(0, 0),
		ready:       false,
		client:      c,
		templ:       templ,
	}
}

func (d Detail) Init() tea.Cmd {
	return nil
}

func (d *Detail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		titleHeight := lipgloss.Height(style.Title.Render())
		// TODO: change the 1 and 5 into something constant
		menubarHeight := 6 // this is supposed to be the height of the menubar, but since it bounds to the rank model methods, it will be 6
		headerHeight := lipgloss.Height(d.headerView())
		footerHeight := lipgloss.Height(d.footerView())
		verticalMarginHeight := headerHeight + footerHeight + titleHeight + menubarHeight

		if !d.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.

			// TODO: change the 20 into something constant
			d.viewport = viewport.New(msg.Width-20, msg.Height-verticalMarginHeight)
			d.viewport.YPosition = headerHeight
			d.ready = true

			// TODO: create the placeholder for viewport.Content when there is no data
		} else {
			// TODO: change the 20 into something constant
			d.viewport.Width = msg.Width - 20
			d.viewport.Height = msg.Height - verticalMarginHeight

			// TODO: recreate the viewport.Content based on the new window size
		}
	case tea.KeyMsg:
		switch msg.String() {
		// TODO: case up, back to the menubar
		case "esc":
			return d, func() tea.Msg { return message.BackToMenubarMsg{} }
		case "ctrcl+c", "q":
			return d, tea.Quit
		}
	case message.DetailMsg:
		data, ok := d.client.AnimeDetail(msg.ID).(*entity.Detail)
		if !ok {
			return d, func() tea.Msg { return message.ErrMsg{Err: fmt.Errorf("failed to get detail")} }
		}

		d.sanitizeAndSetContent(data)

		// Generate markdown
		err := d.generateMarkdown()
		if err != nil {
			return d, func() tea.Msg { return message.ErrMsg{Err: err} }
		}

		// Read and set content
		content, err := os.ReadFile("anime.md")
		if err != nil {
			return d, func() tea.Msg { return message.ErrMsg{Err: err} }
		}
		d.viewport.SetContent(string(content))

		err = os.Remove("anime.md")
		if err != nil {
			return d, func() tea.Msg { return message.ErrMsg{Err: err} }
		}
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

func (d Detail) View() string {
	if !d.ready {
		// TODO: align to the center
		// Return something to occupy space or an initializing message
		// Ensure it has a defined width if lipgloss.Place or Join is trying to center it.
		return lipgloss.NewStyle().
			Width(d.viewport.Width).
			Render("You need to choose an anime first in rank page or search page")
	}

	return lipgloss.JoinVertical(lipgloss.Center, // Or lipgloss.Top
		d.headerView(),
		d.viewport.View(),
		d.footerView(),
	)
}

// TODO: can do this better, when the cut is within a word
// we can postponed the cut to the next word or add - to the end
func (d *Detail) sanitizeAndSetContent(data *entity.Detail) {
	var newSynopsis strings.Builder
	// TODO: do something with number 4, make it a constant
	// and handle dynamic width
	width := d.viewport.Width - 4
	// split the synopsis into paragraphs
	splitedSynopsis := strings.Split(data.Synopsis, "\n")
	for _, line := range splitedSynopsis {
		// within each paragraph, split into lines that are less than the viewport width
		for len(line) > width {
			newLine := line[:width]
			line = line[width:]
			newSynopsis.WriteString(newLine + "\n")
		}

		// add the rest of the line and new line for the next paragraph
		newSynopsis.WriteString(line + "\n")
	}

	var newBackground strings.Builder
	splitedBackground := strings.Split(data.Background, "\n")
	for _, line := range splitedBackground {
		for len(line) > width {
			newLine := line[:width]
			line = line[width:]
			newBackground.WriteString(newLine + "\n")
		}

		// add the rest of the line and new line for the next paragraph
		newBackground.WriteString(line + "\n")
	}

	d.AnimeDetail = data
	d.AnimeDetail.Synopsis = newSynopsis.String()
	d.AnimeDetail.Background = newBackground.String()
}

func (d Detail) generateMarkdown() error {
	var buff bytes.Buffer
	err := d.templ.Execute(&buff, d.AnimeDetail)
	if err != nil {
		return err
	}

	err = os.WriteFile("anime.md", buff.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (d Detail) headerView() string {
	// TODO: add title of the anime
	// title := style.Title.Render(fmt.Sprintf("%s", d.AnimeDetail.AlternativeTitle.EngTitle))
	title := style.Title.Render()
	line := strings.Repeat("─", max(0, d.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line)
}

func (d Detail) footerView() string {
	info := style.Info.Render(fmt.Sprintf("%3.f%%", d.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, d.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
