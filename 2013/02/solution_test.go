package main

import (
	"errors"
	"fmt"
	"testing"
)

func assertColor(pixel Pixel, rgb ...byte) error {
	if pixel.Color().Red != rgb[0] {
		error_str := fmt.Sprintf("Wrong Red component: expected %d, got %d", rgb[0],
			pixel.Color().Red)
		return errors.New(error_str)
	}

	if pixel.Color().Green != rgb[1] {
		error_str := fmt.Sprintf("Wrong Green component: expected %d, got %d", rgb[1],
			pixel.Color().Green)
		return errors.New(error_str)
	}

	if pixel.Color().Blue != rgb[2] {
		error_str := fmt.Sprintf("Wrong Blue component: expected %d, got %d", rgb[2],
			pixel.Color().Blue)
		return errors.New(error_str)
	}

	return nil
}

func TestBasicRGBCall(t *testing.T) {
	data := []byte{
		0, 12, 244, 13, 26, 52, 31, 33, 41,
	}
	header := Header{"RGB", 3}
	picture := ParseImage(data, header)

	if err := assertColor(picture.InspectPixel(0, 0), 0, 12, 244); err != nil {
		t.Error(err)
	}
}

func TestBasicRGBACall(t *testing.T) {
	data := []byte{
		0, 12, 244, 128, 14, 26, 52, 127, 31, 33, 41, 255, 36, 133, 241, 255,
	}
	header := Header{"RGBA", 4}
	picture := ParseImage(data, header)

	first_pixel := picture.InspectPixel(0, 0)
	if err := assertColor(first_pixel, 0, 6, 122); err != nil {
		t.Error(err)
	}

	second_pixel := picture.InspectPixel(3, 0)
	if err := assertColor(second_pixel, 36, 133, 241); err != nil {
		t.Error(err)
	}
}

func TestBasicRGBARowsCall(t *testing.T) {
	data := []byte{
		0, 12, 244, 127, 14, 26, 52, 127,
		31, 33, 41, 255, 36, 133, 241, 255,
	}
	header := Header{"RGBA", 2}
	picture := ParseImage(data, header)

	pixel := picture.InspectPixel(1, 1)
	if err := assertColor(pixel, 36, 133, 241); err != nil {
		t.Error(err)
	}
}

func TestBasicBGRARowsCall(t *testing.T) {
	data := []byte{
		0, 12, 244, 127, 14, 26, 52, 127,
		31, 33, 41, 255, 36, 133, 241, 255,
	}
	header := Header{"BGRA", 2}
	picture := ParseImage(data, header)

	pixel := picture.InspectPixel(1, 1)
	if err := assertColor(pixel, 241, 133, 36); err != nil {
		t.Error(err)
	}
}
