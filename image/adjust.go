package image

import (
	"bytes"
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

func convertToGrayByBinary(srcData []byte) (nrgba *image.NRGBA, err error) {
	img, _, err := image.Decode(bytes.NewReader(srcData))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}
	return convertToGrayByImage(img), err
}

func convertToGrayByImage(img image.Image) (nrgba *image.NRGBA) {
	return imaging.Grayscale(img)
}

func invertByImage(img image.Image) (nrgba *image.NRGBA) {
	return imaging.Invert(img)
}
