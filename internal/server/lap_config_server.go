package server

import (
	"FairLAP/internal/domain/service/lapconfig"
	"FairLAP/pkg/failure"
	"encoding/json"
	"net/http"
)

type LapConfigServer struct {
	service *lapconfig.Service
}

func NewLapConfigServer(service *lapconfig.Service) *LapConfigServer {
	return &LapConfigServer{
		service: service,
	}
}

func (s *LapConfigServer) SaveLapConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lapId := r.FormValue("lap_id")

	config := make(map[string]int)
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid config"))
		return
	}

	if err := s.service.SaveLapConfig(ctx, lapId, config); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
func (s *LapConfigServer) GetLapConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lapId := r.FormValue("lap_id")

	cfg, err := s.service.GetConfig(ctx, lapId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		writeAndLogErr(ctx, w, err)
	}
}
