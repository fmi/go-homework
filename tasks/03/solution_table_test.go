package main

import (
	"testing"

	"github.com/fmi/go-homework/geom"
)

func TestTriangle(t *testing.T) {
	tests := []struct {
		description string
		triangle    geom.Intersectable
		ray         geom.Ray
		intersected bool
	}{
		{
			description: "simple intersection",
			triangle: NewTriangle(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(0, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1)),
			intersected: true,
		},
		{
			description: "no back face culling",
			triangle: NewTriangle(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(0, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, -1)),
			intersected: true,
		},
		{
			description: "ray opposite direction",
			triangle: NewTriangle(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(0, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, 1)),
			intersected: false,
		},
		{
			description: "near miss",
			triangle: NewTriangle(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(0, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(10, 10, -1)),
			intersected: false,
		},
		{
			description: "non axis aligned triangle",
			triangle: NewTriangle(
				geom.NewVector(1, 0, 0),
				geom.NewVector(0, 1, 0),
				geom.NewVector(0, 0, 1),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1)),
			intersected: true,
		},
		{
			description: "ray on edge",
			triangle: NewTriangle(
				geom.NewVector(1, 0, 0),
				geom.NewVector(0, 1, 0),
				geom.NewVector(0, 0, 1),
			),
			ray:         geom.NewRay(geom.NewVector(1, 0, 0), geom.NewVector(0, 1, 0)),
			intersected: true,
		},
		{
			description: "origin really close to object",
			triangle: NewTriangle(
				geom.NewVector(-1, 0, 0),
				geom.NewVector(1, 0, 0),
				geom.NewVector(0, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0.5, 1e-6), geom.NewVector(0, 0, 1)),
			intersected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.triangle.Intersect(test.ray)
			if actual != test.intersected {
				t.Errorf(
					"Expected intersection to be %t but it was not",
					test.intersected,
				)
			}
		})
	}
}

func TestQuad(t *testing.T) {
	tests := []struct {
		description string
		quad        geom.Intersectable
		ray         geom.Ray
		intersected bool
	}{
		{
			description: "simple intersection",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(1, 1, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1)),
			intersected: true,
		},
		{
			description: "no back face culling",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(1, 1, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, -1)),
			intersected: true,
		},
		{
			description: "ray opposite direction",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(1, 1, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, 1)),
			intersected: false,
		},
		{
			description: "near miss",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(1, 1, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(10, 10, -1)),
			intersected: false,
		},
		{
			description: "non axis aligned quad",
			quad: NewQuad(
				geom.NewVector(1, 0, 0),
				geom.NewVector(0.5946035575013605, 0.5946035575013605, 0),
				geom.NewVector(0, 1, 0),
				geom.NewVector(0, 0, 1),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1)),
			intersected: true,
		},
		{
			description: "ray on edge",
			quad: NewQuad(
				geom.NewVector(1, 0, 0),
				geom.NewVector(0.5946035575013605, 0.5946035575013605, 0),
				geom.NewVector(0, 1, 0),
				geom.NewVector(0, 0, 1),
			),
			ray:         geom.NewRay(geom.NewVector(1, 0, 0), geom.NewVector(0, 1, 0)),
			intersected: true,
		},
		{
			description: "origin really close to object",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(1, 1, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 1e-6), geom.NewVector(0, 0, 1)),
			intersected: false,
		},
		{
			description: "irregular quad hit",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(10, 10, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1)),
			intersected: true,
		},
		{
			description: "second irregular quad hit",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(0.2, 0.2, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1)),
			intersected: true,
		},
		{
			description: "irregular quad miss",
			quad: NewQuad(
				geom.NewVector(-1, -1, 0),
				geom.NewVector(1, -1, 0),
				geom.NewVector(10, 10, 0),
				geom.NewVector(-1, 1, 0),
			),
			ray:         geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(-10, 10, -1)),
			intersected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.quad.Intersect(test.ray)
			if actual != test.intersected {
				t.Errorf(
					"Expected intersection to be %t but it was not",
					test.intersected,
				)
			}
		})
	}
}

func TestSphere(t *testing.T) {
	tests := []struct {
		description string
		sphere      geom.Intersectable
		ray         geom.Ray
		intersected bool
	}{
		{
			description: "simple intersection",
			sphere:      NewSphere(geom.NewVector(5, 5, 5), 3),
			ray:         geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1)),
			intersected: true,
		},
		{
			description: "no back face culling",
			sphere:      NewSphere(geom.NewVector(5, 5, 5), 3),
			ray:         geom.NewRay(geom.NewVector(5, 5, 5), geom.NewVector(3, 2, 1)),
			intersected: true,
		},
		{
			description: "ray opposite direction",
			sphere:      NewSphere(geom.NewVector(5, 5, 5), 3),
			ray:         geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(-1, -1, -1)),
			intersected: false,
		},
		{
			description: "near miss",
			sphere:      NewSphere(geom.NewVector(5, 5, 5), 1),
			ray:         geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 5)),
			intersected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := test.sphere.Intersect(test.ray)
			if actual != test.intersected {
				t.Errorf(
					"Expected intersection to be %t but it was not",
					test.intersected,
				)
			}
		})
	}
}
