package gocsv

import (
	"bytes"
	"encoding/csv"
	"math"
	"testing"
)

func assertLine(t *testing.T, expected, actual []string) {
	if len(expected) != len(actual) {
		t.Fatalf("line length mismatch between expected: %d and actual: %d", len(expected), len(actual))
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Fatalf("mismatch on field %d: %s != %s", i, expected[i], actual[i])
		}

	}
}

func Test_writeTo(t *testing.T) {
	b := bytes.Buffer{}
	e := &encoder{out: &b}
	s := []Sample{
		{Foo: "f", Bar: 1, Baz: "baz", Frop: 0.1},
		{Foo: "e", Bar: 3, Baz: "b", Frop: 6.0 / 13},
	}
	if err := writeTo(csv.NewWriter(e.out), s); err != nil {
		t.Fatal(err)
	}

	lines, err := csv.NewReader(&b).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	assertLine(t, []string{"foo", "BAR", "Baz", "Quux"}, lines[0])
	assertLine(t, []string{"f", "1", "baz", "0.1"}, lines[1])
	assertLine(t, []string{"e", "3", "b", "0.46153846"}, lines[2])
}

func Test_writeTo_embed(t *testing.T) {
	b := bytes.Buffer{}
	e := &encoder{out: &b}
	s := []EmbedSample{
		{
			Qux:    "aaa",
			Sample: Sample{Foo: "f", Bar: 1, Baz: "baz", Frop: 0.2},
			Ignore: "shouldn't be marshalled",
			Quux:   "zzz",
			Grault: math.Pi,
		},
	}
	if err := writeTo(csv.NewWriter(e.out), s); err != nil {
		t.Fatal(err)
	}

	lines, err := csv.NewReader(&b).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	assertLine(t, []string{"first", "foo", "BAR", "Baz", "Quux", "garply", "last"}, lines[0])
	assertLine(t, []string{"aaa", "f", "1", "baz", "0.2", "3.141592653589793", "zzz"}, lines[1])
}

func Test_writeTo_complex_embed(t *testing.T) {
	b := bytes.Buffer{}
	e := &encoder{out: &b}
	sfs := []SkipFieldSample{
		{
			EmbedSample: EmbedSample{
				Qux: "aaa",
				Sample: Sample{
					Foo:  "bbb",
					Bar:  111,
					Baz:  "ddd",
					Frop: 1.2e22,
				},
				Ignore: "eee",
				Grault: 0.1,
				Quux:   "fff",
			},
			MoreIgnore: "ggg",
			Corge:      "hhh",
		},
	}
	if err := writeTo(csv.NewWriter(e.out), sfs); err != nil {
		t.Fatal(err)
	}
	lines, err := csv.NewReader(&b).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	assertLine(t, []string{"first", "foo", "BAR", "Baz", "Quux", "garply", "last", "abc"}, lines[0])
	assertLine(t, []string{"aaa", "bbb", "111", "ddd", "12000000000000000000000", "0.1", "fff", "hhh"}, lines[1])
}
