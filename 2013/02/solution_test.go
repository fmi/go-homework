package main

import (
	"testing"
	"fmt"
)

func TestBasicRGBCall(t *testing.T) {
	data := byte[]{0, 12, 244, 13, 26, 52, 31, 33, 41}
	header := Header{"RGB", 3, 1, nil}

	picture := ParseImage(data, header)
	pixel := picture.InspectPixel(0, 0)

	if pixel.color.Red != 0 {
		t.Error("Red pixel whuld not be ", parsePath(path))
	}

	if pixel.color.Green != 12 {
		t.Error("Green pixel whuld not be ", parsePath(path))
	}

	if pixel.color.Blue != 244 {
		t.Error("Blue pixel whuld not be ", parsePath(path))
	}
}

func TestBasicRGBACall(t *testing.T) {
	data := byte[]{0, 12, 244, 127, 14, 26, 52, 127, 31, 33, 41, 255, 36, 133, 241, 255}
	header := Header{"RGB", 4, 1, nil}

	picture := ParseImage(data, header)
	first_pixel := picture.InspectPixel(0, 0)

	if first_pixel.color.Red != 0 {
		t.Error("Red pixel whuld not be ", parsePath(path))
	}

	if first_pixel.color.Green != 6 {
		t.Error("Green pixel whuld not be ", parsePath(path))
	}

	if first_pixel.color.Blue != 122 {
		t.Error("Blue pixel whuld not be ", parsePath(path))
	}

	second_pixel := picture.InspectPixel(3, 0)

	if second_pixel.color.Red != 36 {
		t.Error("Red pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Green != 133 {
		t.Error("Green pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Blue != 241 {
		t.Error("Blue pixel whuld not be ", parsePath(path))
	}
}

func TestBasicRGBARowsCall(t *testing.T) {
	data := byte[]{0, 12, 244, 127, 14, 26, 52, 127, 31, 33, 41, 255, 36, 133, 241, 255}
	header := Header{"RGBA", 2, 2, nil}

	picture := ParseImage(data, header)

	second_pixel := picture.InspectPixel(1, 1)

	if second_pixel.color.Red != 36 {
		t.Error("Red pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Green != 133 {
		t.Error("Green pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Blue != 241 {
		t.Error("Blue pixel whuld not be ", parsePath(path))
	}
}

func TestBasicBGRARowsCall(t *testing.T) {
	data := byte[]{0, 12, 244, 127, 14, 26, 52, 127, 31, 33, 41, 255, 36, 133, 241, 255}
	header := Header{"RGBA", 2, 2, nil}

	picture := ParseImage(data, header)

	second_pixel := picture.InspectPixel(1, 1)

	if second_pixel.color.Red != 241 {
		t.Error("Red pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Green != 133 {
		t.Error("Green pixel whuld not be ", parsePath(path))
	}

	if second_pixel.color.Blue != 36 {
		t.Error("Blue pixel whuld not be ", parsePath(path))
	}
}
