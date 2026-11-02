package intercept

import (
	"api-gateway/internal/pkg/requestid"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ProxyMetadata сохраняет заголовки и trailers владельца без повторного разбора его ошибок.
func ProxyMetadata() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, options ...grpc.CallOption) error {
		if id := requestid.FromContext(ctx); id != "" {
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()
			md.Set("x-request-id", id)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		var headers, trailers metadata.MD
		options = append(append([]grpc.CallOption(nil), options...), grpc.Header(&headers), grpc.Trailer(&trailers))
		err := invoke(ctx, method, request, reply, conn, options...)
		_ = grpc.SetHeader(ctx, headers)
		_ = grpc.SetTrailer(ctx, trailers)

		if err == nil || len(trailers.Get("x-error-code")) != 0 {
			return err
		}

		// Сбой транспорта происходит до ErrorInterceptor владельца.
		switch status.Code(err) {
		case codes.Unavailable:
			return status.Error(codes.Unavailable, "Сервис временно недоступен")
		case codes.DeadlineExceeded:
			return status.Error(codes.DeadlineExceeded, "Время ожидания ответа сервиса истекло")
		case codes.Canceled:
			return status.Error(
				codes.Canceled,
				"Запрос отменён; результат операции может быть неизвестен",
			)
		default:
			return err
		}
	}
}
