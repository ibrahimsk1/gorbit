package entities

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vec2", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:vector-math", "r:low"), func() {
	const epsilon = 1e-9

	Describe("Basic operations", func() {
		It("creates a new Vec2 with given coordinates", func() {
			v := NewVec2(3.0, 4.0)
			Expect(v.X).To(Equal(3.0))
			Expect(v.Y).To(Equal(4.0))
		})

		It("creates a zero vector", func() {
			z := Zero()
			Expect(z.X).To(Equal(0.0))
			Expect(z.Y).To(Equal(0.0))
		})

		It("adds two vectors", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			result := a.Add(b)
			Expect(result.X).To(Equal(4.0))
			Expect(result.Y).To(Equal(6.0))
		})

		It("subtracts two vectors", func() {
			a := NewVec2(5.0, 7.0)
			b := NewVec2(2.0, 3.0)
			result := a.Sub(b)
			Expect(result.X).To(Equal(3.0))
			Expect(result.Y).To(Equal(4.0))
		})

		It("scales a vector by a scalar", func() {
			v := NewVec2(2.0, 3.0)
			result := v.Scale(2.5)
			Expect(result.X).To(Equal(5.0))
			Expect(result.Y).To(Equal(7.5))
		})

		It("computes dot product", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			result := a.Dot(b)
			Expect(result).To(Equal(11.0)) // 1*3 + 2*4 = 11
		})

		It("computes length of a vector", func() {
			v := NewVec2(3.0, 4.0)
			Expect(v.Length()).To(BeNumerically("~", 5.0, epsilon))
		})

		It("computes squared length of a vector", func() {
			v := NewVec2(3.0, 4.0)
			Expect(v.LengthSq()).To(Equal(25.0)) // 3^2 + 4^2 = 25
		})

		It("normalizes a vector", func() {
			v := NewVec2(3.0, 4.0)
			normalized := v.Normalize()
			Expect(normalized.Length()).To(BeNumerically("~", 1.0, epsilon))
		})
	})

	Describe("Property tests - Commutativity", func() {
		It("vector addition is commutative", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			Expect(a.Add(b)).To(Equal(b.Add(a)))
		})

		It("dot product is commutative", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			Expect(a.Dot(b)).To(Equal(b.Dot(a)))
		})
	})

	Describe("Property tests - Associativity", func() {
		It("vector addition is associative", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			c := NewVec2(5.0, 6.0)
			left := a.Add(b).Add(c)
			right := a.Add(b.Add(c))
			Expect(left).To(Equal(right))
		})
	})

	Describe("Property tests - Distributivity", func() {
		It("scalar multiplication distributes over vector addition", func() {
			a := NewVec2(1.0, 2.0)
			b := NewVec2(3.0, 4.0)
			s := 2.5
			left := a.Scale(s).Add(b.Scale(s))
			right := a.Add(b).Scale(s)
			Expect(left.X).To(BeNumerically("~", right.X, epsilon))
			Expect(left.Y).To(BeNumerically("~", right.Y, epsilon))
		})

		It("scalar addition distributes over vector scaling", func() {
			v := NewVec2(2.0, 3.0)
			s1 := 1.5
			s2 := 2.5
			left := v.Scale(s1 + s2)
			right := v.Scale(s1).Add(v.Scale(s2))
			Expect(left.X).To(BeNumerically("~", right.X, epsilon))
			Expect(left.Y).To(BeNumerically("~", right.Y, epsilon))
		})
	})

	Describe("Property tests - Identity elements", func() {
		It("zero vector has zero length", func() {
			z := Zero()
			Expect(z.Length()).To(Equal(0.0))
		})

		It("adding zero vector is identity", func() {
			v := NewVec2(3.0, 4.0)
			z := Zero()
			Expect(v.Add(z)).To(Equal(v))
		})

		It("scaling by 1 is identity", func() {
			v := NewVec2(3.0, 4.0)
			Expect(v.Scale(1.0)).To(Equal(v))
		})
	})

	Describe("Property tests - Length properties", func() {
		It("length is always non-negative", func() {
			v := NewVec2(3.0, 4.0)
			Expect(v.Length()).To(BeNumerically(">=", 0.0))
		})

		It("normalized vector has length 1", func() {
			v := NewVec2(3.0, 4.0)
			normalized := v.Normalize()
			Expect(normalized.Length()).To(BeNumerically("~", 1.0, epsilon))
		})

		It("length squared equals length squared", func() {
			v := NewVec2(3.0, 4.0)
			length := v.Length()
			lengthSq := v.LengthSq()
			Expect(lengthSq).To(BeNumerically("~", length*length, epsilon))
		})
	})

	Describe("Edge cases", func() {
		It("handles very small vectors", func() {
			v := NewVec2(1e-10, 1e-10)
			Expect(v.Length()).To(BeNumerically(">=", 0.0))
			normalized := v.Normalize()
			Expect(normalized.Length()).To(BeNumerically("~", 1.0, epsilon))
		})

		It("handles very large vectors", func() {
			v := NewVec2(1e10, 1e10)
			Expect(v.Length()).To(BeNumerically(">=", 0.0))
			normalized := v.Normalize()
			Expect(normalized.Length()).To(BeNumerically("~", 1.0, epsilon))
		})

		It("handles negative scalars", func() {
			v := NewVec2(2.0, 3.0)
			result := v.Scale(-1.5)
			Expect(result.X).To(Equal(-3.0))
			Expect(result.Y).To(Equal(-4.5))
		})

		It("handles zero scalar", func() {
			v := NewVec2(2.0, 3.0)
			result := v.Scale(0.0)
			Expect(result).To(Equal(Zero()))
		})

		It("normalize of zero vector returns zero vector", func() {
			z := Zero()
			normalized := z.Normalize()
			Expect(normalized).To(Equal(Zero()))
		})
	})
})

func TestEntities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Entities Suite")
}
