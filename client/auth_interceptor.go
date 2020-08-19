package client

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor is a client interceptor for authentication
type AuthInterceptor struct {
	authClient  *AuthClient
	username    string
	password    string
	accessToken string
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(
	authClient *AuthClient,
	refreshDuration time.Duration,
) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient: authClient,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

// Unary returns a client interceptor to authenticate unary RPC
func (intr *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
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
func (intr *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
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

func (intr *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", intr.accessToken)
}

func (intr *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
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

func (intr *AuthInterceptor) refreshToken() error {
	accessToken, err := intr.authClient.Login(intr.username, intr.password)
	if err != nil {
		return err
	}

	intr.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}
