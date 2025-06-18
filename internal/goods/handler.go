package goods

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func Routes(router fiber.Router) {
	router.Post("/", create)
	router.Get("/:id", get)
	router.Patch("/:id", patch)
	router.Delete("/:id", delete)
}

func create(c *fiber.Ctx) error {
	var input struct {
		ProjectID   int    `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	if input.ProjectID == 0 || input.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing required fields"})
	}

	good, err := Create(c.Context(), input.ProjectID, input.Name, input.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(good)
}

func get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	good, err := GetWithCache(c.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    3,
				"message": "errors.common.notFound",
				"details": fiber.Map{},
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(good)
}

func patch(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	good, err := Update(c.Context(), id, input.Name, input.Description)
	if err != nil {
		if err == ErrNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    3,
				"message": "errors.common.notFound",
				"details": fiber.Map{},
			})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(good)
}

func delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	err = Delete(c.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    3,
				"message": "errors.common.notFound",
				"details": fiber.Map{},
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
