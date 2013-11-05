package main

import (
	"errors"
	"fmt"
	"testing"
)

func assertColor(pixel Pixel, rgb ...byte) (error, byte) {
	if pixel.Color().Red != rgb[0] {
		error_str := fmt.Sprintf("Red colour component was supposed to be %d but it was",
			rgb[0])
		return errors.New(error_str), pixel.Color().Red
	}

	if pixel.Color().Green != rgb[1] {
		error_str := fmt.Sprintf("Red colour component was supposed to be %d but it was",
			rgb[1])
		return errors.New(error_str), pixel.Color().Green
	}

	if pixel.Color().Blue != rgb[2] {
		error_str := fmt.Sprintf("Red colour component was supposed to be %d but it was",
			rgb[2])
		return errors.New(error_str), pixel.Color().Blue
	}

	return nil, 0
}

func TestBasicRGBCall(t *testing.T) {
	data := []byte{
		0, 12, 244, 13, 26, 52, 31, 33, 41,
	}
	header := Header{"RGB", 3}
	picture := ParseImage(data, header)

	if err, value := assertColor(picture.InspectPixel(0, 0), 0, 12, 244); err != nil {
		t.Error(err, value)
	}
}

func TestBasicRGBACall(t *testing.T) {
	data := []byte{
		0, 12, 244, 128, 14, 26, 52, 127, 31, 33, 41, 255, 36, 133, 241, 255,
	}
	header := Header{"RGBA", 4}
	picture := ParseImage(data, header)

	first_pixel := picture.InspectPixel(0, 0)
	if err, value := assertColor(first_pixel, 0, 6, 122); err != nil {
		t.Error(err, value)
	}

	second_pixel := picture.InspectPixel(3, 0)
	if err, value := assertColor(second_pixel, 36, 133, 241); err != nil {
		t.Error(err, value)
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
	if err, value := assertColor(pixel, 36, 133, 241); err != nil {
		t.Error(err, value)
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
	if err, value := assertColor(pixel, 241, 133, 36); err != nil {
		t.Error(err, value)
	}
}
