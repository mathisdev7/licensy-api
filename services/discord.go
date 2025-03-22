package services

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mathisdev7/licensy-dashboard-backend/utils"
)

type Guild struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Icon        *string `json:"icon"`
	Banner      *string `json:"banner"`
	Owner       bool    `json:"owner"`
	Permissions string  `json:"permissions"`
}

type RoleColors struct {
	PrimaryColor   int    `json:"primary_color"`
	SecondaryColor *int   `json:"secondary_color"`
	TertiaryColor  *int   `json:"tertiary_color"`
}

type Role struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description *string     `json:"description"`
	Permissions string      `json:"permissions"`
	Position    int         `json:"position"`
	Color       int         `json:"color"`
	Colors      RoleColors  `json:"colors"`
	Hoist       bool        `json:"hoist"`
	Managed     bool        `json:"managed"`
	Mentionable bool        `json:"mentionable"`
	Icon        *string     `json:"icon"`
	UniEmoji    *string     `json:"unicode_emoji"`
	Flags       int         `json:"flags"`
}

func CheckOauth2Token(accessToken string) error {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fiber.NewError(resp.StatusCode, "Failed to verify OAuth2 token")
	}

	return nil
}

func IsRoleGreaterThan(userToken, accessToken, guildID, userID, roleID string) (bool, error) {
	userRoles, err := GetUserRolesInGuild(userToken, accessToken, guildID, userID)
	if err != nil {
		return false, err
	}

	roleToCompare, err := GetRoleInGuild(accessToken, guildID, roleID)
	if err != nil {
		return false, err
	}

	var highestRolePosition int
	for _, userRole := range userRoles {
		if userRole.Position > highestRolePosition {
			highestRolePosition = userRole.Position
		}
	}

	if roleToCompare.Position > highestRolePosition {
		return true, nil
	}
	return false, nil
}

func GetAllRolesInGuild(userToken, accessToken, guildID string) ([]Role, error) {
	tokenErr := CheckOauth2Token(userToken)
	if tokenErr != nil {
		return nil, tokenErr
	}
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/guilds/" + guildID + "/roles", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fiber.NewError(resp.StatusCode, "Failed to fetch roles in guild")
	}

	var roles []Role
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}


func GetUserRolesInGuild(userToken, accessToken, guildID, userID string) ([]Role, error) {
	tokenErr := CheckOauth2Token(userToken)
	if tokenErr != nil {
		return nil, tokenErr
	}
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/guilds/" + guildID + "/members/" + userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bot " + accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fiber.NewError(resp.StatusCode, "Failed to fetch user roles")
	}

	var member struct {
		Roles []string `json:"roles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return nil, err
	}

	var roles []Role
	for _, roleID := range member.Roles {
		role, err := GetRoleInGuild(accessToken, guildID, roleID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func GetRoleInGuild(accessToken, guildID, roleID string) (Role, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/guilds/"+guildID+"/roles", nil)
	if err != nil {
		return Role{}, err
	}
	req.Header.Set("Authorization", "Bot "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Role{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Role{}, fiber.NewError(resp.StatusCode, "Failed to fetch roles in guild")
	}

	var roles []Role
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return Role{}, err
	}

	for _, role := range roles {
		if role.ID == roleID {
			return role, nil
		}
	}

	return Role{}, fiber.NewError(404, "Role not found")
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
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Icon        *string `json:"icon"`
		Owner       bool    `json:"owner"`
		Permissions string  `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawGuilds); err != nil {
		return nil, err
	}

	var userGuilds []Guild
	for _, g := range rawGuilds {
		userGuilds = append(userGuilds, Guild{
			ID:          g.ID,
			Name:        g.Name,
			Icon:        utils.FormatIconURL(g.ID, g.Icon),
			Banner:      nil,
			Owner:       g.Owner,
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
