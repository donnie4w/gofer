package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

type ResizeType int
type Mode int8
type AdMode int8

const (
	SCALE ResizeType = iota
	THUMBNAIL
)

// size
const (
	Mode0 Mode = iota
	Mode1
	Mode2
	Mode3
	Mode4
	Mode5
)

// adjust
const (
	AdMode0 AdMode = iota //Original
	AdMode1               //Gray
	AdMode2               //Invert
)

func Resize(srcData []byte, width, height int, mode Mode, admode AdMode) (destData []byte, err error) {
	defer func() {
		if er := recover(); er != nil {
			err = errors.New(fmt.Sprint(er))
		}
	}()
	if width == 0 && height == 0 {
		return srcData, nil
	}
	img, _, er := image.Decode(bytes.NewReader(srcData))
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	nw, nh, resizeType := praseMode(mode, w, h, width, height)
	if er == nil {

		switch admode {
		case AdMode1:
			img = convertToGrayByImage(img)
		case AdMode2:
			img = invertByImage(img)
		}

		var nrgba *image.NRGBA
		switch resizeType {
		case SCALE:
			nrgba = imaging.Resize(img, nw, nh, imaging.Lanczos)
		case THUMBNAIL:
			nrgba = imaging.Fill(img, nw, nh, imaging.Center, imaging.Lanczos)
		}

		var buf bytes.Buffer
		switch imageType(srcData) {
		case "jpeg":
			err = jpeg.Encode(&buf, nrgba, nil)
		case "png":
			err = png.Encode(&buf, nrgba)
		case "gif":
			err = gif.Encode(&buf, nrgba, nil)
		case "bmp":
			err = bmp.Encode(&buf, nrgba)
		case "tiff":
			err = tiff.Encode(&buf, nrgba, nil)
		case "webp":
			err = webp.Encode(&buf, nrgba, nil)
		default:
			return srcData, nil
		}
		if err == nil {
			return buf.Bytes(), nil
		}
	}

	return srcData, nil
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
		nw, nh = getMax(w, h, width, height, true)
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
