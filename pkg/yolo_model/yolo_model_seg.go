package yolo_model

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"math"
	"sort"
	"sync"
)

type ModelSegConfig struct {
	Size          Size    `yaml:"size" json:"size"`
	ConfThreshold float32 `yaml:"conf-threshold" json:"conf-threshold"`
	NMSThreshold  float32 `yaml:"NMS-threshold" json:"NMS-threshold"`
}

func ReadSegConfig(path string) (*ModelSegConfig, error) {
	cfg := new(ModelSegConfig)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type ModelSeg struct {
	net *gocv.Net
	cfg *ModelSegConfig
	mu  sync.Mutex
}

func NewModelSeg(modelPath string, cfg *ModelSegConfig) *ModelSeg {
	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		log.Fatal("failed to load ONNX-seg model from ", modelPath)
	}

	if err := net.SetPreferableBackend(gocv.NetBackendOpenCV); err != nil {
		log.Fatal(err)
	}

	return &ModelSeg{
		net: &net,
		cfg: cfg,
	}
}

func (m *ModelSeg) DetectPolygons(img image.Image) ([][]image.Point, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mat, err := imageToMat(img)
	if err != nil {
		return nil, fmt.Errorf("read image failed: %w", err)
	}
	defer mat.Close()

	blob := gocv.BlobFromImage(mat, 1.0/255.0, image.Pt(m.cfg.Size.Width, m.cfg.Size.Height), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	m.net.SetInput(blob, "images")

	outputDet := m.net.Forward("output0") // [1,37,8400]
	defer outputDet.Close()

	outputMask := m.net.Forward("output1") // [1,32,160,160]
	defer outputMask.Close()

	polygons, err := m.postProcess(outputDet, outputMask, m.cfg.ConfThreshold, m.cfg.NMSThreshold,
		image.Pt(img.Bounds().Dx(), img.Bounds().Dy()),
	)
	if err != nil {
		return nil, err
	}

	polygons, err = filterDuplicateMasks(polygons, 0.8)
	if err != nil {
		return nil, err
	}

	points := make([][]image.Point, len(polygons))

	for i, polygon := range polygons {
		points[i] = polygon.Mask
	}

	return points, nil
}

func (m *ModelSeg) Close() {
	m.net.Close()
}

type DetectionSeg struct {
	Score float32
	Mask  []image.Point
	BBox  image.Rectangle
}

func (m *ModelSeg) postProcess(outputDet gocv.Mat, outputMask gocv.Mat, confThreshold float32, nmsThreshold float32, origSize image.Point) ([]DetectionSeg, error) {
	var detections []DetectionSeg

	sizes := outputDet.Size()
	if len(sizes) != 3 || sizes[0] != 1 || sizes[1] != 37 || sizes[2] != 8400 {
		return nil, fmt.Errorf("incorrect detection output: %v, expected [1,37,8400]", sizes)
	}

	data, err := outputDet.DataPtrFloat32()
	if err != nil {
		return nil, fmt.Errorf("error get detection data: %v", err)
	}

	scaleX := float32(origSize.X) / float32(m.cfg.Size.Width)
	scaleY := float32(origSize.Y) / float32(m.cfg.Size.Height)

	for i := 0; i < 8400; i++ {
		rawData := make([]float32, 37)
		for j := 0; j < 37; j++ {
			idx := j*8400 + i
			if idx < len(data) {
				rawData[j] = data[idx]
			}
		}

		x := rawData[0]
		y := rawData[1]
		w := rawData[2]
		h := rawData[3]

		objectness := rawData[4]

		finalScore := objectness

		if finalScore < confThreshold {
			continue
		}

		x1 := (x - w/2) * scaleX
		y1 := (y - h/2) * scaleY
		x2 := (x + w/2) * scaleX
		y2 := (y + h/2) * scaleY

		rect := image.Rect(
			int(math.Max(0, float64(x1))),
			int(math.Max(0, float64(y1))),
			int(math.Min(float64(origSize.X), float64(x2))),
			int(math.Min(float64(origSize.Y), float64(y2))),
		)

		maskWeights := rawData[5:37]

		maskPoints, err := processMask(maskWeights, outputMask, origSize)
		if err != nil {
			return nil, err
		}

		if len(maskPoints) > 0 {
			detections = append(detections, DetectionSeg{
				Score: finalScore,
				Mask:  maskPoints,
				BBox:  rect,
			})
		}
	}

	return nms(detections, nmsThreshold), nil
}

func processMask(maskWeights []float32, maskProto gocv.Mat, origSize image.Point) ([]image.Point, error) {
	protoSize := maskProto.Size()
	if len(protoSize) != 4 || protoSize[0] != 1 || protoSize[1] != 32 {
		return nil, fmt.Errorf("incorrect detection output: %v, expected [1,32,160,160]", protoSize)
	}

	protoH, protoW := protoSize[2], protoSize[3] // 160, 160

	protoData, err := maskProto.DataPtrFloat32()
	if err != nil {
		return nil, fmt.Errorf("maskProto.DataPtrFloat32: %w", err)
	}

	mask160 := gocv.NewMatWithSize(protoH, protoW, gocv.MatTypeCV32F)
	defer mask160.Close()

	maskData, err := mask160.DataPtrFloat32()
	if err != nil {
		return nil, fmt.Errorf("mask160.DataPtrFloat32: %w", err)
	}

	for y := 0; y < protoH; y++ {
		for x := 0; x < protoW; x++ {
			var sum float32
			for i := 0; i < len(maskWeights); i++ {
				idx := i*protoH*protoW + y*protoW + x
				if idx < len(protoData) {
					sum += maskWeights[i] * protoData[idx]
				}
			}

			sigmoid := float32(1.0 / (1.0 + math.Exp(float64(-sum))))
			maskData[y*protoW+x] = sigmoid
		}
	}

	fullSizeMask := gocv.NewMat()
	defer fullSizeMask.Close()
	if err := gocv.Resize(mask160, &fullSizeMask, image.Pt(origSize.X, origSize.Y), 0, 0, gocv.InterpolationLinear); err != nil {
		return nil, fmt.Errorf("gocv.Resize: %w", err)
	}

	binaryMask := gocv.NewMat()
	defer binaryMask.Close()
	gocv.Threshold(fullSizeMask, &binaryMask, 0.5, 1.0, gocv.ThresholdBinary)

	binaryMaskU8 := gocv.NewMat()
	defer binaryMaskU8.Close()
	if err := binaryMask.ConvertTo(&binaryMaskU8, gocv.MatTypeCV8U); err != nil {
		return nil, fmt.Errorf("binaryMask.ConvertTo: %w", err)
	}

	contours := gocv.FindContours(binaryMaskU8, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	if contours.Size() == 0 {
		return nil, nil
	}

	var largestContour gocv.PointVector
	maxArea := 0.0
	for i := range contours.Size() {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > maxArea {
			maxArea = area
			largestContour = contour
		}
	}

	if largestContour.IsNil() || largestContour.Size() == 0 {
		return nil, nil
	}

	epsilon := 0.005 * gocv.ArcLength(largestContour, true)
	simplified := gocv.ApproxPolyDP(largestContour, epsilon, true)

	return simplified.ToPoints(), nil
}

func nms(detections []DetectionSeg, threshold float32) []DetectionSeg {
	if len(detections) == 0 {
		return detections
	}

	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Score > detections[j].Score
	})

	var result []DetectionSeg
	suppressed := make([]bool, len(detections))

	for i := 0; i < len(detections); i++ {
		if suppressed[i] {
			continue
		}

		result = append(result, detections[i])

		for j := i + 1; j < len(detections); j++ {
			if suppressed[j] {
				continue
			}

			iou := calculateIoU(detections[i].BBox, detections[j].BBox)
			if iou > threshold {
				suppressed[j] = true
			}
		}
	}

	return result
}

func calculateIoU(a, b image.Rectangle) float32 {
	intersection := a.Intersect(b)
	if intersection.Empty() {
		return 0
	}

	areaIntersection := intersection.Dx() * intersection.Dy()
	areaA := a.Dx() * a.Dy()
	areaB := b.Dx() * b.Dy()

	return float32(areaIntersection) / float32(areaA+areaB-areaIntersection)
}

func filterDuplicateMasks(detections []DetectionSeg, iouThreshold float32) ([]DetectionSeg, error) {
	if len(detections) <= 1 {
		return detections, nil
	}

	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Score > detections[j].Score
	})

	var result []DetectionSeg
	used := make([]bool, len(detections))

	for i := 0; i < len(detections); i++ {
		if used[i] {
			continue
		}

		result = append(result, detections[i])
		used[i] = true

		for j := i + 1; j < len(detections); j++ {
			if used[j] {
				continue
			}

			maskIoU, err := calculateMaskIoU(detections[i].Mask, detections[j].Mask)
			if err != nil {
				return nil, err
			}
			if maskIoU > iouThreshold {
				used[j] = true
			}
		}
	}

	return result, nil
}

func calculateMaskIoU(mask1, mask2 []image.Point) (float32, error) {
	if len(mask1) == 0 || len(mask2) == 0 {
		return 0, nil
	}

	minX, minY, maxX, maxY := getBoundingBox(append(mask1, mask2...))
	width := maxX - minX + 1
	height := maxY - minY + 1

	if width <= 0 || height <= 0 {
		return 0, nil
	}

	mat1 := gocv.NewMatWithSize(width, height, gocv.MatTypeCV8U)
	defer mat1.Close()
	mat2 := gocv.NewMatWithSize(width, height, gocv.MatTypeCV8U)
	defer mat2.Close()

	points1 := gocv.NewPointsVector()
	for _, p := range mask1 {
		points1.Append(gocv.NewPointVectorFromPoints([]image.Point{{X: p.X - minX, Y: p.Y - minY}}))
	}
	points2 := gocv.NewPointsVector()
	for _, p := range mask2 {
		points1.Append(gocv.NewPointVectorFromPoints([]image.Point{{X: p.X - minX, Y: p.Y - minY}}))
	}

	if err := gocv.FillPoly(&mat1, points1, color.RGBA{A: 255}); err != nil {
		return 0, fmt.Errorf("gocv.FillPoly: %w", err)
	}
	if err := gocv.FillPoly(&mat2, points2, color.RGBA{A: 255}); err != nil {
		return 0, fmt.Errorf("gocv.FillPoly: %w", err)
	}

	and := gocv.NewMat()
	defer and.Close()
	if err := gocv.BitwiseAnd(mat1, mat2, &and); err != nil {
		return 0, fmt.Errorf("gocv.BitwiseAnd: %w", err)
	}

	or := gocv.NewMat()
	defer or.Close()
	if err := gocv.BitwiseAnd(mat1, mat2, &or); err != nil {
		return 0, fmt.Errorf("gocv.BitwiseAnd: %w", err)
	}

	andSum := gocv.CountNonZero(and)
	orSum := gocv.CountNonZero(or)

	if orSum == 0 {
		return 0, nil
	}

	return float32(andSum) / float32(orSum), nil
}

func getBoundingBox(points []image.Point) (minX, minY, maxX, maxY int) {
	if len(points) == 0 {
		return 0, 0, 0, 0
	}

	minX, minY = points[0].X, points[0].Y
	maxX, maxY = points[0].X, points[0].Y

	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	return minX, minY, maxX, maxY
}
