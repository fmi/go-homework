package geom

// Vector represents a point or a direction vector in a 3D cartesian coordinate system.
type Vector struct {
	// X, Y and Z are the coordinates of a point in each of the three
	// dimensions or alternatively the direction defined by a vector
	// with origin (0,0,0) and end at X, Y and Z.
	X, Y, Z float64
}

// NewVector returns a Vector with coordinates `x`, `y` and `z`.
func NewVector(x, y, z float64) Vector {
	return Vector{
		X: x,
		Y: y,
		Z: z,
	}
}
