// Package grpcserver реализует gRPC-сервис Metrics для приёма метрик от агентов.
// Метод UpdateMetrics принимает список метрик и сохраняет их через MetricsRepository.
package grpcserver

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	models "github.com/postman17/metrics/internal/model"
	repo "github.com/postman17/metrics/internal/repository"
	pb "github.com/postman17/metrics/pkg/proto"
)

// Server реализует metricspb.MetricsServer поверх хранилища MetricsRepository.
type Server struct {
	pb.UnimplementedMetricsServer
	storage repo.MetricsRepository
}

// New создаёт gRPC-сервер метрик, использующий переданное хранилище.
func New(storage repo.MetricsRepository) *Server {
	return &Server{storage: storage}
}

// UpdateMetrics принимает список метрик от агента и сохраняет их в хранилище.
// Возвращает пустой ответ при успехе либо ошибку с соответствующим кодом.
func (s *Server) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	list := protoToModels(req.GetMetrics())
	if err := s.storage.AddBatch(list); err != nil {
		slog.Warn("grpc AddBatch failed", "err", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Debug("grpc metrics updated", "count", len(list))
	return &pb.UpdateMetricsResponse{}, nil
}

// protoToModels преобразует срез proto-метрик во внутреннюю модель MetricsList.
// Gauge-метрики используют поле value, counter-метрики — поле delta.
func protoToModels(metrics []*pb.Metric) models.MetricsList {
	result := make(models.MetricsList, 0, len(metrics))
	for _, m := range metrics {
		if m == nil {
			continue
		}
		mm := models.Metrics{ID: m.GetId()}
		switch m.GetType() {
		case pb.Metric_COUNTER:
			mm.MType = models.Counter
			d := m.GetDelta()
			mm.Delta = &d
		default:
			mm.MType = models.Gauge
			v := m.GetValue()
			mm.Value = &v
		}
		result = append(result, mm)
	}
	return result
}
