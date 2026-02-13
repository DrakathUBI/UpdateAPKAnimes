package main

import "testing"

func TestEvalSimple(t *testing.T) {
	got, err := evalSimple("40+2*2")
	if err != nil {
		t.Fatal(err)
	}
	if got != 44 {
		t.Fatalf("want 44 got %v", got)
	}
}

func TestExpand(t *testing.T) {
	rt := &Runtime{vars: map[string]string{"nome": "julio"}}
	if got := rt.expand("ola ${nome}"); got != "ola julio" {
		t.Fatalf("unexpected: %s", got)
	}
}
