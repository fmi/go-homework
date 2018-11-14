package geom

// Intersectable represents an object in the 3D space which can be tested for
// intersections with rays.
type Intersectable interface {

	// Intersect returns true when `ray` intersects this object.
	Intersect(ray Ray) bool
}
