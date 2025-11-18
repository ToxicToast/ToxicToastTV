package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/toxictoast/toxictoastgo/shared/config"
)

type KeycloakAuth struct {
	config    *config.KeycloakConfig
	publicKey *rsa.PublicKey
}

type KeycloakClaims struct {
	jwt.RegisteredClaims
	Email             string                 `json:"email"`
	PreferredUsername string                 `json:"preferred_username"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	RealmAccess       map[string]interface{} `json:"realm_access"`
	ResourceAccess    map[string]interface{} `json:"resource_access"`
}

type UserContext struct {
	UserID   string
	Email    string
	Username string
	Roles    []string
}

func NewKeycloakAuth(cfg *config.KeycloakConfig) (*KeycloakAuth, error) {
	auth := &KeycloakAuth{
		config: cfg,
	}

	// If public key is provided in config, use it
	if cfg.PublicKey != "" {
		publicKey, err := parsePublicKey(cfg.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		auth.publicKey = publicKey
		log.Println("Keycloak auth initialized with provided public key")
	} else {
		// Fetch public key from Keycloak
		publicKey, err := auth.fetchPublicKey()
		if err != nil {
			log.Printf("Warning: Failed to fetch Keycloak public key: %v", err)
			log.Println("Authentication will fail until public key is available")
		} else {
			auth.publicKey = publicKey
			log.Println("Keycloak auth initialized with fetched public key")
		}
	}

	return auth, nil
}

// fetchPublicKey fetches the public key from Keycloak's JWKS endpoint
func (k *KeycloakAuth) fetchPublicKey() (*rsa.PublicKey, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", k.config.URL, k.config.Realm)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf("no keys found in JWKS")
	}

	// Use the first RSA key
	key := jwks.Keys[0]

	return parseRSAPublicKey(key.N, key.E)
}

// ValidateToken validates a JWT token and extracts user information
func (k *KeycloakAuth) ValidateToken(tokenString string) (*UserContext, error) {
	if k.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}

	token, err := jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return k.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	// Extract roles
	roles := extractRoles(claims)

	return &UserContext{
		UserID:   claims.Subject,
		Email:    claims.Email,
		Username: claims.PreferredUsername,
		Roles:    roles,
	}, nil
}

// UnaryInterceptor is a gRPC unary interceptor for JWT authentication
func (k *KeycloakAuth) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for health checks and public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
		userContext, err := k.ValidateToken(tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add user context to the request context
		ctx = context.WithValue(ctx, userContextKey, userContext)

		return handler(ctx, req)
	}
}

// StreamInterceptor is a gRPC stream interceptor for JWT authentication
func (k *KeycloakAuth) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for public endpoints
		if isPublicEndpoint(info.FullMethod) {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
		userContext, err := k.ValidateToken(tokenString)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Create a new context with user information
		ctx := context.WithValue(ss.Context(), userContextKey, userContext)
		wrappedStream := &wrappedServerStream{ServerStream: ss, ctx: ctx}

		return handler(srv, wrappedStream)
	}
}

// Helper functions

type contextKey string

const userContextKey contextKey = "user"

func GetUserContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	if !ok {
		return nil, fmt.Errorf("user context not found")
	}
	return user, nil
}

func extractRoles(claims *KeycloakClaims) []string {
	roles := []string{}

	// Extract realm roles
	if realmAccess, ok := claims.RealmAccess["roles"].([]interface{}); ok {
		for _, role := range realmAccess {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
	}

	// Extract resource roles (client-specific roles)
	for _, resource := range claims.ResourceAccess {
		if resourceMap, ok := resource.(map[string]interface{}); ok {
			if resourceRoles, ok := resourceMap["roles"].([]interface{}); ok {
				for _, role := range resourceRoles {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	return roles
}

func isPublicEndpoint(method string) bool {
	publicEndpoints := []string{
		"/grpc.health.v1.Health/Check",
		"/blog.BlogService/ListPosts",
		"/blog.BlogService/GetPost",
		"/blog.BlogService/ListCategories",
		"/blog.BlogService/GetCategory",
		"/blog.BlogService/ListTags",
		"/blog.BlogService/GetTag",
		"/blog.BlogService/ListComments",
		"/blog.BlogService/GetComment",
	}

	for _, endpoint := range publicEndpoints {
		if method == endpoint {
			return true
		}
	}
	return false
}

func parsePublicKey(keyStr string) (*rsa.PublicKey, error) {
	// Try to parse as PEM format first
	// If that fails, try base64 decode and parse as raw key
	// This is a simplified implementation - you may need to adjust based on your key format
	return nil, fmt.Errorf("public key parsing not implemented - please provide via JWKS endpoint")
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes).Int64()

	return &rsa.PublicKey{
		N: n,
		E: int(e),
	}, nil
}

// wrappedServerStream wraps a grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
