package intercept

import (
	"api-gateway/internal/pkg/requestid"
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestProxyPreservesContextAndOwnerErrors(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	ownerError := status.Error(
		codes.InvalidArgument,
		"Длительность бронирования не разрешена для номера",
	)
	err := ProxyMetadata()(ctx, "/rental/ReserveRental", nil, nil, nil,
		func(actual context.Context, _ string, _, _ any, _ *grpc.ClientConn, opts ...grpc.CallOption) error {
			if actual != ctx {
				t.Fatal("proxy replaced the incoming context")
			}
			for _, opt := range opts {
				if trailer, ok := opt.(grpc.TrailerCallOption); ok {
					*trailer.TrailerAddr = metadata.Pairs("x-error-code", "DURATION_INVALID")
				}
			}
			return ownerError
		},
	)
	if err != ownerError {
		t.Fatalf("owner error was rewritten: %v", err)
	}
}

func TestTransportFailuresHaveRussianMessages(t *testing.T) {
	for _, test := range []struct {
		code    codes.Code
		message string
	}{
		{
			codes.Unavailable,
			"Сервис временно недоступен",
		},
		{
			codes.DeadlineExceeded,
			"Время ожидания ответа сервиса истекло",
		},
		{
			codes.Canceled,
			"Запрос отменён; результат операции может быть неизвестен",
		},
	} {
		t.Run(test.code.String(), func(t *testing.T) {
			err := ProxyMetadata()(context.Background(), "/rental/ReserveRental", nil, nil, nil,
				func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
					return status.Error(test.code, "transport details: private host")
				},
			)
			if status.Code(err) != test.code || status.Convert(err).Message() != test.message {
				t.Fatalf("transport error = %v", err)
			}
		})
	}
}

// A server context makes the correlation test exercise the real ingress interceptor.
type testTransportStream struct{}

func (testTransportStream) Method() string {
	return "/test.Service/Get"
}
func (testTransportStream) SetHeader(metadata.MD) error {
	return nil
}
func (testTransportStream) SendHeader(metadata.MD) error {
	return nil
}
func (testTransportStream) SetTrailer(metadata.MD) error {
	return nil
}

func TestProxyForwardsRequestIDWithoutMutatingCaller(t *testing.T) {
	ctx := grpc.NewContextWithServerTransportStream(t.Context(), testTransportStream{})
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-request-id", "test-request"))
	original := metadata.Pairs("other", "keep", "x-request-id", "old")
	ctx = metadata.NewOutgoingContext(ctx, original)
	options := []grpc.CallOption{grpc.WaitForReady(false), grpc.MaxCallRecvMsgSize(123), grpc.MaxCallSendMsgSize(456)}
	sentinel := options[1]
	called := false
	_, err := requestid.Interceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, _ any) (any, error) {
		return nil, ProxyMetadata()(ctx, "/test.Service/Get", nil, nil, nil,
			func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				called = true
				md, _ := metadata.FromOutgoingContext(ctx)
				if got := md.Get("x-request-id"); len(got) != 1 || got[0] != "test-request" {
					t.Fatalf("incorrect outgoing request ID: %v", got)
				}
				if md.Get("other")[0] != "keep" {
					t.Fatal("lost caller metadata")
				}
				return nil
			}, options[:1]...)
	})
	if err != nil || !called {
		t.Fatalf("call failed: %v", err)
	}
	if original.Get("x-request-id")[0] != "old" || options[1] != sentinel {
		t.Fatal("modified caller-owned metadata or option slice")
	}
}
