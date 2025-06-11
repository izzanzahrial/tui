package url

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/izzanzahrial/tui/entity"
	"github.com/izzanzahrial/tui/message"
)

const (
	defaultLimit    = "100"
	defaultRankType = "airing"
)

type RankingType int

const (
	all RankingType = iota + 1
	airing
	upcoming
	tv
	ova
	movie
	special
	bypopularity
	favorite
)

var ranks = map[RankingType]string{
	all:          "all",
	airing:       "airing",
	upcoming:     "upcoming",
	tv:           "tv",
	ova:          "ova",
	movie:        "movie",
	special:      "special",
	bypopularity: "bypopularity",
	favorite:     "favorite",
}

func (c *Client) AnimeRank(typeRank RankingType, limit, offset *int) tea.Msg {
	var airingAnimeUrl strings.Builder
	airingAnimeUrl.WriteString(baseURL)
	airingAnimeUrl.WriteString("/ranking")

	data := &entity.Data{}
	clientID := os.Getenv("CLIENT_ID")
	request := c.client.R().
		SetHeader(ClientIDHeader, clientID).
		SetResult(data)

	if limit != nil && *limit > 0 {
		request.SetQueryParam("limit", fmt.Sprintf("%d", *limit))
	} else {
		request.SetQueryParam("limit", defaultLimit)
	}

	if offset != nil && *offset > 0 {
		request.SetQueryParam("offset", fmt.Sprintf("%d", *offset))
	}

	rankingType, ok := ranks[typeRank]
	if !ok {
		rankingType = defaultRankType
	}
	request.SetQueryParam("ranking_type", rankingType)

	_, err := request.Get(airingAnimeUrl.String())
	if err != nil {
		return message.ErrMsg{Err: err}
	}

	return data
}
