package main

import (
	"crypto/md5"
	"image/color"
	"image/png"
	"net/http"
	"regexp"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func handleIcon(w http.ResponseWriter, r *http.Request) {
	stop := r.URL.Query().Get("stop")
	if stop == "" {
		http.Error(w, "stop parameter missing", http.StatusBadRequest)
		return
	}

	matched, err := regexp.MatchString(`\d\d\d\d\d`, stop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !matched {
		http.Error(w, "not 5 digits", http.StatusBadRequest)
		return
	}

	hash := md5.Sum([]byte(stop))
	bgColor := color.RGBA{R: hash[0], G: hash[1], B: hash[2], A: 255}

	const S = 200
	maxWidth := float64(S) - 20
	dc := gg.NewContext(S, S)
	dc.SetColor(bgColor)
	dc.Clear()
	dc.SetRGB(1, 1, 1)

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	face := truetype.NewFace(font, &truetype.Options{Size: 48})
	dc.SetFontFace(face)

	dc.DrawStringWrapped(stop, S/2, S/2, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	w.Header().Set("Content-Type", "image/png")
	err = png.Encode(w, dc.Image())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
