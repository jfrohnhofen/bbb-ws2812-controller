package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

func LoadImage(file string) (*image.RGBA, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
	return rgba, nil
}

func GetPixel(img *image.RGBA, x, y float64) color.RGBA {
	xIntF, xFrac := math.Modf(x)
	yIntF, yFrac := math.Modf(y)
	xInt := int(xIntF)
	yInt := int(yIntF)

	x0 := img.RGBAAt(xInt, yInt)
	x1 := img.RGBAAt(xInt+1, yInt)
	x2 := img.RGBAAt(xInt, yInt+1)
	x3 := img.RGBAAt(xInt+1, yInt+1)

	return color.RGBA{
		R: gammaCorrection(interploate(x0.R, x1.R, x2.R, x3.R, xFrac, yFrac)),
		G: gammaCorrection(interploate(x0.G, x1.G, x2.G, x3.G, xFrac, yFrac)),
		B: gammaCorrection(interploate(x0.B, x1.B, x2.B, x3.B, xFrac, yFrac)),
	}
}

func gammaCorrection(x uint8) uint8 {
	return []uint8{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2,
		2, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 5, 5, 5,
		5, 6, 6, 6, 6, 7, 7, 7, 7, 8, 8, 8, 9, 9, 9, 10,
		10, 10, 11, 11, 11, 12, 12, 13, 13, 13, 14, 14, 15, 15, 16, 16,
		17, 17, 18, 18, 19, 19, 20, 20, 21, 21, 22, 22, 23, 24, 24, 25,
		25, 26, 27, 27, 28, 29, 29, 30, 31, 32, 32, 33, 34, 35, 35, 36,
		37, 38, 39, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 50,
		51, 52, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 66, 67, 68,
		69, 70, 72, 73, 74, 75, 77, 78, 79, 81, 82, 83, 85, 86, 87, 89,
		90, 92, 93, 95, 96, 98, 99, 101, 102, 104, 105, 107, 109, 110, 112, 114,
		115, 117, 119, 120, 122, 124, 126, 127, 129, 131, 133, 135, 137, 138, 140, 142,
		144, 146, 148, 150, 152, 154, 156, 158, 160, 162, 164, 167, 169, 171, 173, 175,
		177, 180, 182, 184, 186, 189, 191, 193, 196, 198, 200, 203, 205, 208, 210, 213,
		215, 218, 220, 223, 225, 228, 231, 233, 236, 239, 241, 244, 247, 249, 252, 255,
	}[x]
}

func interploate(x0, x1, x2, x3 uint8, xFrac, yFrac float64) uint8 {
	x := float64(x0)*(1-xFrac)*(1-yFrac) +
		float64(x1)*xFrac*(1-yFrac) +
		float64(x2)*(1-xFrac)*yFrac +
		float64(x3)*xFrac*yFrac
	return uint8(math.Round(x))
}
