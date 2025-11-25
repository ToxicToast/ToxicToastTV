# JWT Authentication Middleware

HTTP middleware for JWT-based authentication and authorization in ToxicToastGo services.

## Overview

The `shared/middleware` package provides HTTP middleware for protecting endpoints with JWT authentication, role-based access control (RBAC), and permission-based authorization.

## Features

- **JWT Token Validation** - Validates JWT tokens from Authorization header
- **Role-Based Access Control** - Restrict endpoints to specific roles
- **Permission-Based Authorization** - Restrict endpoints to specific permissions
- **Context Integration** - Extracts user claims and stores them in request context
- **Flexible Authorization** - Support for single or multiple roles/permissions

## Installation

The middleware is part of the shared module and automatically available to all services:

```go
import "github.com/toxictoast/toxictoastgo/shared/middleware"
```

## Configuration

### JWT Secret

**CRITICAL:** All services must use the same JWT secret to validate tokens.

Set in environment variables:
```bash
# Must match across auth-service and all services using middleware
JWT_SECRET=your-secret-key-please-change-in-production
```

### Initialize Middleware

```go
import (
    "time"
    "github.com/toxictoast/toxictoastgo/shared/jwt"
    "github.com/toxictoast/toxictoastgo/shared/middleware"
)

func main() {
    // Initialize JWT helper (same secret as auth-service!)
    jwtHelper := jwt.NewJWTHelper(
        cfg.JWTSecret,
        15*time.Minute,  // Access token duration
        7*24*time.Hour,  // Refresh token duration
    )

    // Create auth middleware
    authMiddleware := middleware.NewAuthMiddleware(jwtHelper)
}
```

## Usage Examples

### 1. Protecting a Single Endpoint

```go
import (
    "github.com/gorilla/mux"
    "github.com/toxictoast/toxictoastgo/shared/middleware"
)

router := mux.NewRouter()

// Protected endpoint - requires valid JWT token
router.HandleFunc("/protected", myHandler).
    Methods("GET").
    Use(authMiddleware.Authenticate)
```

### 2. Protecting a Subrouter

```go
// All routes under /api/* require authentication
apiRouter := router.PathPrefix("/api").Subrouter()
apiRouter.Use(authMiddleware.Authenticate)

// These are all protected
apiRouter.HandleFunc("/users", listUsers).Methods("GET")
apiRouter.HandleFunc("/posts", listPosts).Methods("GET")
```

### 3. Role-Based Protection

```go
// Require 'admin' role
adminRouter := router.PathPrefix("/admin").Subrouter()
adminRouter.Use(authMiddleware.Authenticate)
adminRouter.Use(authMiddleware.RequireRole("Administrator"))
adminRouter.HandleFunc("", adminDashboard).Methods("GET")

// Require 'editor' OR 'admin' role
editorRouter := router.PathPrefix("/editor").Subrouter()
editorRouter.Use(authMiddleware.Authenticate)
editorRouter.Use(authMiddleware.RequireAnyRole("editor", "admin"))
editorRouter.HandleFunc("", editorDashboard).Methods("GET")
```

### 4. Permission-Based Protection

```go
// Require 'posts.write' permission
writeRouter := router.PathPrefix("/posts/write").Subrouter()
writeRouter.Use(authMiddleware.Authenticate)
writeRouter.Use(authMiddleware.RequirePermission("posts.write"))
writeRouter.HandleFunc("", createPost).Methods("POST")

// Require ANY of the specified permissions
moderateRouter := router.PathPrefix("/moderate").Subrouter()
moderateRouter.Use(authMiddleware.Authenticate)
moderateRouter.Use(authMiddleware.RequireAnyPermission(
    "posts.delete",
    "posts.edit",
    "posts.moderate",
))
moderateRouter.HandleFunc("", moderatePosts).Methods("GET")
```

### 5. Optional Authentication

```go
// Endpoint works for both authenticated and anonymous users
router.HandleFunc("/public", publicHandler).
    Methods("GET").
    Use(authMiddleware.AuthenticateOptional)

func publicHandler(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetClaims(r.Context())
    if claims != nil {
        // User is authenticated
        fmt.Fprintf(w, "Hello, %s!", claims.Username)
    } else {
        // Anonymous user
        fmt.Fprint(w, "Hello, guest!")
    }
}
```

### 6. Accessing Claims in Handlers

```go
import "github.com/toxictoast/toxictoastgo/shared/middleware"

func myHandler(w http.ResponseWriter, r *http.Request) {
    // Get JWT claims from context
    claims := middleware.GetClaims(r.Context())
    if claims == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Access user information
    userID := claims.UserID
    email := claims.Email
    username := claims.Username
    roles := claims.Roles
    permissions := claims.Permissions

    // Check specific role
    if middleware.HasRole(claims, "admin") {
        // User is an admin
    }

    // Check specific permission
    if middleware.HasPermission(claims, "posts.write") {
        // User can write posts
    }

    // Check multiple roles
    if middleware.HasAnyRole(claims, "editor", "moderator", "admin") {
        // User has at least one of these roles
    }
}
```

## Middleware Functions

### Authentication Middleware

#### `Authenticate(next http.Handler) http.Handler`
Requires valid JWT token in Authorization header. Rejects requests without valid token.

```go
router.Use(authMiddleware.Authenticate)
```

#### `AuthenticateOptional(next http.Handler) http.Handler`
Validates JWT token if present, but allows requests without token to proceed.

```go
router.Use(authMiddleware.AuthenticateOptional)
```

### Role-Based Middleware

#### `RequireRole(role string) func(http.Handler) http.Handler`
Requires user to have a specific role. Must be used after `Authenticate`.

```go
router.Use(authMiddleware.Authenticate)
router.Use(authMiddleware.RequireRole("Administrator"))
```

#### `RequireAnyRole(roles ...string) func(http.Handler) http.Handler`
Requires user to have at least one of the specified roles. Must be used after `Authenticate`.

```go
router.Use(authMiddleware.Authenticate)
router.Use(authMiddleware.RequireAnyRole("editor", "moderator", "admin"))
```

### Permission-Based Middleware

#### `RequirePermission(permission string) func(http.Handler) http.Handler`
Requires user to have a specific permission. Must be used after `Authenticate`.

```go
router.Use(authMiddleware.Authenticate)
router.Use(authMiddleware.RequirePermission("posts.write"))
```

#### `RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler`
Requires user to have at least one of the specified permissions. Must be used after `Authenticate`.

```go
router.Use(authMiddleware.Authenticate)
router.Use(authMiddleware.RequireAnyPermission("posts.write", "posts.edit"))
```

## Helper Functions

### Context Helpers

#### `GetClaims(ctx context.Context) *jwt.Claims`
Extracts JWT claims from request context. Returns nil if no claims found.

```go
claims := middleware.GetClaims(r.Context())
if claims != nil {
    userID := claims.UserID
}
```

### Role Checking

#### `HasRole(claims *jwt.Claims, role string) bool`
Checks if claims contain a specific role.

```go
if middleware.HasRole(claims, "admin") {
    // User is admin
}
```

#### `HasAnyRole(claims *jwt.Claims, roles ...string) bool`
Checks if claims contain any of the specified roles.

```go
if middleware.HasAnyRole(claims, "editor", "moderator") {
    // User is either editor or moderator
}
```

### Permission Checking

#### `HasPermission(claims *jwt.Claims, permission string) bool`
Checks if claims contain a specific permission.

```go
if middleware.HasPermission(claims, "posts.write") {
    // User can write posts
}
```

#### `HasAnyPermission(claims *jwt.Claims, permissions ...string) bool`
Checks if claims contain any of the specified permissions.

```go
if middleware.HasAnyPermission(claims, "posts.write", "posts.edit") {
    // User can write or edit posts
}
```

## Client Authentication

### Making Authenticated Requests

```bash
# 1. Login to get JWT token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Response:
# {
#   "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "expires_in": 900
# }

# 2. Use access token in Authorization header
curl http://localhost:8080/api/protected \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Token Format

The Authorization header must follow this format:
```
Authorization: Bearer <access_token>
```

## Error Responses

### 401 Unauthorized

Returned when:
- No Authorization header provided (with `Authenticate` middleware)
- Invalid token format
- Token signature invalid
- Token expired
- User not authenticated (when required)

```
HTTP/1.1 401 Unauthorized
Content-Type: text/plain

Authorization header required
```

### 403 Forbidden

Returned when:
- User lacks required role
- User lacks required permission

```
HTTP/1.1 403 Forbidden
Content-Type: text/plain

Forbidden: required role 'admin' not found
```

## Best Practices

### 1. Always Use HTTPS in Production
JWT tokens should only be transmitted over HTTPS to prevent token theft.

### 2. Keep Tokens Short-Lived
Access tokens should expire quickly (15 minutes recommended). Use refresh tokens for long-term access.

### 3. Validate on Every Request
Always use middleware for protected endpoints. Never skip validation.

### 4. Store Tokens Securely
- Never store tokens in localStorage (XSS vulnerability)
- Use httpOnly cookies or secure storage
- Never log tokens

### 5. Match JWT Secrets
**CRITICAL:** Ensure all services use the same JWT_SECRET environment variable.

### 6. Layer Middleware Correctly
Apply middleware in the correct order:
```go
// CORRECT ORDER:
router.Use(authMiddleware.Authenticate)          // 1. Authenticate first
router.Use(authMiddleware.RequireRole("Administrator"))  // 2. Then check role
```

### 7. Use Context Helpers
Always use `GetClaims()` to access user information from context:
```go
claims := middleware.GetClaims(r.Context())
if claims != nil {
    // Safe to use claims
}
```

## Testing

### Test with curl

```bash
# Public endpoint (no auth required)
curl http://localhost:8080/api/test/public

# Protected endpoint (requires auth)
curl http://localhost:8080/api/test/protected \
  -H "Authorization: Bearer YOUR_TOKEN"

# Admin-only endpoint
curl http://localhost:8080/api/test/admin \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Permission-based endpoint
curl http://localhost:8080/api/test/write \
  -H "Authorization: Bearer YOUR_TOKEN_WITH_PERMISSION"
```

### Example Test Endpoints

See `services/gateway-service/internal/handler/protected_handler.go` for complete example implementations:
- `/api/test/public` - Public endpoint
- `/api/test/protected` - Requires authentication
- `/api/test/admin` - Requires 'admin' role
- `/api/test/editor` - Requires 'editor' or 'admin' role
- `/api/test/write` - Requires 'posts.write' permission

## JWT Claims Structure

```go
type Claims struct {
    UserID      string   `json:"user_id"`
    Email       string   `json:"email"`
    Username    string   `json:"username"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    jwt.RegisteredClaims
}
```

Example decoded token:
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "johndoe",
  "roles": ["editor", "moderator"],
  "permissions": ["posts.write", "posts.edit"],
  "exp": 1735689600,
  "nbf": 1735688700,
  "iat": 1735688700
}
```

## Troubleshooting

### "Invalid or expired token: signature is invalid"

**Cause:** JWT_SECRET mismatch between auth-service and gateway/services.

**Solution:**
1. Check auth-service JWT_SECRET: `docker exec auth-service printenv | grep JWT_SECRET`
2. Check gateway JWT_SECRET: `docker exec gateway-service printenv | grep JWT_SECRET`
3. Ensure both use the same value (default: `your-secret-key-please-change-in-production`)
4. Rebuild and restart services after changing config

### "Authorization header required"

**Cause:** No Authorization header provided or missing Bearer prefix.

**Solution:**
```bash
# Include Authorization header with Bearer prefix
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/protected
```

### "Forbidden: required role 'X' not found"

**Cause:** User doesn't have the required role.

**Solution:**
1. Assign role to user via auth API
2. Login again to get new token with role
3. Use new token in Authorization header

## Related Documentation

- Auth Service API: `services/auth-service/QUICKSTART.md`
- JWT Helper: `shared/jwt/jwt.go`
- Gateway Integration: `services/gateway-service/README.md`

## License

Part of ToxicToastGo monorepo.
