package entities

import "math"

// Vec2 represents a 2D vector with X and Y components.
type Vec2 struct {
	X, Y float64
}

// NewVec2 creates a new Vec2 with the given coordinates.
func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}

// Zero returns a zero vector.
func Zero() Vec2 {
	return Vec2{X: 0, Y: 0}
}

// Add returns the sum of this vector and another.
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub returns the difference of this vector and another.
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Scale returns this vector scaled by the given scalar.
func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{
		X: v.X * s,
		Y: v.Y * s,
	}
}

// Dot returns the dot product of this vector and another.
func (v Vec2) Dot(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

// LengthSq returns the squared length of this vector.
// This is useful for comparisons where the actual length is not needed,
// avoiding the square root computation.
func (v Vec2) LengthSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Length returns the magnitude (length) of this vector.
func (v Vec2) Length() float64 {
	return math.Sqrt(v.LengthSq())
}

// Normalize returns a normalized (unit) vector in the same direction.
// If the vector is zero, it returns a zero vector.
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Zero()
	}
	return v.Scale(1.0 / length)
}

