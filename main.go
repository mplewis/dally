package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	white = color.RGBA{255, 255, 255, 255}
	black = color.RGBA{0, 0, 0, 255}
)

func noise(w int, h int) image.Image {
	i := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			color := white
			if rand.Float32() < 0.5 {
				color = black
			}
			i.Set(x, y, color)
		}
	}
	return i
}

func rgbaToColor(r uint32, g uint32, b uint32, _a uint32) colorful.Color {
	return colorful.Color{R: float64(r/257) / 255, G: float64(g/257) / 255, B: float64(b/257) / 255}
}

func dist(a image.Image, b image.Image) (float64, error) {
	wa, ha, psa, err := getPixels(a)
	if err != nil {
		return 0, err
	}
	wb, hb, psb, err := getPixels(b)
	if err != nil {
		return 0, err
	}
	if wa != wb || ha != hb {
		return 0, fmt.Errorf("Dimension mismatch: %dx%d, %dx%d", wa, ha, wb, hb)
	}

	dist := 0.0
	for i, pa := range psa {
		pb := psb[i]
		dist += pa.DistanceLab(pb)
	}

	return dist, err
}

// Return all pixels in a single array
func getPixels(img image.Image) (int, int, []colorful.Color, error) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var pixels []colorful.Color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels = append(pixels, rgbaToColor(img.At(x, y).RGBA()))
		}
	}
	return width, height, pixels, nil
}

func save(i image.Image, fn string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	return png.Encode(f, i)
}

// Combine base and noise into a mutation, then evaluate it against the goal image
func mutateAndEval(goal image.Image, base image.Image, chg float64) (image.Image, float64, error) {
	b := goal.Bounds().Max
	w, h := b.X, b.Y
	mutated := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if rand.Float64() > chg {
				mutated.Set(x, y, base.At(x, y))
				continue
			}

			if base.At(x, y) == black {
				mutated.Set(x, y, white)
			} else {
				mutated.Set(x, y, black)
			}
		}
	}
	bl := blur.Gaussian(mutated, 2.0)
	blsh := effect.UnsharpMask(bl, 2.0, 0.5)
	d, err := dist(blsh, goal)
	return mutated, d, err
}

func main() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	f, err := os.Open("bc.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	orig, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	cand := noise(orig.Bounds().Max.X, orig.Bounds().Max.Y)
	candDist, err := dist(cand, orig)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000000; i++ {
		mut := rand.Float64() * 0.05
		newCand, newDist, err := mutateAndEval(orig, cand, mut)
		if err != nil {
			log.Fatal(err)
		}

		if newDist < candDist {
			fmt.Printf("Promoted at gen %d: (mut: %f) %f -> %f\n", i, mut, candDist, newDist)
			cand = newCand
			candDist = newDist
			save(cand, fmt.Sprintf("zz_cand_%d.png", i))
		}
	}

	fmt.Printf("Final distance: %f\n", candDist)
	save(cand, "cand.png")

	// bl := blur.Gaussian(orig, 2.0)
	// save(bl, "blur.png")

	// sh := effect.UnsharpMask(orig, 2.0, 0.5)
	// save(sh, "sharp.png")

	// blsh := effect.UnsharpMask(bl, 2.0, 0.5)
	// save(blsh, "blsh.png")

	// n := noise(w, h)
	// save(n, "noise.png")
}
