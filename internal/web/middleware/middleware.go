package middleware

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
	"io"
	"net/http"
)

type AuthMiddleware struct {
	cfg *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func (a *AuthMiddleware) Handle(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header missing",
		})
	}

	req, err := http.NewRequest("POST", "http://control-api/api/check-auth-user", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "auth request failed"})
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "auth service unreachable"})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read auth response"})
	}

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusUnauthorized).Send(body)
	}

	return c.Next()
}
