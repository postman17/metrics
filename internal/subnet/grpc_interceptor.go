package subnet

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCInterceptor возвращает UnaryServerInterceptor, который проверяет,
// что IP-адрес агента из метаданных x-real-ip принадлежит доверенной подсети.
// Если проверка не пройдена, возвращается ошибка codes.PermissionDenied.
func GRPCInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		slog.Error("grpc cidr configuration error", "err", err)
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return nil, status.Error(codes.Internal, "invalid network configuration")
		}
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if trustedNet == nil {
			return nil, status.Error(codes.Internal, "invalid network configuration")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 || values[0] == "" {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		realIP := net.ParseIP(values[0])
		if realIP == nil {
			return nil, status.Error(codes.PermissionDenied, "invalid ip format")
		}

		if !trustedNet.Contains(realIP) {
			slog.Warn("grpc access forbidden", "ip", values[0], "subnet", trustedSubnet)
			return nil, status.Error(codes.PermissionDenied, "ip address is not trusted")
		}

		return handler(ctx, req)
	}
}
