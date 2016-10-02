package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/golang/freetype"
)

const (
	fontFile = "./segoepr.ttf"
	fontSize = 9
	fontDPI  = 401
)

var (
	cyan color.Color = color.RGBA{22, 153, 226, 255}
)

func index(w http.ResponseWriter, r *http.Request) {

	// fmt.Println("GET params were:", r.URL.Query())
	stop := r.URL.Query().Get("stop")

	matched, err := regexp.MatchString(`\d\d\d\d\d`, stop)
	if !matched {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	img := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		panic(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return
	}
	left := img.Bounds()
	left.Max = image.Pt(200, 200)
	draw.Draw(img, left, &image.Uniform{cyan}, image.ZP, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(fontDPI)
	c.SetFont(font)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)
	pt := freetype.Pt(7, 82)
	_, err = c.DrawString(stop, pt)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

func main() {
	http.HandleFunc("/", index)
	http.ListenAndServe(":7777", nil)
}
