package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"

	"github.com/lucasb-eyer/go-colorful"
)

func rgbaToColor(r uint32, g uint32, b uint32, _a uint32) colorful.Color {
	fmt.Println(r, g, b)
	return colorful.Color{R: float64(r/257) / 255, G: float64(g/257) / 255, B: float64(b/257) / 255}
}

// squared distance between two values that do not wrap on overflow
func diffN(a float64, b float64) float64 {
	return (a - b) * (a - b)
}

// squared distance between two values that DO wrap on overflow @ range
func diffNCirc(a float64, b float64, rang float64) float64 {
	x := math.Min(a, b)
	y1 := math.Max(a, b)
	y2 := y1 - rang
	return math.Min(diffN(x, y1), diffN(x, y2))
}

// squared distance between H, S, V values of two points
func diff(a colorful.Color, b colorful.Color) float64 {
	hA, cA, lA := a.Hcl()
	hB, cB, lB := b.Hcl()
	diffH := diffNCirc(hA, hB, 1.0)
	diffC := diffN(cA, cB)
	diffL := diffN(lA, lB)
	return diffH + diffC + diffL
}

// Return all pixels in a single array
func getPixels(file io.Reader) ([]colorful.Color, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels []colorful.Color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels = append(pixels, rgbaToColor(img.At(x, y).RGBA()))
		}
	}
	return pixels, nil
}

func main() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	file, err := os.Open("./image.png")
	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}
	defer file.Close()

	pixels, err := getPixels(file)
	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	white := colorful.Color{R: 1.0, G: 1.0, B: 1.0}
	black := colorful.Color{R: 0.0, G: 0.0, B: 0.0}
	wh, ws, wv := white.Hcl()
	fmt.Printf("%f, %f, %f\n", wh, ws, wv)
	bh, bs, bv := black.Hcl()
	fmt.Printf("%f, %f, %f\n", bh, bs, bv)

	for _, pixel := range pixels {
		fmt.Printf("%f, %f, %f\n", pixel.R, pixel.G, pixel.B)
		h, s, v := pixel.Hsv()
		fmt.Printf("%f, %f, %f, %f, %f\n", h, s, v, diff(pixel, white), diff(pixel, black))
	}

	fmt.Println(len(pixels))
}
