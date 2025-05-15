package middleware

import (
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"gitlab.smartbet.am/golang/smart-image/internal/service/config"
)

type AuthMiddleware struct {
	cfg    *config.Config
	client *resty.Client
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	client := resty.New()
	return &AuthMiddleware{
		cfg:    cfg,
		client: client,
	}
}

func (a *AuthMiddleware) Handle(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	service := c.Get("X-Service")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header missing",
		})
	}
	var url string
	switch service {
	case "Affiliate":
		url = "http://affiliate-new/api/check-auth-user"
		break
	case "Control":
		url = "http://control-api/api/check-auth-user"
		break
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "X-Service header missing",
		})
	}

	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+token).
		Post(url)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "auth service unreachable",
		})
	}

	if resp.StatusCode() != fiber.StatusOK {
		return c.Status(fiber.StatusUnauthorized).Send(resp.Body())
	}

	return c.Next()
}
