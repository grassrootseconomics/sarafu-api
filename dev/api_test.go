package dev

import (
	"context"
	"testing"
)

func TestApiRequestAlias(t *testing.T) {
	ctx := context.Background()
	svc := NewDevAccountService()
	ra, err := svc.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	addr := ra.PublicKey
	
	_, err = svc.RequestAlias(ctx, addr, "+254f00")
	if err == nil {
		t.Fatalf("expected error")		
	}
	alias := "+254712345678"
	rb, err := svc.RequestAlias(ctx, addr, alias)
	if err != nil {
		t.Fatal(err)
	}
	if rb.Alias != alias {
		t.Fatalf("expected '%s', got '%s'", alias, rb.Alias)
	}
	_, err = svc.CheckAliasAddress(ctx, alias)
	if err == nil {
		t.Fatalf("expected error")
	}

	alias = "foo"
	rb, err = svc.RequestAlias(ctx, addr, alias)
	if err != nil {
		t.Fatal(err)
	}
	if rb.Alias != alias {
		t.Fatalf("expected '%s', got '%s'", alias, rb.Alias)
	}
	rc, err := svc.CheckAliasAddress(ctx, alias)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Address != addr {
		t.Fatalf("expected '%s', got '%s'", addr, rc.Address)
	}

	// create a second account
	ra, err = svc.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	addr = ra.PublicKey

	rb, err = svc.RequestAlias(ctx, addr, alias)
	if err != nil {
		t.Fatal(err)
	}
	alias = "foox"
	if rb.Alias != alias {
		t.Fatalf("expected '%s', got '%s'", alias, rb.Alias)
	}
	rc, err = svc.CheckAliasAddress(ctx, alias)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Address != addr {
		t.Fatalf("expected '%s', got '%s'", addr, rc.Address)
	}
}
