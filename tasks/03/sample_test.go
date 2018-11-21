package main

import (
	"testing"

	"github.com/fmi/go-homework/geom"
)

func TestSampleSimpleOperations(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect triangle %#v but it did not.", ray, prim)
	}
}

// Intersect the Triangle in the C dot test
func TestIntersectWithTriangleWithC(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(0, 1, -1), geom.NewVector(0, 0, 6))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect triangle %#v but it did not.", ray, prim)
	}
}

// Intersect with a triangle side
func TestIntersectWithTriangleSidе(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(0, -1, -1), geom.NewVector(0, 0, 6))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect triangle %#v but it did not.", ray, prim)
	}
}

// Ray inside the triangle
func TestRayInsideTriangle(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(-2, -2, 0), geom.NewVector(0, 0, 0))

	if prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to be inside the triangle %#v to be parallel.", ray, prim)
	}
}

// Test when Ray is parallel with the triangle but not into the triangle
func TestRayParallelToTheTriangle(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(-2, -2, 1), geom.NewVector(0, 0, 1))

	if prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to be parallel to the triangle %#v.", ray, prim)
	}
}

//Sphere test should intersect the sphere
func TestRayIntersectSphere(t *testing.T) {

	center := geom.NewVector(0, 0, 0)
	radius := 2.0
	sphere := NewSphere(center, radius)
	ray := geom.NewRay(geom.NewVector(0, 0, -2), geom.NewVector(0, 0, 7))
	if !sphere.Intersect(ray) {
		t.Errorf("Expected the ray %#v to intersect the sphere %#v but it did not.", ray, sphere)
	}
}

func TestSampleIntersectableImplementations(t *testing.T) {
	var prim geom.Intersectable

	a, b, c, d := geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0, 1, 0),
		geom.NewVector(-1, 1, 0)

	prim = NewTriangle(a, b, c)
	prim = NewQuad(a, b, c, d)
	prim = NewSphere(a, 5)

	_ = prim
}
