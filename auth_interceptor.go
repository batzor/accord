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

// AuthInterceptorServer is a server interceptor for authentication and authorization
type AuthInterceptorServer struct {
	allowedRoles map[string][]string
	jwtManager   *JWTManager
}

// NewAuthInterceptorServer returns a new auth interceptor
func NewAuthInterceptorServer(jwtManager *JWTManager) *AuthInterceptorServer {
	return &AuthInterceptorServer{
		allowedRoles: make(map[string][]string),
		jwtManager:   jwtManager,
	}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptorServer) Unary() grpc.UnaryServerInterceptor {
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
func (interceptor *AuthInterceptorServer) Stream() grpc.StreamServerInterceptor {
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

func (interceptor *AuthInterceptorServer) Authorize(ctx context.Context, method string) error {
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

// AuthInterceptorClient is a client interceptor for authentication
type AuthInterceptorClient struct {
	authClient  *AuthClient
	username    string
	password    string
	accessToken string
}

// NewAuthInterceptorClient returns a new auth interceptor
func NewAuthInterceptorClient(
	authClient *AuthClient,
	username string,
	password string,
	refreshDuration time.Duration,
) (*AuthInterceptorClient, error) {
	interceptor := &AuthInterceptorClient{
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
func (intr *AuthInterceptorClient) Unary() grpc.UnaryClientInterceptor {
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
func (intr *AuthInterceptorClient) Stream() grpc.StreamClientInterceptor {
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

func (intr *AuthInterceptorClient) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", intr.accessToken)
}

func (intr *AuthInterceptorClient) scheduleRefreshToken(refreshDuration time.Duration) error {
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

func (intr *AuthInterceptorClient) refreshToken() error {
	accessToken, err := intr.authClient.Login(intr.username, intr.password)
	if err != nil {
		return err
	}

	intr.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}
