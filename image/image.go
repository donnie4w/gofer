// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/image

package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"sort"
	"strings"

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

// Options holds configuration for image processing operations
type Options struct {
	Gray       bool    // Convert to grayscale
	Invert     bool    // Invert colors
	Format     string  // Target format (jpg, png, gif, etc.)
	Rotate     int     // Rotation angle in degrees
	FlipH      bool    // Flip horizontally
	FlipV      bool    // Flip vertically
	Colors     int     // Not currently used
	Quality    int     // Compression quality (1-10)
	CropAnchor []int   // Crop by anchor [x, y, width, height]
	CropSide   []int   // Crop by side [x, y, width, height]
	Blur       float64 // Gaussian blur sigma
	ScaleUpper []int   // Scale to maximum size with ratio
	ScaleLower []int   // Scale to minimum size with ratio
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

// Image provides image processing capabilities with configurable resize filter
type Image struct {
	ResizeFilter ResampleFilter
}

// ResizeGIF resizes an animated GIF while attempting to preserve animation quality.
// Note: GIF resizing often introduces noise due to 256-color limitation.
func ResizeGIF(srcData []byte, targetWidth, targetHeight int) ([]byte, error) {
	gifImg, err := gif.DecodeAll(bytes.NewReader(srcData))
	if err != nil {
		return srcData, err
	}
	resizedGif, err := resizeGIF(gifImg, targetWidth, targetHeight)
	if err != nil {
		return srcData, err
	}

	var buf bytes.Buffer
	gif.EncodeAll(&buf, resizedGif)

	return buf.Bytes(), nil
}

// resizeGIF is the core implementation for resizing GIF animations.
// Note: Uses Blur + FloydSteinberg dithering which may produce visible noise on downscaling.
func resizeGIF(g *gif.GIF, targetWidth, targetHeight int) (*gif.GIF, error) {
	if g == nil || len(g.Image) == 0 {
		return nil, nil
	}

	// Calculate the overall canvas size to handle frames with different bounds
	canvasRect := image.Rect(0, 0, 0, 0)
	for _, frame := range g.Image {
		canvasRect = canvasRect.Union(frame.Bounds())
	}
	canvasW := canvasRect.Dx()
	canvasH := canvasRect.Dy()

	// Calculate scale factor while maintaining aspect ratio
	var scale float64 = 1.0
	if targetWidth > 0 || targetHeight > 0 {
		scaleX := math.Inf(1)
		scaleY := math.Inf(1)
		if targetWidth > 0 {
			scaleX = float64(targetWidth) / float64(canvasW)
		}
		if targetHeight > 0 {
			scaleY = float64(targetHeight) / float64(canvasH)
		}
		scale = math.Min(scaleX, scaleY)
		if scale > 1.0 {
			scale = 1.0 // prevent upscaling
		}
	}

	newW := int(float64(canvasW) * scale)
	newH := int(float64(canvasH) * scale)

	newGIF := &gif.GIF{
		Delay:     g.Delay,
		Disposal:  g.Disposal,
		LoopCount: g.LoopCount,
		Config: image.Config{
			ColorModel: g.Image[0].ColorModel(),
			Width:      newW,
			Height:     newH,
		},
	}

	for _, frame := range g.Image {
		// Create full-size RGBA canvas to handle frame offsets
		fullCanvas := image.NewRGBA(canvasRect)
		draw.Draw(fullCanvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Src)

		// Resize with blur to reduce aliasing (may introduce some noise)
		scaled := imaging.Resize(imaging.Blur(fullCanvas, 0.5), newW, newH, imaging.Lanczos)

		// Convert back to paletted image using FloydSteinberg dithering
		pal := image.NewPaletted(scaled.Bounds(), frame.Palette)
		draw.FloydSteinberg.Draw(pal, pal.Rect, scaled, image.Point{})

		newGIF.Image = append(newGIF.Image, pal)
	}

	return newGIF, nil
}

// Encode processes an image with resizing, cropping, effects, and format conversion.
func (t *Image) Encode(srcData []byte, width, height int, mode Mode, options *Options) (destData []byte, err error) {
	defer func() {
		if er := recover(); er != nil {
			err = errors.New(fmt.Sprint(er))
		}
	}()

	img, itype, decodeErr := image.Decode(bytes.NewReader(srcData))
	if decodeErr != nil {
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

	if err == nil && buf.Len() > 0 {
		return buf.Bytes(), nil
	}
	return srcData, err
}

// selectFilter returns the corresponding imaging resample filter.
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

// Resize is a convenience method to resize an image using the specified mode.
func (t *Image) Resize(srcData []byte, width, height int, mode Mode) (destData []byte, err error) {
	if width == 0 && height == 0 {
		return srcData, nil
	}
	return t.Encode(srcData, width, height, mode, nil)
}

func imageType(data []byte) string {
	l := len(data)
	if l == 0 {
		return ""
	}

	// ---------- Binary magic first ----------

	// PNG
	if l >= 8 && bytes.Equal(data[:8], []byte{
		0x89, 0x50, 0x4E, 0x47,
		0x0D, 0x0A, 0x1A, 0x0A,
	}) {
		return "png"
	}

	// JPEG
	if l >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return "jpeg"
	}

	// GIF
	if l >= 6 && (bytes.Equal(data[:6], []byte("GIF89a")) ||
		bytes.Equal(data[:6], []byte("GIF87a"))) {
		return "gif"
	}

	// BMP
	if l >= 2 && bytes.Equal(data[:2], []byte("BM")) {
		return "bmp"
	}

	// TIFF
	if l >= 4 && (bytes.Equal(data[:4], []byte{0x49, 0x49, 0x2A, 0x00}) ||
		bytes.Equal(data[:4], []byte{0x4D, 0x4D, 0x00, 0x2A})) {
		return "tiff"
	}

	// WEBP: RIFF + WEBP
	if l >= 12 &&
		bytes.Equal(data[:4], []byte("RIFF")) &&
		bytes.Equal(data[8:12], []byte("WEBP")) {
		return "webp"
	}

	// ICO
	if l >= 4 && (bytes.Equal(data[:4], []byte{0x00, 0x00, 0x01, 0x00}) ||
		bytes.Equal(data[:4], []byte{0x00, 0x00, 0x02, 0x00})) {
		return "ico"
	}

	// PSD
	if l >= 4 && bytes.Equal(data[:4], []byte("8BPS")) {
		return "psd"
	}

	// ---------- ISO BMFF based images ----------

	if l >= 12 && bytes.Equal(data[4:8], []byte("ftyp")) {
		brand := string(data[8:12])

		switch brand {
		case "avif", "avis":
			return "avif"
		case "heic", "heix", "hevc", "hevx":
			return "heic"
		case "jxl ":
			return "jxl"
		case "jp2 ":
			return "jp2"
		}
	}

	// ---------- JPEG XL codestream ----------
	if l >= 2 && data[0] == 0xFF && data[1] == 0x0A {
		return "jxl"
	}

	// ---------- TGA ----------
	if l >= 18 {
		imgType := data[2]
		if imgType == 2 || imgType == 3 || imgType == 10 || imgType == 11 {
			pixelDepth := data[16]
			if pixelDepth == 8 || pixelDepth == 16 || pixelDepth == 24 || pixelDepth == 32 {
				return "tga"
			}
		}
	}

	// ---------- SVG ----------
	if l >= 5 {
		n := l
		if n > 256 {
			n = 256
		}
		head := strings.TrimSpace(strings.ToLower(string(data[:n])))

		if strings.HasPrefix(head, "<svg") ||
			(strings.HasPrefix(head, "<?xml") && strings.Contains(head, "<svg")) {
			return "svg"
		}
	}

	return ""
}

func praseMode(mode Mode, w, h, preWidth, preHeight int) (nw, nh int, resizeType ResizeType) {
	width, height := newSide4mode(w, h, preWidth, preHeight)
	if width > w && height > h {
		return w, h, SCALE
	}
	switch mode {
	case Mode0:
		nw, nh = getMinAndkeepRatio(w, h, width, height)
		resizeType = SCALE
	case Mode1:
		nw, nh = getSideByThubnail(w, h, preWidth, preHeight)
		resizeType = THUMBNAIL
	case Mode2:
		if preWidth == 0 {
			preWidth = int(float64(preHeight) * float64(w) / float64(h))
			nw, nh = getMaxAndkeepRatio(w, h, preWidth, preHeight)
		} else if preHeight == 0 {
			preHeight = int(float64(preWidth) / float64(w) * float64(h))
			nw, nh = getMaxAndkeepRatio(w, h, preWidth, preHeight)
		} else {
			nw, nh = getMaxAndkeepRatio(w, h, width, height)
		}
		resizeType = SCALE
	case Mode3:
		if preWidth == 0 {
			preWidth = preHeight
		} else if preHeight == 0 {
			preHeight = preWidth
		}
		nw, nh = getMaxAndkeepRatio(w, h, preWidth, preHeight)
		resizeType = SCALE
	case Mode4:
		if width == 0 {
			width = height
		} else if height == 0 {
			height = width
		}
		nw, nh = getMaxAndkeepRatio(w, h, width, height)
		resizeType = SCALE
	case Mode5:
		nw, nh = getSideByThubnail(w, h, width, height)
		resizeType = THUMBNAIL
	default:
		return w, h, SCALE
	}
	return
}

func newSide4mode(w, h, width, height int) (int, int) {
	if w >= h {
		return width, height
	} else {
		return height, width
	}
}

func getMinAndkeepRatio(w, h, newwidth, newheight int) (int, int) {
	if newwidth == 0 {
		newwidth = w
	}
	if newheight == 0 {
		newheight = h
	}
	if newwidth >= w && newheight >= h {
		return w, h
	}
	r := float64(w) / float64(h)
	if newwidth < w && newheight < h {
		if newwidth > newheight {
			newwidth = int(float64(newheight) * r)
		} else {
			newheight = int(float64(newwidth) / r)
		}
		return newwidth, newheight
	}
	if newwidth >= w {
		newwidth = int(float64(newheight) * r)
	} else {
		newheight = int(float64(newwidth) / r)
	}
	return newwidth, newheight
}

func getMaxAndkeepRatio(w, h, newwidth, newheight int) (int, int) {
	if newwidth == 0 {
		newwidth = w
	}
	if newheight == 0 {
		newheight = h
	}
	if newwidth >= w || newheight >= h {
		return w, h
	}
	if newwidth < w && newheight < h {
		r := float64(w) / float64(h)
		if float64(newwidth)/float64(w) > float64(newheight)/float64(h) {
			newheight = int(float64(newwidth) / r)
		} else {
			newwidth = int(float64(newheight) * r)
		}
		return newwidth, newheight
	}
	return newwidth, newheight
}

func getSideByThubnail(w, h, width, height int) (nw, nh int) {
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

// QualityByBinary compresses an image based on a quality level from 1 to 10.
// It first decodes the binary data, then applies format-specific compression.
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

// Quality applies compression/optimization to an already decoded image.Image.
// The quality parameter is normalized to 1-10 and mapped differently per format.
func Quality(img image.Image, imagetype string, quality int) (_r []byte, err error) {
	if quality > 10 {
		quality = quality%10 + 1
	}
	buf := buffer.NewBuffer()
	switch imagetype {
	case "jpeg": // Map 1-10 to JPEG quality (roughly 10-75)
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: int(float64(quality) * 7.5)})
	case "png": // Choose PNG compression level based on quality
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

// ConvertImage converts an image from one format to another.
// If the source format is already the same as the target, it returns the original data.
func ConvertImage(bs []byte, format string) ([]byte, error) {
	img, itype, decodeErr := image.Decode(bytes.NewReader(bs))
	if decodeErr != nil {
		return nil, decodeErr
	}
	if itype == format {
		return bs, nil
	}
	buf, err := convertImageFormat(img, format)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func convertImageFormat(img image.Image, format string) (buff bytes.Buffer, err error) {
	switch format {
	case "jpg", "jpeg":
		err = imaging.Encode(&buff, img, imaging.JPEG)
	case "png":
		err = imaging.Encode(&buff, img, imaging.PNG)
	case "gif":
		err = imaging.Encode(&buff, img, imaging.GIF)
	case "bmp":
		err = imaging.Encode(&buff, img, imaging.BMP)
	case "tif", "tiff":
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
		var tb [][2]uint8
		if k > 0 {
			tb = [][2]uint8{}
			for i := 0; i < k; i++ {
				tb = append(tb, [][2]uint8{{uint8(sizes[i]), uint8(sizes[i])}}...)
			}
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
