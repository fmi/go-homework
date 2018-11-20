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

// Cross returns a Vector, the cross product of v1 and v2.
func Cross(v1, v2 Vector) (v Vector) {
	v.X = v1.Y*v2.Z - v1.Z*v2.Y
	v.Y = v1.Z*v2.X - v1.X*v2.Z
	v.Z = v1.X*v2.Y - v1.Y*v2.X
	return
}

// Dot returns a float64, the dot product of v1 and v2.
func Dot(v1, v2 Vector) (d float64) {
	d = v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
	return
}

// Sub returns a Vector, the result of subtracting v2 from v1.
func Sub(v1, v2 Vector) (v Vector) {
	v.X = v1.X - v2.X
	v.Y = v1.Y - v2.Y
	v.Z = v1.Z - v2.Z
	return
}

// Add returns a Vector, the result of subtracting v2 from v1.
func Add(v1, v2 Vector) (v Vector) {
	v.X = v1.X + v2.X
	v.Y = v1.Y + v2.Y
	v.Z = v1.Z + v2.Z
	return
}

// Mul returns a Vector, multiplied by the scalar value n
func Mul(v Vector, n float64) (result Vector) {
	result.X = v.X * n
	result.Y = v.Y * n
	result.Z = v.Z * n
	return
}
