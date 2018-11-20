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
