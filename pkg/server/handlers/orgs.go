package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ResolveOrgQuery reads the ?org= query parameter and resolves it to an org UUID.
//
// An absent or empty parameter returns an empty UUID, which every read path treats
// as "no org filter" — the response then spans all orgs, exactly as it did before
// orgs existed. An unknown org is an error rather than a silent empty result, so a
// typo does not look like "this org has no data".
//
// The second return value is a ready-to-return 400 response, so list handlers stay
// a two-line call. It is 400 rather than the 404 orgError would produce, because
// here the org is a malformed filter argument, not the resource being addressed.
func ResolveOrgQuery(ctx context.Context, c *fiber.Ctx) (string, error) {
	ref := strings.TrimSpace(c.Query("org"))
	if ref == "" {
		return "", nil
	}
	orgUUID, err := database.ResolveOrgUUID(ctx, ref)
	if err != nil {
		return "", c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
	return orgUUID, nil
}

// orgStatusFor maps an org error to an HTTP status.
func orgStatusFor(err error) int {
	switch {
	case errors.Is(err, database.ErrOrgNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, database.ErrDefaultOrgImmutable):
		return fiber.StatusForbidden
	default:
		return fiber.StatusInternalServerError
	}
}

func orgError(c *fiber.Ctx, err error) error {
	return c.Status(orgStatusFor(err)).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}

// ListOrgs handles listing all orgs
// @Summary List orgs
// @Description Get every org with its workspace, asset and vulnerability counts
// @Tags Orgs
// @Produce json
// @Success 200 {object} map[string]interface{} "List of orgs"
// @Failure 500 {object} map[string]interface{} "Failed to fetch orgs"
// @Security BearerAuth
// @Router /osm/api/orgs [get]
func ListOrgs(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		orgs, err := database.ListOrgs(ctx)
		if err != nil {
			return orgError(c, err)
		}

		data := make([]fiber.Map, 0, len(orgs))
		for i := range orgs {
			org := &orgs[i]
			stats, err := database.GetOrgStats(ctx, org.UUID)
			if err != nil {
				return orgError(c, err)
			}
			data = append(data, fiber.Map{
				"uuid":        org.UUID,
				"name":        org.Name,
				"description": org.Description,
				"tags":        org.Tags,
				"is_default":  org.IsDefault(),
				"created_at":  org.CreatedAt,
				"updated_at":  org.UpdatedAt,
				"stats":       stats,
			})
		}

		return c.JSON(fiber.Map{"data": data, "total": len(data)})
	}
}

// GetOrg handles fetching a single org
// @Summary Get org
// @Description Get one org by name or UUID
// @Tags Orgs
// @Produce json
// @Param uuid path string true "Org name or UUID"
// @Success 200 {object} map[string]interface{} "Org details"
// @Failure 404 {object} map[string]interface{} "Org not found"
// @Security BearerAuth
// @Router /osm/api/orgs/{uuid} [get]
func GetOrg(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		org, err := database.ResolveOrgRef(ctx, c.Params("uuid"))
		if err != nil {
			return orgError(c, err)
		}
		return c.JSON(fiber.Map{"data": org})
	}
}

// GetOrgStats handles fetching an org's aggregate counts
// @Summary Get org stats
// @Description Get workspace, asset, vulnerability and run counts for an org
// @Tags Orgs
// @Produce json
// @Param uuid path string true "Org name or UUID"
// @Success 200 {object} map[string]interface{} "Org statistics"
// @Failure 404 {object} map[string]interface{} "Org not found"
// @Security BearerAuth
// @Router /osm/api/orgs/{uuid}/stats [get]
func GetOrgStats(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		org, err := database.ResolveOrgRef(ctx, c.Params("uuid"))
		if err != nil {
			return orgError(c, err)
		}

		stats, err := database.GetOrgStats(ctx, org.UUID)
		if err != nil {
			return orgError(c, err)
		}
		return c.JSON(fiber.Map{"data": stats})
	}
}

// createOrgRequest is the body for POST /osm/api/orgs
type createOrgRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	UUID        string   `json:"uuid"`
	Tags        []string `json:"tags"`
}

// CreateOrg handles creating an org
// @Summary Create org
// @Description Create a new org
// @Tags Orgs
// @Accept json
// @Produce json
// @Param body body createOrgRequest true "Org to create"
// @Success 201 {object} map[string]interface{} "Created org"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Security BearerAuth
// @Router /osm/api/orgs [post]
func CreateOrg(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req createOrgRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "invalid request body: " + err.Error(),
			})
		}
		if strings.TrimSpace(req.Name) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "name is required",
			})
		}

		ctx := context.Background()
		org, err := database.CreateOrg(ctx, req.Name, req.Description, req.UUID, req.Tags)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": org})
	}
}

// updateOrgRequest is the body for PUT /osm/api/orgs/:uuid
type updateOrgRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	// Workspaces, when non-empty, assigns those workspaces to this org and
	// cascades the stamp to their assets, vulnerabilities and runs.
	Workspaces []string `json:"workspaces"`
}

// UpdateOrg handles updating an org and optionally assigning workspaces to it
// @Summary Update org
// @Description Update an org's metadata, and optionally assign workspaces to it
// @Tags Orgs
// @Accept json
// @Produce json
// @Param uuid path string true "Org name or UUID"
// @Param body body updateOrgRequest true "Fields to update"
// @Success 200 {object} map[string]interface{} "Updated org"
// @Failure 403 {object} map[string]interface{} "Default org cannot be renamed"
// @Failure 404 {object} map[string]interface{} "Org not found"
// @Security BearerAuth
// @Router /osm/api/orgs/{uuid} [put]
func UpdateOrg(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req updateOrgRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "invalid request body: " + err.Error(),
			})
		}

		ctx := context.Background()

		org, err := database.ResolveOrgRef(ctx, c.Params("uuid"))
		if err != nil {
			return orgError(c, err)
		}

		updated, err := database.UpdateOrg(ctx, org.UUID, req.Name, req.Description, req.Tags)
		if err != nil {
			return orgError(c, err)
		}

		response := fiber.Map{"data": updated}

		if len(req.Workspaces) > 0 {
			counts, err := database.AssignWorkspacesToOrg(ctx, org.UUID, req.Workspaces)
			if err != nil {
				return orgError(c, err)
			}
			response["assigned"] = counts
		}

		return c.JSON(response)
	}
}

// DeleteOrg handles deleting an org
// @Summary Delete org
// @Description Delete an org. Its data moves to the default org unless purge=true.
// @Tags Orgs
// @Produce json
// @Param uuid path string true "Org name or UUID"
// @Param purge query bool false "Also delete the org's workspaces, assets, vulnerabilities and runs" default(false)
// @Success 200 {object} map[string]interface{} "Deletion result"
// @Failure 403 {object} map[string]interface{} "Default org cannot be deleted"
// @Failure 404 {object} map[string]interface{} "Org not found"
// @Security BearerAuth
// @Router /osm/api/orgs/{uuid} [delete]
func DeleteOrg(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		org, err := database.ResolveOrgRef(ctx, c.Params("uuid"))
		if err != nil {
			return orgError(c, err)
		}

		purge := c.QueryBool("purge", false)
		if err := database.DeleteOrg(ctx, org.UUID, purge); err != nil {
			return orgError(c, err)
		}

		return c.JSON(fiber.Map{
			"message": "org deleted",
			"uuid":    org.UUID,
			"name":    org.Name,
			"purged":  purge,
		})
	}
}
