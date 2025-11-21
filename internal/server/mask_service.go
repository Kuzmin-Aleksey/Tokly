package server

import (
	"FairLAP/internal/domain/service/mask"
	"FairLAP/pkg/failure"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"image/png"
	"net/http"
	"strconv"
)

type MaskServer struct {
	service *mask.Service
}

func NewMaskService(service *mask.Service) *MaskServer {
	return &MaskServer{
		service: service,
	}
}

func (s *MaskServer) GetRect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	detectionId, err := strconv.Atoi(vars["detection_id"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid detection_id"))
		return
	}

	rect, err := s.service.GetRectMask(ctx, detectionId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")

	if err := png.Encode(w, rect); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *MaskServer) GetPolygon(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	imageUid, err := uuid.Parse(vars["image_uid"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid image_uid"))
		return
	}

	polygon, err := s.service.GetPolygonMask(ctx, imageUid)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")

	if err := png.Encode(w, polygon); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
