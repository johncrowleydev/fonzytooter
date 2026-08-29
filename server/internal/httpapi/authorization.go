package httpapi

import (
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/johncrowleydev/helix-academy/server/internal/auth"
)

const (
	accessPolicyMetadataKey = "helix-academy.access-policy"
	accessPublic            = "public"
	accessAuthenticated     = "authenticated"
)

func configureAuthorization(api huma.API) {
	if api.OpenAPI().Components.SecuritySchemes == nil {
		api.OpenAPI().Components.SecuritySchemes = make(map[string]*huma.SecurityScheme)
	}
	api.OpenAPI().Components.SecuritySchemes[sessionSecurityScheme] = &huma.SecurityScheme{
		Type:        "apiKey",
		In:          "cookie",
		Name:        auth.DefaultCookieName,
		Description: "Opaque Helix Academy authentication session cookie.",
	}
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		policy, ok := ctx.Operation().Metadata[accessPolicyMetadataKey].(string)
		if !ok {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "API access policy is not configured")
			return
		}
		switch policy {
		case accessPublic:
			next(ctx)
		case accessAuthenticated:
			if _, err := auth.RequireUser(ctx.Context()); err != nil {
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "authentication required")
				return
			}
			next(ctx)
		default:
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError, "API access policy is invalid")
		}
	})
}

func publicOperation(operation huma.Operation) huma.Operation {
	operation.Metadata = withAccessPolicy(operation.Metadata, accessPublic)
	return operation
}

func authenticatedOperation(operation huma.Operation) huma.Operation {
	operation.Metadata = withAccessPolicy(operation.Metadata, accessAuthenticated)
	operation.Security = []map[string][]string{{sessionSecurityScheme: {}}}
	if !slices.Contains(operation.Errors, http.StatusUnauthorized) {
		operation.Errors = append(operation.Errors, http.StatusUnauthorized)
	}
	return operation
}

func withAccessPolicy(metadata map[string]any, policy string) map[string]any {
	if metadata == nil {
		metadata = make(map[string]any, 1)
	}
	metadata[accessPolicyMetadataKey] = policy
	return metadata
}
