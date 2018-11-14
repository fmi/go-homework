package geom

// Ray represents a geometric ray defined by an origin point and direction vector.
type Ray struct {
	Origin    Vector
	Direction Vector
}

// NewRay returns a ray defined by its origin point `origin` and direction `dir`.
func NewRay(origin, dir Vector) Ray {
	return Ray{
		Origin:    origin,
		Direction: dir,
	}
}
