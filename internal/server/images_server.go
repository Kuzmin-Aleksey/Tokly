package server

import (
	"FairLAP/internal/infrastructure/persistence/images"
	"FairLAP/pkg/failure"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

type ImagesServer struct {
	images *images.Images
}

func NewImagesServer(images *images.Images) *ImagesServer {
	return &ImagesServer{images: images}
}

func (s *ImagesServer) HandleImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	groupId, err := strconv.Atoi(vars["group_id"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid group_id"))
		return
	}
	imageUid, err := uuid.Parse(vars["image_uid"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid image_uid"))
		return
	}

	f, err := s.images.Open(groupId, imageUid)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "image/jpeg")

	io.Copy(w, f)
}

func (s *ImagesServer) HandleMask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	groupId, err := strconv.Atoi(vars["group_id"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid group_id"))
		return
	}
	imageUid, err := uuid.Parse(vars["image_uid"])
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid image_uid"))
		return
	}

	if r.Method == http.MethodPost {
		if err := s.images.SaveMask(groupId, imageUid, r.Body); err != nil {
			writeAndLogErr(ctx, w, err)
		}
	} else {
		f, err := s.images.OpenMask(groupId, imageUid)
		if err != nil {
			writeAndLogErr(ctx, w, err)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "image/png")

		io.Copy(w, f)
	}

}
