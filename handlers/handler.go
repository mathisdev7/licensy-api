package handlers

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gofiber/fiber/v2"
	"github.com/mathisdev7/licensy-dashboard-backend/services"
)

func GetUserHandler(discord *discordgo.Session) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Authorization format"})
		}
		accessToken := parts[1]

		err := services.CheckOauth2Token(accessToken)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user"})
		}
		return c.JSON(fiber.Map{"authorized": true})
	}
}

func GetUserGuildsHandler(discord *discordgo.Session) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid Authorization format"})
		}
		accessToken := parts[1]

		userGuilds, err := services.GetUserGuilds(accessToken)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch user guilds"})
		}

		botGuilds := make(map[string]struct{})
		for _, guild := range discord.State.Guilds {
			botGuilds[guild.ID] = struct{}{}
		}

		commonGuilds := services.GetCommonGuilds(userGuilds, botGuilds)
		return c.JSON(commonGuilds)
	}
}
