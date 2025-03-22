package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/mathisdev7/licensy-dashboard-backend/config"
	"github.com/mathisdev7/licensy-dashboard-backend/handlers"
	"github.com/mathisdev7/licensy-dashboard-backend/services"
)

func main() {
	config.LoadConfig()

	botToken := config.GetEnv("BOT_TOKEN", "")
	if botToken == "" {
		log.Fatal("Bot token is required")
	}

	discord, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}
	discord.Identify.Intents = discordgo.IntentsGuilds
	if err := discord.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer discord.Close()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://licensy.xyz, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Get("/api/users/@me", handlers.GetUserHandler(discord))
	app.Get("/api/users/@me/guilds", handlers.GetUserGuildsHandler(discord))
	app.Get("/api/guilds/:guildID/roles", func(c *fiber.Ctx) error {
		guildID := c.Params("guildID")
		accessToken := c.Get("Authorization")
		if accessToken == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		roles, err := services.GetAllRolesInGuild(accessToken, botToken, guildID)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"roles": roles})
	})
	app.Get("/api/guilds/:guildID/roles", func(c *fiber.Ctx) error {
		guildID := c.Params("guildID")
		userID := c.Query("userID")
		roleID := c.Query("roleID")
		accessToken := c.Get("Authorization")

		if userID == "" || roleID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing userID or roleID query parameter"})
		}

		if accessToken == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing Authorization header"})
		}

		isRoleGreaterThan, err := services.IsRoleGreaterThan(accessToken, botToken, guildID, userID, roleID)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"isRoleGreaterThan": !isRoleGreaterThan})
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "Server is running", "code": 200})
	})

	port := config.GetEnv("API_PORT", "4000")
	log.Println("Server running on port " + port)
	app.Listen(":" + port)
}
