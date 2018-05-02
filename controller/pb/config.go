package pb

import (
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Dial will dial the controller client.
func Dial(address string, opts ...grpc.DialOption) (ControllerClient, error) {
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}
	c := NewControllerClient(conn)
	return c, nil
}

// TokenKey is the header key used to transport the lock token.
const TokenKey = "x-lock-token"

// ContextGetLockToken is used to retrieve the lock token from the context.
func ContextGetLockToken(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	if len(md[TokenKey]) >= 1 {
		return md[TokenKey][0]
	}
	return ""
}

// ContextWithLockToken uses grpc metadata to return a new context with a lock
// token.
func ContextWithLockToken(ctx context.Context, token string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.Pairs()
	}
	md = metadata.Join(md, metadata.Pairs(TokenKey, token))
	return metadata.NewOutgoingContext(ctx, md)
}
