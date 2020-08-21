package accord

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ServerAuthInterceptor is a server interceptor for authentication and authorization
type ServerAuthInterceptor struct {
	allowedRoles map[string][]string
	jwtManager   *JWTManager
}

// NewServerAuthInterceptor returns a new auth interceptor
func NewServerAuthInterceptor(jwtManager *JWTManager) *ServerAuthInterceptor {
	return &ServerAuthInterceptor{
		allowedRoles: make(map[string][]string),
		jwtManager:   jwtManager,
	}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *ServerAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		err := interceptor.Authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *ServerAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		err := interceptor.Authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *ServerAuthInterceptor) Authorize(ctx context.Context, method string) error {
	allowedRoles, ok := interceptor.allowedRoles[method]
	if !ok {
		// everyone can access
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	claims, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	for _, role := range allowedRoles {
		if role == claims.Role {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}

// ClientAuthInterceptor is a client interceptor for authentication
type ClientAuthInterceptor struct {
	authClient  *AuthClient
	username    string
	password    string
	accessToken string
}

// NewClientAuthInterceptor returns a new auth interceptor
func NewClientAuthInterceptor(
	authClient *AuthClient,
	username string,
	password string,
	refreshDuration time.Duration,
) (*ClientAuthInterceptor, error) {
	interceptor := &ClientAuthInterceptor{
		authClient: authClient,
		username:   username,
		password:   password,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

// Unary returns a client interceptor to authenticate unary RPC
func (intr *ClientAuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log.Printf("--> unary interceptor: %s", method)

		return invoker(intr.attachToken(ctx), method, req, reply, cc, opts...)
	}
}

// Stream returns a client interceptor to authenticate stream RPC
func (intr *ClientAuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %s", method)

		return streamer(intr.attachToken(ctx), desc, cc, method, opts...)
	}
}

func (intr *ClientAuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", intr.accessToken)
}

func (intr *ClientAuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := intr.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			time.Sleep(wait)
			err := intr.refreshToken()
			if err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()

	return nil
}

func (intr *ClientAuthInterceptor) refreshToken() error {
	accessToken, err := intr.authClient.Login(intr.username, intr.password)
	if err != nil {
		return err
	}

	intr.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}
