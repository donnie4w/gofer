// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/image
package image

import (
	"bytes"
	"fmt"
	"image"
	"math"

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

func cropImage(img image.Image, width, height int, x, y int) (image.Image, error) {
	if x < 0 || y < 0 || x+width > img.Bounds().Dx() || y+height > img.Bounds().Dy() {
		return nil, fmt.Errorf("the clipping area is out of range")
	}
	cropRect := image.Rect(x, y, x+width, y+height)
	croppedImg := imaging.Crop(img, cropRect)
	return croppedImg, nil
}

func cropImageByAnchor(img image.Image, width, height int, x, y int) (image.Image, error) {
	bounds := img.Bounds()
	if x >= 0 && y >= 0 && x+width <= bounds.Dx() && y+height <= bounds.Dy() {
		cropRect := image.Rect(x, y, x+width, y+height)
		croppedImg := imaging.Crop(img, cropRect)
		return croppedImg, nil
	}
	newWidth := width
	newHeight := height
	if x+width > bounds.Dx() {
		newWidth = bounds.Dx() - x
	}
	if y+height > bounds.Dy() {
		newHeight = bounds.Dy() - y
	}
	cropRect := image.Rect(x, y, x+newWidth, y+newHeight)
	croppedImg := imaging.Crop(img, cropRect)

	return croppedImg, nil
}

func cropImageBySide(img image.Image, width, height int, x, y int) (image.Image, error) {
	bounds := img.Bounds()

	x = clamp(x, 0, bounds.Dx()-width)
	y = clamp(y, 0, bounds.Dy()-height)

	cropRect := image.Rect(x, y, x+width, y+height)
	croppedImg := imaging.Crop(img, cropRect)

	return croppedImg, nil
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func blurGaussianImage(img image.Image, sigma float64) image.Image {
	return imaging.Blur(img, sigma)
}

func scaleImageWithRatio(img image.Image, newWidth, newHeight int, maxPixel int, lowerMode bool) (image.Image, error) {
	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	if newWidth == 0 {
		newWidth = 1
	}
	if newHeight == 0 {
		newHeight = 1
	}
	aspectRatio := float64(originalWidth) / float64(originalHeight)
	if lowerMode {
		if aspectRatio > float64(newWidth)/float64(newHeight) {
			newHeight = int(float64(newWidth) / aspectRatio)
		} else {
			newWidth = int(float64(newHeight) * aspectRatio)
		}
	} else {
		if aspectRatio > float64(newWidth)/float64(newHeight) {
			newWidth = int(float64(newHeight) * aspectRatio)
		} else {
			newHeight = int(float64(newWidth) / aspectRatio)
		}
	}

	if maxPixel > 0 && newWidth*newHeight > maxPixel {
		newWidth, newHeight = scaleDownToLimit(newWidth, newHeight, maxPixel)
	}

	resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	return resizedImg, nil
}

func scaleDownToLimit(a, b, limit int) (int, int) {
	if a <= 0 || b <= 0 || limit <= 0 {
		return 0, 0
	}
	if a*b < limit {
		return a, b
	}
	fs := float64(limit) / float64(a*b)
	sqrt := math.Sqrt(fs)
	return int(float64(a) * sqrt), int(float64(b) * sqrt)
}
