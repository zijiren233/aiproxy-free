package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy-free/config"
	"github.com/labring/aiproxy-free/db"
	"github.com/labring/aiproxy-free/server/module"
	log "github.com/sirupsen/logrus"
)

const (
	NamespaceKey = "namespace"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(
				http.StatusUnauthorized,
				module.NewAuthenticationError("Authorization header required"),
			)
			c.Abort()

			return
		}

		apiKey := extractAPIKey(authHeader)
		if apiKey == "" {
			c.JSON(
				http.StatusUnauthorized,
				module.NewAuthenticationError("Invalid authorization format"),
			)
			c.Abort()

			return
		}

		namespace, err := getOrCreateNamespace(c.Request.Context(), apiKey)
		if err != nil {
			log.Errorf("Failed to get/create namespace for key %s: %v", apiKey, err)
			c.JSON(http.StatusUnauthorized, module.NewAuthenticationError("Invalid API key"))
			c.Abort()
			return
		}

		c.Set(NamespaceKey, namespace)
		c.Next()
	}
}

func extractAPIKey(authHeader string) string {
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return authHeader
}

func getOrCreateNamespace(ctx context.Context, key string) (string, error) {
	namespace, err := db.GetNamespace(key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ns, authorized := checkKeyAndGetNamespace(ctx, key)
			if !authorized {
				return "", errors.New("key not authorized")
			}

			if ns == "" {
				return "", errors.New("upstream implementation is incorrect: missing Group header")
			}

			err = db.SaveMapping(key, ns)
			if err != nil {
				return "", fmt.Errorf("failed to save mapping: %w", err)
			}

			return ns, nil
		}

		return "", err
	}

	return namespace, nil
}

func checkKeyAndGetNamespace(ctx context.Context, key string) (string, bool) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		config.UpstreamBaseURL+"/v1/models",
		nil,
	)
	if err != nil {
		log.Errorf("Failed to create permission check request: %v", err)
		return "", false
	}

	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to check API key permission: %v", err)
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	namespace := resp.Header.Get("Group")

	return namespace, true
}
