package yolo_model

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"gocv.io/x/gocv"
	"image"
	"log"
)

type ModelConfig struct {
	ClassList     []string `yaml:"class-list" json:"class-list"`
	Size          Size     `yaml:"size" json:"size"`
	ConfThreshold float32  `yaml:"conf-threshold" json:"conf-threshold"`
	NMSThreshold  float32  `yaml:"NMS-threshold" json:"NMS-threshold"`
}

type Size struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

func ReadConfig(path string) (*ModelConfig, error) {
	cfg := new(ModelConfig)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type Model struct {
	net *gocv.Net
	cfg *ModelConfig
}

func NewModel(modelPath string, cfg *ModelConfig) *Model {
	net := gocv.ReadNetFromONNX(modelPath)
	if net.Empty() {
		log.Fatal("failed to load ONNX model from ", modelPath)
	}

	if err := net.SetPreferableBackend(gocv.NetBackendOpenCV); err != nil {
		log.Fatal(err)
	}

	return &Model{
		net: &net,
		cfg: cfg,
	}
}

type Detection struct {
	ClassID    int
	ClassName  string
	Confidence float32
	BBox       image.Rectangle
}

func (m *Model) Detect(img image.Image) ([]Detection, error) {
	mat, err := imageToMat(img)
	if err != nil {
		return nil, fmt.Errorf("read image failed: %w", err)
	}
	defer mat.Close()

	blob := gocv.BlobFromImage(mat, 1.0/255.0, image.Pt(m.cfg.Size.Width, m.cfg.Size.Height), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	m.net.SetInput(blob, "images")
	output := m.net.Forward("output0")
	defer output.Close()

	detections, err := m.processYOLOv8Output(output, mat.Cols(), mat.Rows())
	if err != nil {
		return nil, err
	}

	return detections, nil
}

func (m *Model) processYOLOv8Output(output gocv.Mat, origWidth, origHeight int) ([]Detection, error) {
	sizes := output.Size()
	if len(sizes) != 3 || sizes[0] != 1 {
		log.Fatalf("Неожиданный формат вывода: %v", sizes)
	}

	numFeatures := sizes[1]
	numPredictions := sizes[2]

	data, err := output.DataPtrFloat32()
	if err != nil {
		return nil, fmt.Errorf("get data ptr failed: %w", err)
	}

	predictions := make([][]float32, numPredictions)
	for i := range predictions {
		predictions[i] = make([]float32, numFeatures)
	}

	for feature := 0; feature < numFeatures; feature++ {
		for pred := 0; pred < numPredictions; pred++ {
			predictions[pred][feature] = data[feature*numPredictions+pred]
		}
	}

	var boxes []image.Rectangle
	var confidences []float32
	var classIDs []int

	for _, pred := range predictions {
		cx, cy, w, h := pred[0], pred[1], pred[2], pred[3]

		classData := pred[4:]
		maxProb := float32(0)
		maxClass := 0
		for i, val := range classData {
			if val > maxProb {
				maxProb = val
				maxClass = i
			}
		}

		if maxProb < m.cfg.ConfThreshold {
			continue
		}

		x1 := cx - w/2
		y1 := cy - h/2
		x2 := cx + w/2
		y2 := cy + h/2

		scaleX := float32(origWidth) / float32(m.cfg.Size.Width)
		scaleY := float32(origHeight) / float32(m.cfg.Size.Height)

		x1 = x1 * scaleX
		y1 = y1 * scaleY
		x2 = x2 * scaleX
		y2 = y2 * scaleY

		ix1 := int(x1)
		iy1 := int(y1)
		ix2 := int(x2)
		iy2 := int(y2)

		ix1 = clamp(ix1, 0, origWidth)
		iy1 = clamp(iy1, 0, origHeight)
		ix2 = clamp(ix2, 0, origWidth)
		iy2 = clamp(iy2, 0, origHeight)

		if ix2 <= ix1 || iy2 <= iy1 {
			continue
		}

		boxes = append(boxes, image.Rect(ix1, iy1, ix2, iy2))
		confidences = append(confidences, maxProb)
		classIDs = append(classIDs, maxClass)
	}

	if len(boxes) == 0 {
		return []Detection{}, nil
	}

	indices := gocv.NMSBoxes(boxes, confidences, m.cfg.ConfThreshold, m.cfg.NMSThreshold)

	var detections []Detection
	for _, idx := range indices {
		d := Detection{
			ClassID:    classIDs[idx],
			Confidence: confidences[idx],
			BBox:       boxes[idx],
		}

		if len(m.cfg.ClassList) <= d.ClassID {
			log.Fatal("Class id out of range. Check model settings.")
		}

		d.ClassName = m.cfg.ClassList[d.ClassID]

		detections = append(detections, d)

	}

	return detections, nil
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func imageToMat(img image.Image) (gocv.Mat, error) {
	bounds := img.Bounds()
	x, y := bounds.Dx(), bounds.Dy()

	bytes := make([]byte, 0, x*y)
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			r, g, b, _ := img.At(i, j).RGBA()
			bytes = append(bytes, byte(b>>8))
			bytes = append(bytes, byte(g>>8))
			bytes = append(bytes, byte(r>>8))
		}
	}

	return gocv.NewMatFromBytes(y, x, gocv.MatTypeCV8UC3, bytes)
}

func (m *Model) Close() {
	m.net.Close()
}
