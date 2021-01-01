package renderer

import (
	"bytes"
	"crypto/md5"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/moethu/imaging"
	libjpeg "github.com/pixiv/go-libjpeg/jpeg"
)

// onRender event handler for onRender event
func (app *RenderingApp) onRender(evname string, ev interface{}) {
	app.makeScreenShot()
}

var md5SumBuffer [16]byte

// makeScreenShot reads the opengl buffer, encodes it as jpeg and sends it to the channel
func (app *RenderingApp) makeScreenShot() {
	w := app.Width
	h := app.Height
	data := app.Gl().ReadPixels(0, 0, w, h, 6408, 5121)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Pix = data

	if app.imageSettings.getPixelation() > 1.0 {
		img = imaging.Fit(img, int(float64(w)/app.imageSettings.getPixelation()), int(float64(h)/app.imageSettings.getPixelation()), imaging.NearestNeighbor)
	}
	if app.imageSettings.brightness != 0 {
		img = imaging.AdjustBrightness(img, app.imageSettings.brightness)
	}
	if app.imageSettings.contrast != 0 {
		img = imaging.AdjustContrast(img, app.imageSettings.contrast)
	}
	if app.imageSettings.saturation != 0 {
		img = imaging.AdjustSaturation(img, app.imageSettings.saturation)
	}
	if app.imageSettings.blur != 0 {
		img = imaging.Blur(img, app.imageSettings.blur)
	}
	if app.imageSettings.invert {
		img = imaging.Invert(img)
	}

	img = imaging.FlipV(img)

	if app.Debug {
		img = DrawByteGraph(img)
	}

	buf := new(bytes.Buffer)
	var err interface{}
	switch app.imageSettings.encoder {
	case "png":
		err = png.Encode(buf, img)
	case "jpeg":
		var opt jpeg.Options
		opt.Quality = app.imageSettings.getJpegQuality()
		err = jpeg.Encode(buf, img, &opt)
	default:
		var opt libjpeg.EncoderOptions
		opt.Quality = app.imageSettings.getJpegQuality()
		err = libjpeg.Encode(buf, img, &opt)
	}

	if err != nil {
		panic(err)
	}
	imageBit := buf.Bytes()

	// get md5 checksum from image to check if image changed
	// only send a new image to the client if there has been any change.
	md := md5.Sum(imageBit)
	if md5SumBuffer != md {
		if app.Debug {
			AddToByteBuffer(len(imageBit))
		}
		app.cImagestream <- imageBit
		md5SumBuffer = md
	}
}
