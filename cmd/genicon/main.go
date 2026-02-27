//go:build ignore

package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

func main() {
	uiSrc, err := loadPNG("icon.png")
	if err != nil {
		panic(err)
	}

	// Main UI icon: always sourced from icon.png.
	uiVisible := nonTransparentBounds(uiSrc)
	uiPath := filepath.Join("cmd", "iconsrefresh-tray", "icon.png")
	uiImg := scaleToSquare(uiSrc, uiVisible, 128)
	if err := writePNG(uiPath, uiImg); err != nil {
		panic(err)
	}
	fmt.Printf("icon.png %dx%d (cropped %dx%d) -> UI PNG 128x128 (%s)\n", uiSrc.Bounds().Dx(), uiSrc.Bounds().Dy(), uiVisible.Dx(), uiVisible.Dy(), uiPath)

	fmt.Println("tray icon asset written (resource icon uses default Windows icon)")
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func scaleToSquare(src image.Image, visible image.Rectangle, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	targetW := size
	targetH := size
	if vw, vh := visible.Dx(), visible.Dy(); vw > 0 && vh > 0 {
		if vw > vh {
			targetH = vh * size / vw
		} else if vh > vw {
			targetW = vw * size / vh
		}
	}
	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}
	dstRect := image.Rect((size-targetW)/2, (size-targetH)/2, (size-targetW)/2+targetW, (size-targetH)/2+targetH)
	draw.CatmullRom.Scale(dst, dstRect, src, visible, draw.Over, nil)
	return dst
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}

func nonTransparentBounds(img image.Image) image.Rectangle {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X-1, b.Min.Y-1

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if maxX < minX || maxY < minY {
		return b
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}
