package main

import (
	"math"

	"github.com/fmi/go-homework/geom"
)

// Triangle is an Intersectable which represents a triangle in the 3D space.
type Triangle struct {
	a, b, c vector
}

// Intersect implements the geom.Intersecatble interface. It uses the
// Möller–Trumbore ray-triangle intersection algorithm from 1997. Wiki link:
// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm
func (t *Triangle) Intersect(r geom.Ray) bool {
	ray := rayFromGeom(r)
	edge1 := t.b.Minus(t.a)
	edge2 := t.c.Minus(t.a)

	s1 := ray.Direction.Cross(edge2)
	divisor := edge1.Product(s1)

	// Not culling:
	if divisor > -epsilon && divisor < epsilon {
		return false
	}

	invDivisor := 1.0 / divisor

	s := ray.Origin.Minus(t.a)
	b1 := s.Product(s1) * invDivisor

	if b1 < 0.0 || b1 > 1.0 {
		return false
	}

	s2 := s.Cross(edge1)
	b2 := ray.Direction.Product(s2) * invDivisor

	if b2 < 0.0 || b1+b2 > 1.0 {
		return false
	}

	tt := edge2.Product(s2) * invDivisor

	if tt < 0 {
		return false
	}

	return true
}

// NewTriangle returns a new Triangle, defined with the points `a`, `b` and `c`.
func NewTriangle(a, b, c geom.Vector) *Triangle {
	return &Triangle{
		a: vectorFromGeom(a),
		b: vectorFromGeom(b),
		c: vectorFromGeom(c),
	}
}

// Quad is an Intersectable which represents a quadrilateral in the 3D space.
type Quad struct {
	vertices [4]vector
}

// Intersect implements the geom.Intersecatble interface. It is based on the
// Ares Lagae and Philip Dutre (2005) algorithm.
func (q *Quad) Intersect(r geom.Ray) bool {
	ray := rayFromGeom(r)
	e01 := q.vertices[1].Minus(q.vertices[0])
	e03 := q.vertices[3].Minus(q.vertices[0])

	p := ray.Direction.Cross(e03)
	det := e01.Product(p)
	if det == 0 {
		return false
	}
	invDet := 1 / det
	t := ray.Origin.Minus(q.vertices[0])
	alfa := t.Product(p) * invDet
	if alfa < 0 || alfa > 1 {
		return false
	}
	w := t.Cross(e01)
	beta := ray.Direction.Product(w) * invDet
	if beta < 0 || beta > 1 {
		return false
	}

	if alfa+beta > 1 {
		e21 := q.vertices[1].Minus(q.vertices[2])
		e23 := q.vertices[3].Minus(q.vertices[2])

		pp := ray.Direction.Cross(e21)
		detp := e23.Product(pp)
		if detp == 0 {
			return false
		}
		invDetp := 1 / detp
		tp := ray.Origin.Minus(q.vertices[2])
		alfap := tp.Product(pp) * invDetp
		if alfap < 0 {
			return false
		}
		qp := tp.Cross(e23)
		betap := ray.Direction.Product(qp) * invDetp
		if betap < 0 {
			return false
		}
	}

	tDist := e03.Product(w) * invDet

	if tDist < 0 {
		return false
	}

	return true
}

// NewQuad returns a new Quad which is definied byt he four points `a`, `b`, `c`
// and `d`.
func NewQuad(a, b, c, d geom.Vector) *Quad {
	return &Quad{
		vertices: [4]vector{
			vectorFromGeom(a),
			vectorFromGeom(b),
			vectorFromGeom(c),
			vectorFromGeom(d),
		},
	}
}

// Sphere is an Intersectable which represents a perfect sphere in the 3D space.
type Sphere struct {
	o geom.Vector
	r float64
}

// Intersect implements the geom.Intersecatble interface.
func (s *Sphere) Intersect(ray geom.Ray) bool {
	var d = ray.Direction
	var o = ray.Origin

	// Normalize the direction so that we can later check whether the
	// intersection is behind the origin or in front of it.
	dl := 1.0 / math.Sqrt(d.X*d.X+d.Y*d.Y+d.Z*d.Z)
	d.X *= dl
	d.Y *= dl
	d.Z *= dl

	// To make calculations easier, change the coord system so that
	// the sphere center goes in 0,0,0.
	o.X -= s.o.X
	o.Y -= s.o.Y
	o.Z -= s.o.Z

	var a = d.X*d.X + d.Y*d.Y + d.Z*d.Z
	var b = 2 * (d.X*o.X + d.Y*o.Y + d.Z*o.Z)
	var c = o.X*o.X + o.Y*o.Y + o.Z*o.Z - s.r*s.r

	tNear, tFar, ok := quadratic(a, b, c)

	if !ok || (tNear < 0 && tFar < 0) {
		return false
	}

	var retdist = tNear

	if tNear < 0 {
		retdist = tFar
	}

	if retdist < 0 {
		return false
	}

	return true
}

// NewSphere returns a new Sphere with center `o` and radius `r`.
func NewSphere(o geom.Vector, r float64) *Sphere {
	return &Sphere{o: o, r: r}
}

// vector is a algebraic vector which supports few algebraic operations with other
// vectors.
type vector struct {
	X, Y, Z float64
}

func (v vector) Minus(other vector) vector {
	return vector{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v vector) Product(other vector) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func (v vector) Cross(other vector) vector {
	return vector{v.Y*other.Z - v.Z*other.Y,
		v.Z*other.X - v.X*other.Z,
		v.X*other.Y - v.Y*other.X}
}

// vectorFromGeom returns the `vector` which corresponds to `geom.Vector`.
func vectorFromGeom(p geom.Vector) vector {
	return vector{X: p.X, Y: p.Y, Z: p.Z}
}

// ray is a similar to geo.Ray but it is using vector instead of geom.Vector
// values. This makes calculations in this package easier.
type ray struct {
	Direction vector
	Origin    vector
}

// rayFromGeom returns a new `ray` which corresponds to the `geom.Ray`.
func rayFromGeom(r geom.Ray) ray {
	return ray{
		Direction: vectorFromGeom(r.Direction),
		Origin:    vectorFromGeom(r.Origin),
	}
}

// quadratic solves a quadratic equation and returns the two solutions of there are any.
// Its last return value is a boolean and true when there is a solution. The first two
// values are the solutions.
func quadratic(a, b, c float64) (float64, float64, bool) {
	discrim := b*b - 4*a*c
	if discrim <= 0 {
		return 0, 0, false
	}
	rootDiscrim := math.Sqrt(discrim)
	var q float64
	if b < 0 {
		q = -0.5 * (b - rootDiscrim)
	} else {
		q = -0.5 * (b + rootDiscrim)
	}

	t0, t1 := q/a, c/q

	if t0 > t1 {
		t0, t1 = t1, t0
	}

	return t0, t1, true
}

// epsilon is a very small number, indistinguishable from zero in the context of
// calculations in this package. It can be used for defining the precision of
// comparisons and calculations with float values.
const epsilon = 1e-7
