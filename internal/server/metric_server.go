package server

import (
	"FairLAP/internal/domain/service/metrics"
	"FairLAP/pkg/failure"
	"net/http"
	"strconv"
)

type MetricServer struct {
	metrics *metrics.Service
}

func NewMetricServer(metrics *metrics.Service) *MetricServer {
	return &MetricServer{
		metrics: metrics,
	}
}

func (s *MetricServer) GetLaps(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	laps, err := s.metrics.GetLaps(ctx)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, laps, http.StatusOK)
}

func (s *MetricServer) GetGroupMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groupId, err := strconv.Atoi(r.FormValue("group_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid group_id"))
		return
	}

	metric, err := s.metrics.GetGroupMetric(ctx, groupId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, metric, http.StatusOK)
}
