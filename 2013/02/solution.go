package main

import (
	"fmt"
)

type Color struct {
	Red   byte
	Green byte
	Blue  byte
}

func (c *Color) String() string {
	return fmt.Sprintf("Red: %s Green: %s Blue: %s", c.Red, c.Green, c.Blue)
}

type Pixel struct {
	red   float32
	green float32
	blue  float32
	alpha float32
}

func (p *Pixel) Color() Color {
	color := Color{byte(255 * p.red * p.alpha), byte(255 * p.green * p.alpha), byte(255 * p.blue * p.alpha)}
	return color
}

type Header struct {
	Format    string
	LineWidth int
	//	Encoding  int
}

type Image interface {
	InspectPixel(x int, y int) Pixel
}

type ImageRGB struct {
	data  []byte
	sizeX int
}

func (p *ImageRGB) InspectPixel(x int, y int) Pixel {
	red := float32(p.data[p.sizeX*3*y+x*3])
	green := float32(p.data[p.sizeX*3*y+x*3+1])
	blue := float32(p.data[p.sizeX*3*y+x*3+2])
	return Pixel{red / 255, green / 255, blue / 255, 1}
}

type ImageRGBA struct {
	data  []byte
	sizeX int
}

func (p *ImageRGBA) InspectPixel(x int, y int) Pixel {
	red := float32(p.data[p.sizeX*4*y+x*4])
	green := float32(p.data[p.sizeX*4*y+x*4+1])
	blue := float32(p.data[p.sizeX*4*y+x*4+2])
	alpha := float32(p.data[p.sizeX*4*y+x*4+3])
	return Pixel{red / 255, green / 255, blue / 255, alpha / 255}
}

func ParseImage(data []byte, header Header) Image {
	switch header.Format {
	case "RGB":
		image := new(ImageRGB)
		image.data = data
		image.sizeX = header.LineWidth
		return image
	case "RGBA":
		image := new(ImageRGBA)
		image.data = data
		image.sizeX = header.LineWidth
		return image
	case "BGRA":
		// Note: should be generalized for the "proper" solution
		image := new(ImageRGBA)

		for i := 0; i < len(data); i += 4 {
			r := data[i+2]
			g := data[i+1]
			b := data[i]
			a := data[i+3]

			image.data = append(image.data, r, g, b, a)
		}

		image.sizeX = header.LineWidth
		return image
	}
	return nil
}

func main() {
	data := []byte{0, 12, 244, 13, 26, 52, 31, 33, 41}
	header := Header{"RGB", 3}
	pixel := ParseImage(data, header).InspectPixel(0, 0)
	fmt.Print(pixel.Color())
}
