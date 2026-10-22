package requestid_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	rentalV1 "api-gateway/internal/pkg/pb/api-gateway/rental/v1"
	"api-gateway/internal/pkg/requestid"
)

type transportStream struct {
	headers metadata.MD
}

func (s *transportStream) Method() string {
	return rentalV1.RentalService_ReserveRental_FullMethodName
}

func (s *transportStream) SetHeader(md metadata.MD) error {
	s.headers = metadata.Join(s.headers, md)
	return nil
}

func (s *transportStream) SendHeader(md metadata.MD) error {
	return s.SetHeader(md)
}

func (s *transportStream) SetTrailer(metadata.MD) error {
	return nil
}

func TestRequestIDBeforeHandler(t *testing.T) {
	for _, test := range []struct {
		name string
		id   string
		role string
		code codes.Code
	}{
		{
			name: "передаёт полученный ID в контекст и ответ",
			id:   "87037374-4b0a-48bd-94ba-a8efbe21f1dc",
			role: "GUEST",
			code: codes.OK,
		},
		{
			name: "создаёт UUID при отсутствии заголовка",
			role: "GUEST",
			code: codes.OK,
		},
		{
			name: "возвращает ID даже при отказе обработчика",
			role: "RECEPTION",
			code: codes.PermissionDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			md := metadata.Pairs("x-demo-role", test.role, "x-demo-actor", "guest")
			if test.id != "" {
				md.Set("x-request-id", test.id)
			}
			stream := &transportStream{}
			ctx := grpc.NewContextWithServerTransportStream(metadata.NewIncomingContext(t.Context(), md), stream)
			info := &grpc.UnaryServerInfo{
				FullMethod: rentalV1.RentalService_ReserveRental_FullMethodName,
			}
			called := false
			var downstreamID string
			_, err := requestid.Interceptor()(ctx, nil, info, func(ctx context.Context, request any) (any, error) {
				if test.code != codes.OK {
					return nil, status.Error(test.code, "Операция отклонена")
				}
				return func(ctx context.Context) (any, error) {
					called = true
					downstreamID = requestid.FromContext(ctx)
					return nil, nil
				}(ctx)
			})
			if status.Code(err) != test.code || called != (test.code == codes.OK) {
				t.Fatalf("ошибка=%v, handler вызван=%t", err, called)
			}
			values := stream.headers.Get("x-request-id")
			if len(values) != 1 {
				t.Fatalf("заголовок x-request-id: %v", values)
			}
			id := values[0]
			if _, err := uuid.FromString(id); err != nil {
				t.Fatal(err)
			}
			if test.id != "" && id != test.id {
				t.Fatalf("изменился входящий ID: %s", id)
			}
			if called && downstreamID != id {
				t.Fatalf("caller=%s, ответ=%s", downstreamID, id)
			}
		})
	}
}
