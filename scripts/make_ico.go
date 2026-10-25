package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

type iconDirEntry struct {
	Width    byte
	Height   byte
	Palette  byte
	Reserved byte
	Planes   uint16
	Bits     uint16
	Size     uint32
	Offset   uint32
}

func resizeNearest(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := srcBounds.Min.X + (x*srcW)/w
			srcY := srcBounds.Min.Y + (y*srcH)/h
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func main() {
	srcPath := "icon.png"
	if len(os.Args) > 1 {
		srcPath = os.Args[1]
	}
	f, err := os.Open(srcPath)
	if err != nil {
		log.Fatalf("Open src (%s): %v", srcPath, err)
	}
	defer f.Close()

	srcImg, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Decode src: %v", err)
	}

	// 1. Save full PNG (512x512) for web UI
	web512 := image.NewRGBA(image.Rect(0, 0, 512, 512))
	// use resize
	resized512 := resizeNearest(srcImg, 512, 512)
	draw.Draw(web512, web512.Bounds(), resized512, image.Point{}, draw.Src)

	pngOut, err := os.Create("icon.png")
	if err == nil {
		_ = png.Encode(pngOut, web512)
		pngOut.Close()
	}

	favOut, err := os.Create(`cmd\fyrnox\web\favicon.png`)
	if err == nil {
		_ = png.Encode(favOut, web512)
		favOut.Close()
	}

	sizes := []int{256, 128, 64, 48, 32, 16}
	var pngBuffers [][]byte

	for _, sz := range sizes {
		resized := resizeNearest(srcImg, sz, sz)
		var buf bytes.Buffer
		if err := png.Encode(&buf, resized); err != nil {
			log.Fatalf("Encode png %d: %v", sz, err)
		}
		pngBuffers = append(pngBuffers, buf.Bytes())
	}

	// Build ICO
	icoFile, err := os.Create("icon.ico")
	if err != nil {
		log.Fatalf("Create ico: %v", err)
	}
	defer icoFile.Close()

	// Header: Reserved (0), Type (1), Count (len(sizes))
	_ = binary.Write(icoFile, binary.LittleEndian, uint16(0))
	_ = binary.Write(icoFile, binary.LittleEndian, uint16(1))
	_ = binary.Write(icoFile, binary.LittleEndian, uint16(len(sizes)))

	offset := uint32(6 + 16*len(sizes))

	for i, sz := range sizes {
		w := byte(sz)
		if sz >= 256 {
			w = 0
		}
		h := byte(sz)
		if sz >= 256 {
			h = 0
		}
		entry := iconDirEntry{
			Width:    w,
			Height:   h,
			Palette:  0,
			Reserved: 0,
			Planes:   1,
			Bits:     32,
			Size:     uint32(len(pngBuffers[i])),
			Offset:   offset,
		}
		_ = binary.Write(icoFile, binary.LittleEndian, entry)
		offset += uint32(len(pngBuffers[i]))
	}

	for _, buf := range pngBuffers {
		_, _ = icoFile.Write(buf)
	}

	// Copy icon.ico to installer payload
	icoBytes, err := os.ReadFile("icon.ico")
	if err == nil {
		_ = os.WriteFile(`cmd\installer\payload\icon.ico`, icoBytes, 0644)
	}

	log.Println("Successfully generated icon.png, favicon.png, and icon.ico!")
}
