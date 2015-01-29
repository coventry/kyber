package poly

import (
	"testing"

	"github.com/dedis/crypto/abstract"
	"github.com/dedis/crypto/edwards"
	"github.com/dedis/crypto/random"
)

/* This file is a testing suite for sharing.go. It provides multiple test cases
 * for ensuring that encryption schemes built upon this package such as Shamir
 * secret sharing are safe and secure.
 *
 * The tests can also serve as references for how to work with this library.
 */

/* Global Variables */

var group abstract.Group = new(edwards.ExtendedCurve).Init(
	edwards.Param25519(), false)
var altGroup abstract.Group = new(edwards.ProjectiveCurve).Init(
	edwards.ParamE382(), false)
var k int = 10
var n int = 20
var secret = group.Secret().Pick(random.Stream)
var point = group.Point().Mul(group.Point().Base(), secret)
var altSecret = altGroup.Secret().Pick(random.Stream)
var altPoint = altGroup.Point().Mul(altGroup.Point().Base(), altSecret)

/* Setup Functions
 *
 * These functions provide greater modularity by consolidating commonly used
 * setup tasks.
 *
 * Not every function uses these methods, since they may have unique set-up
 * needs that do not warrant their own set-up function.
 */

// Tests that checks whether a method panics can use this funcition
func deferTest(t *testing.T, message string) {
	if r := recover(); r == nil {
		t.Error(message)
	}
}

func producePriPoly(g abstract.Group, k int, s abstract.Secret) *PriPoly {
	return new(PriPoly).Pick(g, k, s, random.Stream)
}

func producePriShares(g abstract.Group, k, n int, s abstract.Secret) *PriShares {

	testPoly := producePriPoly(g, k, s)
	return new(PriShares).Split(testPoly, n)
}

func producePubPoly(g abstract.Group, k, n int, s abstract.Secret,
	p abstract.Point) *PubPoly {

	testPriPoly := producePriPoly(g, k, s)
	testPubPoly := new(PubPoly)
	testPubPoly.Init(g, n, p)
	return testPubPoly.Commit(testPriPoly, p)
}

func producePubShares(g abstract.Group, k, n, t int, s abstract.Secret,
	p abstract.Point) *PubShares {

	testPubPoly := producePubPoly(g, k, n, s, p)
	return new(PubShares).Split(testPubPoly, t)
}

/* Test Functions */

func TestPriPolyPick(t *testing.T) {
	// Test that the Pick function creates unique polynomials and
	// unique secrets.
	testPoly1 := producePriPoly(group, k, nil)
	testPoly2 := producePriPoly(group, k, nil)
	testPoly3 := producePriPoly(group, k, nil)
	if testPoly1.Equal(testPoly2) || testPoly1.Equal(testPoly3) ||
		testPoly2.Equal(testPoly3) {
		t.Error("Failed to create unique polynomials.")
	}
	if testPoly1.Secret().Equal(testPoly2.Secret()) ||
		testPoly1.Secret().Equal(testPoly3.Secret()) ||
		testPoly2.Secret().Equal(testPoly3.Secret()) {
		t.Error("Failed to create unique secrets.")
	}

	// Test polynomials that are based on common secrets. Verify that
	// unique polynomials are created but that the base secrets are the same.
	testPoly1 = producePriPoly(group, k, secret)
	testPoly2 = producePriPoly(group, k, secret)
	testPoly3 = producePriPoly(group, k, secret)
	if testPoly1.Equal(testPoly2) || testPoly1.Equal(testPoly3) ||
		testPoly2.Equal(testPoly3) {
		t.Error("Failed to create unique polynomials.")
	}
	if !testPoly1.Secret().Equal(testPoly2.Secret()) ||
		!testPoly1.Secret().Equal(testPoly3.Secret()) ||
		!testPoly2.Secret().Equal(testPoly3.Secret()) {
		t.Error("Polynomials are expected to have the same secret.")
	}
}

// Verify that the secret function works. The function should return the same
// secret it was initialized with.
func TestPriPolySecret(t *testing.T) {
	testPoly := producePriPoly(group, k, secret)
	if !secret.Equal(testPoly.Secret()) {
		t.Error("This secret differs from the one given to it.")
	}
}

func TestPriPolyEqual(t *testing.T) {
	// Verify that Equal returns true for two polynomials that are the same
	testPoly := producePriPoly(group, k, secret)
	if !testPoly.Equal(testPoly) {
		t.Error("Polynomials are expected to be equal.")
	}

	// Verify that Equal returns false for two polynomials that are diffferent
	testPoly1 := producePriPoly(group, k, secret)
	testPoly2 := producePriPoly(group, k, secret)
	if testPoly1.Equal(testPoly2) {
		t.Error("Polynomials are expected to be different.")
	}
}

func TestPriPolyEqual_Error(t *testing.T) {
	test := func(p1, p2 *PriPoly) {
		defer deferTest(t, "The Equal method should have panicked.")
		p1.Equal(p2)
	}

	// Verify that Equal panics if the polynomials are of different degrees.
	testPoly1 := producePriPoly(group, k, secret)
	testPoly2 := producePriPoly(group, k+10, secret)
	test(testPoly1, testPoly2)

	// Verify that Equal panics if the polynomials are of different groups.
	testPoly1 = producePriPoly(group, k, secret)
	testPoly2 = producePriPoly(altGroup, k, altSecret)
	test(testPoly1, testPoly2)
}

// Verify that the add function properly adds two polynomials together.
func TestPriPolyAdd(t *testing.T) {
	testPoly1 := producePriPoly(group, k, secret)
	testPoly2 := producePriPoly(group, k, secret)
	testAddedPoly := new(PriPoly).Add(testPoly1, testPoly2)
	for i := 0; i < k; i++ {
		addedResult := testAddedPoly.g.Secret().Add(testPoly1.s[i],
			testPoly2.s[i])
		if !testAddedPoly.s[i].Equal(addedResult) {
			t.Error("Polynomials not added together properly.")
		}
	}
}

func TestPriPolyAdd_Error(t *testing.T) {
	test := func(p1, p2 *PriPoly) {
		defer deferTest(t, "The Add method should have panicked.")
		new(PriPoly).Add(p1, p2)
	}

	// Verify that the add function panics if the polynomials are different degrees.
	testPoly1 := producePriPoly(group, k, secret)
	testPoly2 := producePriPoly(group, k+10, secret)
	test(testPoly1, testPoly2)

	// Verify the add function panics if the polynomials are of different groups.
	testPoly1 = producePriPoly(group, k, secret)
	testPoly2 = producePriPoly(altGroup, k, altSecret)
	test(testPoly1, testPoly2)
}

// Verify that the string function returns a string representation of the
// polynomial. The test simply assures that the function exits successfully.
func TestPriPolyString(t *testing.T) {
	testPoly := producePriPoly(group, k, secret)
	result := testPoly.String()
	t.Log(result)
}

// Tests the split and share function. Splits a private polynomial and ensures
// that share i is the private polynomial evaluated at point i.
func TestPriSharesSplitShare(t *testing.T) {
	testPoly := producePriPoly(group, k, secret)
	testShares := new(PriShares).Split(testPoly, n)
	errorString := "Share %v should equal the polynomial evaluated at %v"
	for i := 0; i < n; i++ {
		if !testShares.Share(i).Equal(testPoly.Eval(i)) {
			t.Error(errorString, i, i)
		}
	}
}

// This verifies that Empty properly creates a fresh, empty private share.
func TestPriSharesEmpty(t *testing.T) {
	testShares := producePriShares(group, k, n, secret)
	testShares.Empty(group, k+1, n+1)
	if group.String() != testShares.g.String() || testShares.k != k+1 ||
		len(testShares.s) != n+1 {
		t.Error("Empty failed to set the share object properly.")
	}
	for i := 0; i < n+1; i++ {
		if testShares.Share(i) != nil {
			t.Error("Share should be nil.")
		}
	}
}

// This verifies the SetShare function. It sets the share and then ensures that
// the share returned is as expected.
func TestPriSharesSetShare(t *testing.T) {
	testShares := producePriShares(group, k, n, secret)
	testShares.Empty(group, k, n)
	testShares.SetShare(0, altSecret)
	if !altSecret.Equal(testShares.Share(0)) {
		t.Error("The share was not set properly.")
	}
}

// This verifies that the xCoord function can successfully create an array with
// k secrets from a PriShare with sufficient secrets.
func TestPriSharesxCoord(t *testing.T) {
	testShares := producePriShares(group, k, k, secret)
	x := testShares.xCoords()
	c := 0
	for i := 0; i < len(x); i++ {
		if x[i] != nil {
			c += 1
		}
	}
	if c < k {
		t.Error("Expected %v points to be made.", k)
	}
}

// Ensures that if we have k-1 shares, xCoord panics.
func TestPriSharesxCoord_Error(t *testing.T) {
	defer deferTest(t, "The XCoord method should have panicked.")
	testShares := producePriShares(group, k, k, secret)
	testShares.s[0] = nil
	testShares.xCoords()
}

// Ensures that we can successfully reconstruct the secret if given k shares.
func TestPriSharesSecret(t *testing.T) {
	testShares := producePriShares(group, k, k, secret)
	result := testShares.Secret()
	if !secret.Equal(result) {
		t.Error("The secret failed to be reconstructed.")
	}
}

// Ensures that we fail to reconstruct the secret with too little shares.
func TestPriSharesSecret_Error(t *testing.T) {
	defer deferTest(t, "The Secret method should have panicked.")
	testShares := producePriShares(group, k, k, secret)
	testShares.s[0] = nil
	testShares.Secret()
}

// Tests the string function by simply verifying that it runs to completion.
func TestPriSharesString(t *testing.T) {
	testShares := producePriShares(group, k, n, secret)
	result := testShares.String()
	t.Log(result)
}

// Tests Init to insuring it can create a public polynomial correctly.
func TestPubPolyInit(t *testing.T) {
	testPoly := new(PubPoly)
	testPoly.Init(group, k, point)
	if group.String() != testPoly.g.String() || !point.Equal(testPoly.b) ||
		k != len(testPoly.p) {
		t.Error("The public polynomial was not initialized properly.")
	}
}

func TestPubPolyCommit(t *testing.T) {
	// Tests Commit to ensure it properly commits a privatepolynomial.
	testPriPoly := new(PriPoly).Pick(group, k, secret, random.Stream)
	testPubPoly := new(PubPoly)
	testPubPoly.Init(group, k, point)
	testPubPoly = testPubPoly.Commit(testPriPoly, point)
	for i := 0; i < len(testPubPoly.p); i++ {
		if !group.Point().Mul(point, testPriPoly.s[i]).Equal(testPubPoly.p[i]) {
			t.Error("PriPoly should be multiplied by the point")
		}
	}

	// Tests commit to ensure it works with the standard base.
	testPriPoly = producePriPoly(group, k, secret)
	testPubPoly = new(PubPoly)
	testPubPoly.Init(group, k, nil)
	testPubPoly = testPubPoly.Commit(testPriPoly, nil)
	for i := 0; i < len(testPubPoly.p); i++ {
		if !point.Mul(nil, testPriPoly.s[i]).Equal(testPubPoly.p[i]) {
			t.Error("PriPoly should be multiplied by the point")
		}
	}
}

// Verifies SecretCommit returns the altered secret from the private polynomial.
func TestPubPolySecretCommit(t *testing.T) {
	testPriPoly := producePriPoly(group, k, secret)
	testPubPoly := new(PubPoly)
	testPubPoly.Init(group, k, point)
	testPubPoly = testPubPoly.Commit(testPriPoly, point)
	secretCommit := testPubPoly.SecretCommit()
	if !point.Mul(point, testPriPoly.s[0]).Equal(secretCommit) {
		t.Error("The secret commit is not from the private secret")
	}
}

// Encode a public polynomial and verify its length is as expected.
func TestPubPolyLen(t *testing.T) {
	testPubPoly := producePubPoly(group, k, k, secret, point)
	if testPubPoly.Len() != len(testPubPoly.Encode()) {
		t.Error("The length should equal the length of the encoding")
	}
}

// Encode a public polynomial and then decode it.
func TestPubPolyEncodeDecode(t *testing.T) {
	testPubPoly := producePubPoly(group, k, k, secret, point)
	decodePubPoly := new(PubPoly)
	decodePubPoly.Init(group, k, point)
	if err := decodePubPoly.Decode(testPubPoly.Encode()); err != nil ||
		!decodePubPoly.Equal(testPubPoly) {
		t.Error("Failed to encode/ decode properly.")
	}
}

func TestPubPolyEncodeDecode_Error(t *testing.T) {
	test := func(p1 *PubPoly) {
		defer deferTest(t, "The Encode method should have panicked.")
		p1.Encode()
	}
	
	// Verify that encode fails if the group and point are not the same length
	// (aka not from the same group in this case).
	testPubPoly := producePubPoly(group, k, k, secret, altPoint)
	test(testPubPoly)

	// Verify decoding/ encoding fails if the new polynomial is the wrong length.
	testPubPoly    = producePubPoly(group, k, k, secret, point)
	decodePubPoly := new(PubPoly)
	decodePubPoly.Init(group, k+20, point)
	if err := decodePubPoly.Decode(testPubPoly.Encode()); err == nil {
		t.Error("Decode should fail.")
	}
}

func TestPubPolyEqual(t *testing.T) {
	// Verify that Equal returns true for two polynomials that are the same
	testPubPoly := producePubPoly(group, k, k, secret, point)
	if !testPubPoly.Equal(testPubPoly) {
		t.Error("Polynomials are expected to be equal.")
	}

	// Verify that Equal returns false for two polynomials that are diffferent
	testPubPoly1 = producePubPoly(group, k, k, secret, point)
	testPubPoly2 = producePubPoly(group, k, k, secret, nil)
	if testPubPoly1.Equal(testPubPoly2) {
		t.Error("Polynomials are expected to be different.")
	}
}

func TestPubPolyEqual_Error(t *testing.T) {
	test := func(p1, p2 *PubPoly) {
		defer deferTest(t, "The Equal method should have panicked.")
		p11.Equal(p2)
	}

	// Verify that Equal panics if the polynomials are different degrees.
	testPubPoly1 := producePubPoly(group, k, k, secret, point)
	testPubPoly2 := producePubPoly(group, k+10, k+10, secret, point)
	test(testPubPoly1, testPubPoly2)

	// Verify that Equal panics if the polynomials are of different groups.
	testPubPoly1 := producePubPoly(group, k, k, secret, point)
	testPubPoly2 := producePubPoly(altGroup, k, k, altSecret, altPoint)
	test(testPubPoly1, testPubPoly2)
}

// Verify that Add can successfully add two polynomials
func TestPubPolyAdd(t *testing.T) {
	testPubPoly1 := producePubPoly(group, k, k, secret, point)
	testPubPoly2 := producePubPoly(group, k, k, secret, point)
	testAddedPoly := new(PubPoly).Add(testPubPoly1, testPubPoly2)
	for i := 0; i < k; i++ {
		addResult := testAddedPoly.g.Point().Add(testPubPoly1.p[i],
			testPubPoly2.p[i])
		if !testAddedPoly.p[i].Equal(addResult) {
			t.Error("Polynomials not added together properly.")
		}
	}
}

func TestPubPolyAdd_Error(t *testing.T) {
	test := func(p1, p2 *PubPoly) {
		defer deferTest(t, "The Add method should have panicked.")
		new(PubPoly).Add(p1, p2)
	}

	// Verify that Add panics if the polynomials are of different degrees.
	testPubPoly1 := producePubPoly(group, k, k, secret, point)
	testPubPoly2 := producePubPoly(group, k+10, k+10, secret, point)
	test(testPubPoly1, testPubPoly2)

	// Verify that Add panics if the polynomials are of different groups.
	testPubPoly  = producePubPoly(group, k, k, secret, point)
	testPubPoly2 = producePubPoly(altGroup, k, k, altSecret, altPoint)
	test(testPubPoly1, testPubPoly2)
}

// Verifies that Check correctly identifies a valid share.
func TestPubPolyCheck(t *testing.T) {
	testPriPoly := producePriPoly(group, k, secret)
	testPubPoly := new(PubPoly)
	testPubPoly.Init(group, k, point)
	testPubPoly = testPubPoly.Commit(testPriPoly, point)
	testShares := new(PriShares).Split(testPriPoly, n)
	if testPubPoly.Check(1, testShares.Share(1)) == false {
		t.Error("The share should be accepted.")
	}
}

// Verifies that Check correctly rejects an invalid share.
func TestPubPolyCheck_Errpr(t *testing.T) {
	testPubPoly := producePubPoly(group, k, k, secret, point)
	testSharesBad := producePriShares(altGroup, n, n, altSecret)
	if testPubPoly.Check(0, testSharesBad.Share(0)) == true {
		t.Error("The share should be rejected.")
	}
}

// Verify that the string function returns a string representation of the
// polynomial. The test simply assures that the function exits successfully.
func TestPubPolyString(t *testing.T) {
	testPubPoly := producePubPoly(group, k, k, secret, point)
	result := testPubPoly.String()
	t.Log(result)
}

// This function tests the eval functions for both PriPoly and PubPoly
func TestPolyEval(t *testing.T) {
	testPriPoly := new(PriPoly).Pick(group, k, secret, random.Stream)
	testPubPoly := new(PubPoly)
	testPubPoly.Init(group, k, point)
	testPubPoly = testPubPoly.Commit(testPriPoly, point)
	errorString := "PriPoly.Eval(i) * point should equal PubPoly.Eval(i)"
	for i := 0; i < k; i++ {
		priResult := group.Point().Mul(point, testPriPoly.Eval(i))
		if !priResult.Equal(testPubPoly.Eval(i)) {
			t.Error(errorString)
		}
	}
}

// Tests the split and share functions. Splits a public polynomial and
// ensures that share i is the public polynomial evaluated at point i.
func TestPubSharesSplitShare(t *testing.T) {
	testPubPoly := producePubPoly(group, k, k, secret, point)
	testShares := new(PubShares).Split(testPubPoly, n)
	errorString := "Share %v should equal the polynomial evaluated at %v"
	for i := 0; i < n; i++ {
		if !testShares.Share(i).Equal(testPubPoly.Eval(i)) {
			t.Error(errorString, i, i)
		}
	}
}

// This verifies the SetShare function. It sets the share and then ensures that
// the share returned is as expected.
func TestPubSharesSetShare(t *testing.T) {
	testShares := producePubShares(group, k, k, n, secret, point)
	testShares.SetShare(0, point.Add(point, point))
	if !point.Equal(testShares.Share(0)) {
		t.Error("The share was not set properly.")
	}
}

// This verifies that the xCoord function can successfully create an array with
// k secrets from a PubShare with sufficient secrets.
func TestPubSharesxCoord(t *testing.T) {
	testShares := producePubShares(group, k, k, k, secret, point)
	x := testShares.xCoords()
	c := 0
	for i := 0; i < len(x); i++ {
		if x[i] != nil {
			c += 1
		}
	}
	if c < testShares.k {
		t.Error("Expected %v points to be made.", k)
	}
}

// Ensures that if given k-1 shares, xCoord panics.
func TestPubSharesxCoord_Error(t *testing.T) {
	defer deferTest(t, "The XCoord method should have panicked.")
	testShares := producePubShares(group, k, k, k, secret, point)
	testShares.p[0] = nil
	testShares.xCoords()
}

// Ensures that the code successfully reconstructs the secret if given k shares.
func TestPubSharesSecret(t *testing.T) {
	testShares := producePubShares(group, k, k, k, secret, point)
	result := testShares.SecretCommit()
	if !point.Mul(point, secret).Equal(result) {
		t.Error("The point failed to be reconstructed.")
	}
}

// Ensures that the code fails to reconstruct the secret with too little shares.
func TestPubSharesSecret_Error(t *testing.T) {
	defer deferTest(t, "The SecretCommit method should have panicked.")
	testShares := producePubShares(group, k, k, k, secret, point)
	testShares.p[0] = nil
	testShares.SecretCommit()
}

// Tests the string function by simply verifying that it runs to completion.
func TestPubSharesString(t *testing.T) {
	testShares := producePubShares(group, k, k, n, secret, point)
	result := testShares.String()
	t.Log(result)
}
