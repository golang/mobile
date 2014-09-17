// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package f32

import (
	"math"
	"testing"
)

var x = Mat4{
	{0, 1, 2, 3},
	{4, 5, 6, 7},
	{8, 9, 10, 11},
	{12, 13, 14, 15},
}

var xsq = Mat4{
	{56, 62, 68, 74},
	{152, 174, 196, 218},
	{248, 286, 324, 362},
	{344, 398, 452, 506},
}

var identity = Mat4{
	{1, 0, 0, 0},
	{0, 1, 0, 0},
	{0, 0, 1, 0},
	{0, 0, 0, 1},
}

func TestEq(t *testing.T) {
	var tests = []struct {
		m0, m1 Mat4
		eq     bool
	}{
		{x, x, true},
		{identity, identity, true},
		{x, identity, false},
	}

	for _, test := range tests {
		got := test.m0.Eq(&test.m1, 0.01)
		if got != test.eq {
			t.Errorf("Eq=%v, want %v for\n%s\n%s", got, test.eq, test.m0, test.m1)
		}
	}
}

func TestMat4Mul(t *testing.T) {
	var tests = []struct{ m0, m1, want Mat4 }{
		{x, identity, x},
		{identity, x, x},
		{x, x, xsq},
		{
			Mat4{
				{1.811, 0.000, +0.000, +0.000},
				{0.000, 2.414, +0.000, +0.000},
				{0.000, 0.000, -1.010, -1.000},
				{0.000, 0.000, -2.010, +0.000},
			},
			Mat4{
				{+0.992, -0.015, +0.123, 0.000},
				{+0.000, +0.992, +0.123, 0.000},
				{-0.124, -0.122, +0.985, 0.000},
				{-0.000, -0.000, -8.124, 1.000},
			},
			Mat4{
				{+1.797, -0.037, -0.124, -0.123},
				{+0.000, +2.396, -0.124, -0.123},
				{-0.225, -0.295, -0.995, -0.985},
				{+0.000, +0.000, +6.196, +8.124},
			},
		},
	}

	for _, test := range tests {
		got := Mat4{}
		got.Mul(&test.m0, &test.m1)
		if !got.Eq(&test.want, 0.01) {
			t.Errorf("%s *\n%s =\n%s, want\n%s", test.m0, test.m1, got, test.want)
		}
	}
}

func TestLookAt(t *testing.T) {
	var tests = []struct {
		eye, center, up Vec3
		want            Mat4
	}{
		{
			Vec3{1, 1, 8}, Vec3{0, 0, 0}, Vec3{0, 1, 0},
			Mat4{
				{0.992, -0.015, 0.123, 0.000},
				{0.000, 0.992, 0.123, 0.000},
				{-0.124, -0.122, 0.985, 0.000},
				{-0.000, -0.000, -8.124, 1.000},
			},
		},
		{
			Vec3{4, 5, 7}, Vec3{0.1, 0.2, 0.3}, Vec3{0, -1, 0},
			Mat4{
				{-0.864, 0.265, 0.428, 0.000},
				{0.000, -0.850, 0.526, 0.000},
				{0.503, 0.455, 0.735, 0.000},
				{-0.064, 0.007, -9.487, 1.000},
			},
		},
	}

	for _, test := range tests {
		got := Mat4{}
		got.LookAt(&test.eye, &test.center, &test.up)
		if !got.Eq(&test.want, 0.01) {
			t.Errorf("LookAt(%s,%s%s) =\n%s\nwant\n%s", test.eye, test.center, test.up, got, test.want)
		}
	}

}

func TestPerspective(t *testing.T) {
	want := Mat4{
		{1.811, 0.000, 0.000, 0.000},
		{0.000, 2.414, 0.000, 0.000},
		{0.000, 0.000, -1.010, -1.000},
		{0.000, 0.000, -2.010, 0.000},
	}
	got := Mat4{}

	got.Perspective(Radian(math.Pi/4), 4.0/3, 1, 200)

	if !got.Eq(&want, 0.01) {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}

}

func TestRotate(t *testing.T) {
	want := &Mat4{
		{-8.000, -9.000, -10.000, -11.000},
		{4.000, 5.000, 6.000, 7.000},
		{-0.000, 1.000, 2.000, 3.000},
		{12.000, 13.000, 14.000, 15.000},
	}

	got := new(Mat4)
	got.Rotate(&x, Radian(math.Pi/2), &Vec3{0, 1, 0})

	if !got.Eq(want, 0.01) {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestScale(t *testing.T) {
	want := &Mat4{
		{0.000, 2.000, 4.000, 6.000},
		{12.000, 15.000, 18.000, 21.000},
		{32.000, 36.000, 40.000, 44.000},
		{12.000, 13.000, 14.000, 15.000},
	}

	got := new(Mat4)
	got.Scale(&x, &Vec3{2, 3, 4})

	if !got.Eq(want, 0.01) {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestTranslate(t *testing.T) {
	want := &Mat4{
		{0.000, 1.000, 2.000, 3.000},
		{4.000, 5.000, 6.000, 7.000},
		{8.000, 9.000, 10.000, 11.000},
		{15.200, 16.800, 18.400, 20.000},
	}

	got := new(Mat4)
	got.Translate(x, &Vec3{0.1, 0.2, 0.3})

	if !got.Eq(want, 0.01) {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
