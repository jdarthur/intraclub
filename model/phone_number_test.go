package model

import (
	"fmt"
	"testing"
)

func TestOneDash(t *testing.T) {
	p := PhoneNumber("123-4567890")
	err := p.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("%+v\n", err)
}

func TestThreeDashes(t *testing.T) {
	p := PhoneNumber("123-456-78-90")
	err := p.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("%+v\n", err)
}

func TestWrongDashLocation(t *testing.T) {
	p := PhoneNumber("123-45-67890")
	err := p.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("%+v\n", err)
}

func TestWrongAmountOfNumbers(t *testing.T) {
	p := PhoneNumber("123-456-789")
	err := p.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("%+v\n", err)
}

func TestDashes(t *testing.T) {
	p := PhoneNumber("123-456-7890")
	err := p.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoDashes(t *testing.T) {
	p := PhoneNumber("1234567890").AddDashes()
	err := p.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddDashesWhenExisting(t *testing.T) {
	p := PhoneNumber("123-456-7890").AddDashes()
	err := p.StaticallyValid()
	if err != nil {
		t.Fatal(err)
	}

	if p != "123-456-7890" {
		t.Fatal("expected 123-456-7890, got ", p)
	}
}

func TestNoDashesAndDashesEquality(t *testing.T) {
	p := PhoneNumber("123-456-7890").AddDashes()
	p2 := PhoneNumber("1234567890").AddDashes()
	if p != p2 {
		t.Fatal("not equal")
	}
}

func TestInvalidCharacters(t *testing.T) {
	p := PhoneNumber("123-456-789a")
	err := p.StaticallyValid()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("%+v\n", err)
}
