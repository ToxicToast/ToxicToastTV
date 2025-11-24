package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// AuthInterceptor is a gRPC unary interceptor that extracts user information from metadata
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract user info from metadata
	user, err := ExtractUserFromMetadata(ctx)
	if err != nil {
		// No user info in metadata - continue without auth
		// (some endpoints might not require auth)
		return handler(ctx, req)
	}

	// Inject user info into context for handlers to use
	ctx = InjectUserIntoContext(ctx, user)

	// Continue with the handler
	return handler(ctx, req)
}

// StreamAuthInterceptor is a gRPC stream interceptor that extracts user information from metadata
func StreamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	// Extract user info from metadata
	user, err := ExtractUserFromMetadata(ctx)
	if err != nil {
		// No user info in metadata - continue without auth
		return handler(srv, ss)
	}

	// Inject user info into context
	ctx = InjectUserIntoContext(ctx, user)

	// Wrap the stream with the new context
	wrappedStream := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	return handler(srv, wrappedStream)
}

// wrappedServerStream wraps grpc.ServerStream with a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
