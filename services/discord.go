package services

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mathisdev7/licensy-dashboard-backend/utils"
)

type Guild struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Icon   *string `json:"icon"`
	Banner *string `json:"banner"`
	Owner  bool    `json:"owner"`
	Permissions string `json:"permissions"`
}

func GetUserGuilds(accessToken string) ([]Guild, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer " + accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fiber.NewError(resp.StatusCode, "Failed to fetch user guilds")
	}

	var rawGuilds []struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Icon  *string `json:"icon"`
		Owner bool    `json:"owner"`
		Permissions string `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawGuilds); err != nil {
		return nil, err
	}

	var userGuilds []Guild
	for _, g := range rawGuilds {
		userGuilds = append(userGuilds, Guild{
			ID:     g.ID,
			Name:   g.Name,
			Icon:   utils.FormatIconURL(g.ID, g.Icon),
			Banner: nil,
			Owner:  g.Owner,
			Permissions: g.Permissions,
		})
	}
	return userGuilds, nil
}

func GetCommonGuilds(userGuilds []Guild, botGuilds map[string]struct{}) []Guild {
	var commonGuilds []Guild
	for _, guild := range userGuilds {
		if _, exists := botGuilds[guild.ID]; exists {
			permissions, err := strconv.ParseUint(guild.Permissions, 10, 64)
			if err != nil {
				continue
			}
			if permissions&0x0000000000002000 != 0 {
				commonGuilds = append(commonGuilds, guild)
			}
		}
	}
	return commonGuilds
}
