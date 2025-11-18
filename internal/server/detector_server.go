package server

import (
	"FairLAP/internal/domain/service/detector"
	"FairLAP/pkg/failure"
	"fmt"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
)

type DetectorServer struct {
	detector *detector.Service
}

func NewDetectorServer(detector *detector.Service) *DetectorServer {
	return &DetectorServer{detector: detector}
}

func (s *DetectorServer) Detect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groupId, err := strconv.Atoi(r.FormValue("group_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid group_id"))
		return
	}

	defer r.Body.Close()

	img, err := decodeImg(r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError(err.Error()))
		return
	}

	if err := s.detector.Detect(ctx, groupId, img); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func decodeImg(r io.Reader, mime string) (image.Image, error) {
	switch mime {
	case "image/png":
		return png.Decode(r)
	case "image/jpeg", "image/jpg":
		return jpeg.Decode(r)
	case "image/gif":
		return gif.Decode(r)
	case "image/webp":
		return webp.Decode(r)
	case "image/bmp":
		return bmp.Decode(r)
	case "image/tiff":
		return tiff.Decode(r)
	default:
		return nil, fmt.Errorf("unknown image type: %s", mime)
	}
}
