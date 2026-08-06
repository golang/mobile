// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package structs

type S struct {
	X, Y       float64
	unexported bool
}

func (s *S) Sum() float64 {
	return s.X + s.Y
}

func (s *S) Identity() (*S, error) {
	return s, nil
}

func Identity(s *S) *S {
	return s
}

func IdentityWithError(s *S) (*S, error) {
	return s, nil
}

func (s *S) Repeat(n int) []*S {
	return Repeat(s, n)
}

func (s *S) RepeatWithError(n int) ([]*S, error) {
	return Repeat(s, n), nil
}

func Repeat(s *S, n int) []*S {
	t := make([]*S, n)
	for i := range t {
		t[i] = s
	}
	return t
}

func RepeatWithError(s *S, n int) ([]*S, error) {
	return Repeat(s, n), nil
}

func SumAll(s []*S) float64 {
	var sum float64
	for _, e := range s {
		sum += e.Sum()
	}
	return sum
}

func SumAllWithError(s []*S) (float64, error) {
	return SumAll(s), nil
}

type (
	S2 struct{}
	I  interface {
		M()
	}
	// Slicer is implemented by both Go and foreign code, exercising
	// slices of structs in both directions.
	Slicer interface {
		Slice(s []*S) []*S
	}
)

func (s *S2) M() {
}

func (_ *S2) String() string {
	return ""
}

// Structs is a struct with the same name as its package.
type Structs struct {
	Elems []*S
}

func (_ *Structs) M() {
}
