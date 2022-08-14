package main

import (
	"crypto/md5"
	"fmt"
	"image/color"
	"image/png"
	"net/http"
	"regexp"

	"github.com/fogleman/gg"
)

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return
}

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

	bgColor, err := ParseHexColor(fmt.Sprintf("#%.3x", md5.Sum([]byte(stop))))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	const S = 200
	maxWidth := float64(S) - 20
	dc := gg.NewContext(S, S)
	// set bgColor as background color
	dc.SetColor(bgColor)
	dc.Clear()
	dc.SetRGB(1, 1, 1)
	if err := dc.LoadFontFace("static/Go-Regular.ttf", 64); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dc.DrawStringWrapped(stop, S/2, S/2, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	w.Header().Set("Content-Type", "image/png")
	err = png.Encode(w, dc.Image())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
