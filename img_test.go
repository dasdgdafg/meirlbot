package main

import (
	"testing"
)

var img = CuteImage{}

func TestNormalMatches(t *testing.T) {
	matchingStrs := []string{"me irl",
		"me on the left",
		"foo me irl",
		"foo me on the right bar",
		"foo ME IRL",
		"me ON the LEFT foo",
		"me with tags foo"}
	nonMatchingStrs := []string{"",
		"e on the left",
		"me ir",
		"foo me the right bar",
		"me on left",
		"me with tags",
		"me with tags ",
		"skdgjakfgjsdflsjdfasd",
		"meontheright",
		"me 0n th3 l3ft",
		"me  irl"}
	for _, s := range matchingStrs {
		if !img.checkForMatch(s) {
			t.Error("Did not match but should: " + s)
		}
	}
	for _, s := range nonMatchingStrs {
		if img.checkForMatch(s) {
			t.Error("Did match but should not: " + s)
		}
	}
}

func TestColoredMatches(t *testing.T) {
	matchingStrs := []string{"1me1,2 i11rl11,12",
		"me on the le1ft",
		"foo me irl",
		"foo ,2me on the right11 bar",
		"foo ME11,1211,211,121 IRL",
		"m1,2e ON11 the 11,12LEFT foo",
		"m1e w1,2ith tags foo"}
	nonMatchingStrs := []string{"me0 irl",
		"me on the le7ft",
		"foo m,e irl",
		"foo me on the ,-1right bar",
		"foo ME IR111L",
		"me O111,2N the LEFT foo",
		"me with ta1,222gs foo",
		"me i_rl",
		"me on th/e left",
		"foo me\u0000irl",
		"foo me on thあe right bar",
		"foo ME \nIRL",
		"me ON the LEFT foo",
		"me with tags foo"}
	for _, s := range matchingStrs {
		if !img.checkForMatch(s) {
			t.Error("Did not match but should: " + s)
		}
	}
	for _, s := range nonMatchingStrs {
		if img.checkForMatch(s) {
			println(s)
			t.Error("Did match but should not: " + s)
		}
	}
}
