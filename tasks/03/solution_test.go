package main

import (
	"testing"

	"github.com/fmi/go-homework/geom"
)

func TestTriangleSimpleIntersection(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))
	checkFigure(t, triangle, ray, true)
}

func TestTriangleNoBackFaceCulling(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, -1))
	checkFigure(t, triangle, ray, true)
}

func TestTriangleRayOppositeDirection(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, 1))
	checkFigure(t, triangle, ray, false)
}

func TestTriangleRayNearMiss(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(10, 10, -1))
	checkFigure(t, triangle, ray, false)
}

func TestTriangleNotAxisAligned(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(1, 0, 0),
		geom.NewVector(0, 1, 0),
		geom.NewVector(0, 0, 1),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1))
	checkFigure(t, triangle, ray, true)
}

func TestTriangleRayOriginReallyCloseToObject(t *testing.T) {
	triangle := NewTriangle(
		geom.NewVector(-1, 0, 0),
		geom.NewVector(1, 0, 0),
		geom.NewVector(0, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0.5, 1e-6), geom.NewVector(0, 0, 1))
	checkFigure(t, triangle, ray, false)
}

func TestQuadSimpleIntersection(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(1, 1, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, true)
}

func TestQuadNoBackFaceCulling(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(1, 1, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, -1))
	checkFigure(t, quad, ray, true)
}

func TestQuadRayOppositeDirection(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(1, 1, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 1), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, false)
}

func TestQuadNearMiss(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(1, 1, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(10, 10, -1))
	checkFigure(t, quad, ray, false)
}

func TestQuadNonAxisAligned(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(1, 0, 0),
		geom.NewVector(0.5946035575013605, 0.5946035575013605, 0),
		geom.NewVector(0, 1, 0),
		geom.NewVector(0, 0, 1),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1))
	checkFigure(t, quad, ray, true)
}

func TestQuadOriginReallyClosedToObject(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(1, 1, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 1e-6), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, false)
}

func TestQuadIrregularHit(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(10, 10, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, true)
}

func TestQuadSecondIrregularHit(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(0.2, 0.2, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, -1), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, true)
}

func TestQuadThirdIrregularHit(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(-0.5, -0.5, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(-0.75, -0.75, -1), geom.NewVector(0, 0, 1))
	checkFigure(t, quad, ray, true)
}

func TestQuadIrregularMiss(t *testing.T) {
	quad := NewQuad(
		geom.NewVector(-1, -1, 0),
		geom.NewVector(1, -1, 0),
		geom.NewVector(10, 10, 0),
		geom.NewVector(-1, 1, 0),
	)
	ray := geom.NewRay(geom.NewVector(0, 0, 5), geom.NewVector(-10, 10, -1))
	checkFigure(t, quad, ray, false)
}

func TestSphereSimpleIntersection(t *testing.T) {
	sphere := NewSphere(geom.NewVector(5, 5, 5), 3)
	ray := geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 1))
	checkFigure(t, sphere, ray, true)
}

func TestSphereNoBackFaceCulling(t *testing.T) {
	sphere := NewSphere(geom.NewVector(5, 5, 5), 3)
	ray := geom.NewRay(geom.NewVector(5, 5, 5), geom.NewVector(3, 2, 1))
	checkFigure(t, sphere, ray, true)
}

func TestSphereRayOppositeDirection(t *testing.T) {
	sphere := NewSphere(geom.NewVector(5, 5, 5), 3)
	ray := geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(-1, -1, -1))
	checkFigure(t, sphere, ray, false)
}

func TestSphereNearMiss(t *testing.T) {
	sphere := NewSphere(geom.NewVector(5, 5, 5), 1)
	ray := geom.NewRay(geom.NewVector(0, 0, 0), geom.NewVector(1, 1, 8))
	checkFigure(t, sphere, ray, false)
}

func checkFigure(t *testing.T, fig geom.Intersectable, ray geom.Ray, intersection bool) {
	actual := fig.Intersect(ray)
	if actual != intersection {
		t.Errorf("Expected intersection to be %t but it was not", intersection)
	}
}
