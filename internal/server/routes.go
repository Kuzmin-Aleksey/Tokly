package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) InitRoutes(rtr *mux.Router) {
	rtr.HandleFunc("/detect", s.detector.Detect).Methods(http.MethodPost)

	rtr.HandleFunc("/groups/create", s.groups.CreateGroup).Methods(http.MethodPost)
	rtr.HandleFunc("/groups/delete", s.groups.DeleteGroup).Methods(http.MethodDelete)

	rtr.HandleFunc("/metric/laps", s.metrics.GetLaps).Methods(http.MethodGet)
	rtr.HandleFunc("/metric/group", s.metrics.GetGroupMetric).Methods(http.MethodGet)

	rtr.HandleFunc("/image/{group_id}/{image_uid}.jpeg", s.images.HandleFile).Methods(http.MethodGet)
}
