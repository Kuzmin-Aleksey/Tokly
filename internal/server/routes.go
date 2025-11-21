package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) InitRoutes(rtr *mux.Router) {
	rtr.HandleFunc("/detect", s.detector.Detect).Methods(http.MethodPost)

	rtr.HandleFunc("/groups/create", s.groups.CreateGroup).Methods(http.MethodPost)
	rtr.HandleFunc("/groups/by_lap", s.groups.GetByLap).Methods(http.MethodGet)
	rtr.HandleFunc("/groups/delete", s.groups.DeleteGroup).Methods(http.MethodDelete)

	rtr.HandleFunc("/metric/laps", s.metrics.GetLaps).Methods(http.MethodGet)
	rtr.HandleFunc("/metric/group", s.metrics.GetGroupMetric).Methods(http.MethodGet)

	rtr.HandleFunc("/lap_config/get", s.lapConfig.GetLapConfig).Methods(http.MethodGet)
	rtr.HandleFunc("/lap_config/save", s.lapConfig.SaveLapConfig).Methods(http.MethodPost)

	rtr.HandleFunc("/image/{group_id}/{image_uid}.jpeg", s.images.HandleImage).Methods(http.MethodGet)
	rtr.HandleFunc("/image/{group_id}/{image_uid}_mask.png", s.images.HandleMask).Methods(http.MethodGet, http.MethodPost)
	rtr.HandleFunc("/mask/{detection_id}.png", s.mask.GetRect).Methods(http.MethodGet)
	rtr.HandleFunc("/polygon/{image_uid}.png", s.mask.GetPolygon).Methods(http.MethodGet)
}
