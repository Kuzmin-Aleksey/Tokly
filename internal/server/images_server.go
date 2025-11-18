package server

import (
	"FairLAP/internal/infrastructure/persistence/images"
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

func (s *ImagesServer) HandleFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	groupId, _ := strconv.Atoi(vars["group_id"])
	imageUid, _ := uuid.Parse(vars["image_uid"])

	f, err := s.images.Open(groupId, imageUid)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "image/jpeg")

	io.Copy(w, f)
}
