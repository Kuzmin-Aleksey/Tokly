package images

import (
	"FairLAP/pkg/failure"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Images struct {
	path string
}

func New(path string) *Images {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	return &Images{
		path: path,
	}
}

func (images *Images) Save(groupId int, img image.Image) (uuid.UUID, error) {
	path := filepath.Join(images.path, strconv.Itoa(groupId))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return uuid.Nil, fmt.Errorf("make dir failed: %w", err)
		}
	}

	uid := uuid.New()

	path = filepath.Join(path, uid.String()+".jpeg")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return uuid.Nil, fmt.Errorf("open file failed: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (images *Images) SaveMask(groupId int, uid uuid.UUID, mask io.Reader) error {
	path := filepath.Join(images.path, strconv.Itoa(groupId))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("make dir failed: %w", err)
		}
	}
	path = filepath.Join(path, fmt.Sprintf("%s_mask.png", uid))

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, mask); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

func (images *Images) Open(groupId int, uid uuid.UUID) (*os.File, error) {
	path := filepath.Join(images.path, strconv.Itoa(groupId), uid.String()+".jpeg")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, failure.NewNotFoundError(err.Error())
		}
		return nil, fmt.Errorf("open file failed: %w", err)
	}

	return f, nil
}

func (images *Images) OpenMask(groupId int, uid uuid.UUID) (*os.File, error) {
	path := filepath.Join(images.path, strconv.Itoa(groupId), fmt.Sprintf("%s_mask.png", uid))

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, failure.NewNotFoundError(err.Error())
		}
		return nil, fmt.Errorf("open file failed: %w", err)
	}

	return f, nil
}

func (images *Images) DeleteGroup(groupId int) error {
	path := filepath.Join(images.path, strconv.Itoa(groupId))
	if err := os.RemoveAll(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
