package url

import (
	"fmt"
	"os"
	"strings"

	"github.com/izzanzahrial/tui/entity"
	"github.com/izzanzahrial/tui/message"
)

func (c *Client) AnimeDetail(id int) (*entity.Detail, error) {
	var airingAnimeUrl strings.Builder
	airingAnimeUrl.WriteString(baseURL)
	airingAnimeUrl.WriteString("/{id}")

	fields := []string{
		"id", "title", "main_picture", "alternative_titles",
		"start_date", "end_date", "synopsis", "mean",
		"rank", "popularity", "num_list_users", "num_scoring_users",
		"nsfw", "created_at", "updated_at", "media_type",
		"status", "genres", "num_episodes",
		"start_season", "broadcast", "source", "average_episode_duration",
		"rating", "pictures", "background", "related_anime",
		"related_manga", "recommendations", "studios", "statistics",
	}
	fieldsString := strings.Join(fields, ",")

	// var data map[string]any
	data := &entity.Detail{}
	clientID := os.Getenv("CLIENT_ID")
	request := c.client.R().
		SetHeader(ClientIDHeader, clientID).
		SetPathParam("id", fmt.Sprintf("%d", id)).
		SetQueryParam("fields", fieldsString).
		SetResult(data)

	_, err := request.Get(airingAnimeUrl.String())
	if err != nil {
		return nil, message.ErrMsg{Err: err}
	}

	return data, nil
}
