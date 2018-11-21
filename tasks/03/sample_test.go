package main

import (
	"testing"

	"github.com/fmi/go-homework/geom"
)

func TestSampleSimpleOperationsShouldIntersect(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(-1, -1, 0), geom.NewVector(1, -1, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect triangle %#v but it did not.", ray, prim)
	}
}

func TestSampleSimpleTriangleBoundryIntersectionShouldIntersect(t *testing.T) {
	var prim geom.Intersectable

	a, b, c := geom.NewVector(0, -2, 0), geom.NewVector(3, 0, 0), geom.NewVector(0, 1, 0)
	prim = NewTriangle(a, b, c)
	ray := geom.NewRay(geom.NewVector(0, 0, 3), geom.NewVector(0, 0, -2))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect triangle %#v but it did not.", ray, prim)
	}
}

func TestSampleSimpleSphereShouldIntersect(t *testing.T) {
	var prim geom.Intersectable

	origin, r := geom.NewVector(0, 0, 0), 2.0
	prim = NewSphere(origin, r)
	ray := geom.NewRay(geom.NewVector(0, 0, 2.5), geom.NewVector(0, 0, 1))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect sphere %#v but it did not.", ray, prim)
	}
}

func TestSampleSimpleQuadConvexShouldIntersect(t *testing.T) {
	var prim geom.Intersectable

	a, b, c, d := geom.NewVector(0, -2, 0), geom.NewVector(3, 0, 0), geom.NewVector(0, 1, 0), geom.NewVector(-1, 0, 0)
	prim = NewQuad(a, b, c, d)
	ray := geom.NewRay(geom.NewVector(0, 0, 2), geom.NewVector(0, 0, -1))

	if !prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to intersect quad %#v but it did not.", ray, prim)
	}
}

func TestSampleSimpleQuadConcaveShouldNotIntersect(t *testing.T) {
	var prim geom.Intersectable

	a, b, c, d := geom.NewVector(0, -2, 0), geom.NewVector(3, 0, 0), geom.NewVector(0, 1, 0), geom.NewVector(1, 0, 0)
	prim = NewQuad(a, b, c, d)
	ray := geom.NewRay(geom.NewVector(0, 0, 2), geom.NewVector(0, 0, -1))

	if prim.Intersect(ray) {
		t.Errorf("Expected ray %#v to not intersect quad %#v but it did.", ray, prim)
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
