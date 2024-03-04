// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/image
package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"sort"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/donnie4w/gofer/buffer"
	"github.com/donnie4w/ico"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

type ResizeType int
type Mode int8

const (
	SCALE ResizeType = iota
	THUMBNAIL
)

const (
	Mode0 Mode = iota
	Mode1
	Mode2
	Mode3
	Mode4
	Mode5
)

type Options struct {
	Gray       bool
	Invert     bool
	Format     string
	Rotate     int
	FlipH      bool
	FlipV      bool
	Colors     int
	Quality    int
	CropAnchor []int
	CropSide   []int
	Blur       float64
	ScaleUpper []int
	ScaleLower []int
}

type ResampleFilter int

const (
	// NearestNeighbor is a nearest-neighbor filter (no anti-aliasing).
	NearestNeighbor ResampleFilter = iota

	// Box filter (averaging pixels).
	Box

	// Linear filter.
	Linear

	// Hermite cubic spline filter (BC-spline; B=0; C=0).
	Hermite

	// MitchellNetravali is Mitchell-Netravali cubic filter (BC-spline; B=1/3; C=1/3).
	MitchellNetravali

	// CatmullRom is a Catmull-Rom - sharp cubic filter (BC-spline; B=0; C=0.5).
	CatmullRom

	// BSpline is a smooth cubic filter (BC-spline; B=1; C=0).
	BSpline

	// Gaussian is a Gaussian blurring filter.
	Gaussian

	// Bartlett is a Bartlett-windowed sinc filter (3 lobes).
	Bartlett

	// Lanczos filter (3 lobes).
	Lanczos

	// Hann is a Hann-windowed sinc filter (3 lobes).
	Hann

	// Hamming is a Hamming-windowed sinc filter (3 lobes).
	Hamming

	// Blackman is a Blackman-windowed sinc filter (3 lobes).
	Blackman

	// Welch is a Welch-windowed sinc filter (parabolic window, 3 lobes).
	Welch

	// Cosine is a Cosine-windowed sinc filter (3 lobes).
	Cosine
)

type Image struct {
	ResizeFilter ResampleFilter
}

func (t *Image) Encode(srcData []byte, width, height int, mode Mode, options *Options) (destData []byte, err error) {
	defer func() {
		if er := recover(); er != nil {
			err = errors.New(fmt.Sprint(er))
		}
	}()
	img, itype, er := image.Decode(bytes.NewReader(srcData))

	if er != nil {
		return srcData, nil
	}

	if options == nil {
		options = &Options{}
	}

	if options.CropAnchor != nil && len(options.CropAnchor) == 4 {
		if i, err := cropImageByAnchor(img, options.CropAnchor[0], options.CropAnchor[1], options.CropAnchor[2], options.CropAnchor[3]); err == nil {
			img = i
		}
	}

	if options.CropSide != nil && len(options.CropSide) == 4 {
		if i, err := cropImageBySide(img, options.CropSide[0], options.CropSide[1], options.CropSide[2], options.CropSide[3]); err == nil {
			img = i
		}
	}

	if options.ScaleUpper != nil && len(options.ScaleUpper) >= 2 {
		maxPixel := 0
		if len(options.ScaleUpper) == 3 {
			maxPixel = options.ScaleUpper[2]
		}
		if i, err := scaleImageWithRatio(img, options.ScaleUpper[0], options.ScaleUpper[1], maxPixel, false); err == nil {
			img = i
		}
	}

	if options.ScaleLower != nil && len(options.ScaleLower) >= 2 {
		maxPixel := 0
		if len(options.ScaleLower) == 3 {
			maxPixel = options.ScaleLower[2]
		}
		if i, err := scaleImageWithRatio(img, options.ScaleLower[0], options.ScaleLower[1], maxPixel, true); err == nil {
			img = i
		}
	}

	if width > 0 || height > 0 {
		w := img.Bounds().Dx()
		h := img.Bounds().Dy()
		nw, nh, resizeType := praseMode(mode, w, h, width, height)
		switch resizeType {
		case SCALE:
			img = imaging.Resize(img, nw, nh, t.selectFilter())
		case THUMBNAIL:
			img = imaging.Fill(img, nw, nh, imaging.Center, t.selectFilter())
		}
	}

	if options.Gray {
		img = convertToGrayByImage(img)
	}

	if options.Invert {
		img = invertByImage(img)
	}

	if options.Rotate != 0 {
		img = rotateImage(img, options.Rotate)
	}

	if options.FlipH {
		img = flipHImage(img)
	}

	if options.FlipV {
		img = flipVImage(img)
	}

	if options.Blur > 0 {
		img = blurGaussianImage(img, options.Blur)
	}

	if options.Format != "" {
		if buf, err := convertImageFormat(img, options.Format); err == nil {
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer

	if itype == "" {
		itype = imageType(srcData)
	}

	switch itype {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "bmp":
		err = bmp.Encode(&buf, img)
	case "tiff":
		err = tiff.Encode(&buf, img, nil)
	case "webp":
		err = webp.Encode(&buf, img, nil)
	default:
		return srcData, nil
	}
	if err == nil {
		return buf.Bytes(), nil
	}
	return srcData, nil
}

func (t *Image) selectFilter() imaging.ResampleFilter {
	switch t.ResizeFilter {
	case NearestNeighbor:
		return imaging.NearestNeighbor
	case Box:
		return imaging.Box
	case Linear:
		return imaging.Linear
	case Hermite:
		return imaging.Hermite
	case MitchellNetravali:
		return imaging.MitchellNetravali
	case CatmullRom:
		return imaging.CatmullRom
	case BSpline:
		return imaging.BSpline
	case Gaussian:
		return imaging.Gaussian
	case Bartlett:
		return imaging.Bartlett
	case Lanczos:
		return imaging.Lanczos
	case Hann:
		return imaging.Hann
	case Hamming:
		return imaging.Hamming
	case Blackman:
		return imaging.Blackman
	case Welch:
		return imaging.Welch
	case Cosine:
		return imaging.Cosine
	default:
		return imaging.MitchellNetravali
	}
}

func (t *Image) Resize(srcData []byte, width, height int, mode Mode) (destData []byte, err error) {
	if width == 0 && height == 0 {
		return srcData, nil
	}
	return t.Encode(srcData, width, height, mode, nil)
}

func imageType(srcData []byte) (s string) {
	if len(srcData) < 8 {
		return
	}
	switch {
	case bytes.Equal(srcData[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png"
	case bytes.Equal(srcData[:2], []byte{0x42, 0x4D}):
		return "bmp"
	case bytes.Equal(srcData[:2], []byte{0xFF, 0xD8}):
		return "jpeg"
	case bytes.Equal(srcData[:6], []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}) || bytes.Equal(srcData[:6], []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}):
		return "gif"
	case bytes.Equal(srcData[:4], []byte{0x49, 0x49, 0x2A, 0x00}) || bytes.Equal(srcData[:4], []byte{0x4D, 0x4D, 0x00, 0x2A}):
		return "tiff"
	case bytes.Equal(srcData[:4], []byte{0x52, 0x49, 0x46, 0x46}):
		return "webp"
	case bytes.Equal(srcData[:4], []byte{0x38, 0x42, 0x50, 0x53}):
		return "psd"
	case bytes.Equal(srcData[:4], []byte{0x00, 0x00, 0x01, 0x00}):
		return "ico"
	case bytes.Equal(srcData[:8], []byte{0x00, 0x00, 0x00, 0x0C, 0x61, 0x76, 0x69, 0x66}):
		return "avif"
	}
	return
}

func praseMode(mode Mode, w, h, width, height int) (nw, nh int, resizeType ResizeType) {
	if width > w && height > h {
		return w, h, SCALE
	}
	switch mode {
	case Mode0:
		nw, nh = getMin(w, h, width, height, false)
		resizeType = SCALE
	case Mode1:
		nw, nh = getMax(w, h, width, height, true)
		resizeType = THUMBNAIL
	case Mode2:
		nw, nh = getMin(w, h, width, height, false)
		resizeType = SCALE
	case Mode3:
		nw, nh = getMax(w, h, width, height, false)
		resizeType = SCALE
	case Mode4:
		nw, nh = getMax(w, h, width, height, false)
		resizeType = SCALE
	case Mode5:
		nw, nh = getMin(w, h, width, height, true)
		resizeType = THUMBNAIL
	default:
		return w, h, SCALE
	}
	return
}

func getMin(w, h, width, height int, isThubnail bool) (nw, nh int) {
	if width > w && height > h || (width == 0 && height == 0) {
		return w, h
	}
	if isThubnail {
		nw, nh = w, h
		if width < w {
			nw = width
		}
		if height < h {
			nh = height
		}
		if nw == 0 {
			nw = nh
		}
		if nh == 0 {
			nh = nw
		}
		return
	}
	if float32(width)/float32(w) > float32(height)/float32(h) {
		if height > 0 {
			return 0, height
		} else if width < w {
			return width, 0
		}
	} else {
		if width > 0 {
			return width, 0
		} else if height < h {
			return 0, height
		}
	}
	return w, h
}

func getMax(w, h, width, height int, isThubnail bool) (nw, nh int) {
	if width > w && height > h {
		return w, h
	}
	if isThubnail {
		nw, nh = w, h
		if width < w {
			nw = width
		}
		if height < h {
			nh = height
		}
		if nw == 0 {
			nw = nh
		}
		if nh == 0 {
			nh = nw
		}
		return
	}
	if float32(width)/float32(w) > float32(height)/float32(h) {
		if width < w {
			return width, 0
		} else if height > 0 {
			return 0, height
		}
	} else {
		if height < h {
			return 0, height
		} else if width > 0 {
			return width, 0
		}
	}
	return w, h
}

func QualityByBinary(srcData []byte, quality int) (_r []byte, err error) {
	if quality > 10 {
		quality = quality%10 + 1
	}
	img, _, er := image.Decode(bytes.NewReader(srcData))
	if er != nil {
		return nil, er
	}

	if _r, err = Quality(img, imageType(srcData), quality); err != nil || _r == nil {
		_r = srcData
	}
	return
}

func Quality(img image.Image, imagetype string, quality int) (_r []byte, err error) {
	if quality > 10 {
		quality = quality%10 + 1
	}
	buf := buffer.NewBuffer()
	switch imagetype {
	case "jpeg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: int(float64(quality) * 7.5)})
	case "png":
		level := png.BestCompression
		if quality >= 8 {
			level = png.BestSpeed
		} else if quality >= 4 {
			level = png.DefaultCompression
		}
		options := &png.Encoder{
			CompressionLevel: level,
		}
		err = options.Encode(buf, img)
	case "gif":
		err = gif.Encode(buf, img, &gif.Options{NumColors: int(float64(quality) * 25.6)})
	case "tiff":
		err = tiff.Encode(buf, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
	case "webp":
		err = webp.Encode(buf, img, &webp.Options{Quality: float32(quality * 9)})
	}
	if err == nil && buf.Len() > 0 {
		return buf.Bytes(), nil
	} else {
		return nil, err
	}
}

func convertImageFormat(img image.Image, format string) (buff bytes.Buffer, err error) {
	switch format {
	case "jpg":
		fallthrough
	case "jpeg":
		err = imaging.Encode(&buff, img, imaging.JPEG)
	case "png":
		err = imaging.Encode(&buff, img, imaging.PNG)
	case "gif":
		err = imaging.Encode(&buff, img, imaging.GIF)
	case "bmp":
		err = imaging.Encode(&buff, img, imaging.BMP)
	case "tif":
		fallthrough
	case "tiff":
		err = imaging.Encode(&buff, img, imaging.TIFF)
	case "webp":
		err = webp.Encode(&buff, img, &webp.Options{Lossless: true})
	case "ico":
		w := img.Bounds().Dx()
		h := img.Bounds().Dy()
		sizes := []int{16, 32, 48, 64, 128, w, h}
		sort.Ints(sizes)
		i := sort.SearchInts(sizes, w)
		j := sort.SearchInts(sizes, h)
		k := i
		if k > j {
			k = j
		}
		tb := [][2]uint8{}
		for i := 0; i < k; i++ {
			tb = append(tb, [][2]uint8{{uint8(sizes[i]), uint8(sizes[i])}}...)
		}
		err = ico.Encode(&buff, img, &ico.Options{Thumbnails: tb})
	default:
		return buff, fmt.Errorf("unsupported image format: %s", format)
	}
	return
}

func rotateImage(img image.Image, degrees int) image.Image {
	return imaging.Rotate(img, float64(degrees), color.Transparent)
}

func flipHImage(img image.Image) image.Image {
	return imaging.FlipH(img)
}

func flipVImage(img image.Image) image.Image {
	return imaging.FlipV(img)
}
