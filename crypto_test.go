package zksigma

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func TestECPointMethods(t *testing.T) {
	v := big.NewInt(3)
	p := ZKCurve.G.Mult(v)
	negp := p.Neg()
	sum := p.Add(negp)
	if !sum.Equal(Zero) {
		t.Logf("p : %v\n", p)
		t.Logf("negp : %v\n", negp)
		t.Logf("sum : %v\n", sum)
		t.Fatalf("p + -p should be 0\n")
	}
	negnegp := negp.Neg()
	if !negnegp.Equal(p) {
		t.Logf("p : %v\n", p)
		t.Logf("negnegp : %v\n", negnegp)
		t.Fatalf("-(-p) should be p\n")
	}
	sum = p.Add(Zero)
	if !sum.Equal(p) {
		t.Logf("p : %v\n", p)
		t.Logf("sum : %v\n", sum)
		t.Fatalf("p + 0 should be p\n")
	}
	fmt.Println("Passed TestZKCurveMethods")
}

func TestZkpCryptoStuff(t *testing.T) {
	value := big.NewInt(-100)

	testCommit, randomValue, err := PedCommit(value) // xG + rH

	if err != nil {
		t.Fatalf("%v\n", err)
	}

	value = new(big.Int).Mod(value, ZKCurve.C.Params().N)

	// vG
	ValEC := ZKCurve.G.Mult(value)
	InvValEC := ValEC.Neg() // 1/vG (acutally mod operation but whatever you get it)

	t.Logf("         vG : %v --- value : %v \n", ValEC, value)
	t.Logf("       1/vG : %v\n", InvValEC)

	temp := ValEC.Add(InvValEC)
	t.Logf("TestZkpCrypto:")
	t.Logf("Added the above: %v\n", temp)

	if !temp.Equal(Zero) {
		t.Logf("Added the above: %v", temp)
		t.Logf("The above should have been (0,0)")
		t.Fatalf("Failed Addition of inverse points failed")
	}

	testOpen := InvValEC.Add(testCommit)  // 1/vG + vG + rH ?= rH (1/vG + vG = 0, hopefully)
	RandEC := ZKCurve.H.Mult(randomValue) // rH

	if !RandEC.Equal(testOpen) {
		t.Logf("RandEC : %v\n", RandEC)
		t.Logf("testOpen : %v\n", testOpen)
		t.Fatalf("RandEC should have been equal to testOpen\n")
	}

	fmt.Println("Passed TestzkpCryptoStuff")

}

func TestZkpCryptoCommitR(t *testing.T) {

	u, err := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	testCommit := CommitR(ZKCurve.H, u)

	if !(VerifyR(testCommit, ZKCurve.H, u)) {
		t.Logf("testCommit: %v\n", testCommit)
		t.Logf("ZKCurve.H: %v, \n", ZKCurve.H)
		t.Logf("u : %v\n", u)
		t.Fatalf("testCommit should have passed verification\n")
	}

	fmt.Println("Passed TestzkpCryptoCommitR")
}

func TestPedersenCommit(t *testing.T) {

	x := big.NewInt(1000)
	badx := big.NewInt(1234)

	commit, u, err := PedCommit(x)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	commitR := PedCommitR(x, u)

	if !commit.Equal(commitR) {
		t.Logf("x : %v --- u : %v\n", x, u)
		t.Logf("commit: %v\n", commit)
		t.Logf("commitR: %v\n", commitR)
		t.Fatalf("commit and commitR should be equal")
	}

	if !Open(x, u, commit) || !Open(x, u, commitR) {
		t.Logf("x : %v --- u : %v\n", x, u)
		t.Logf("commit: %v\n", commit)
		t.Logf("commitR: %v\n", commitR)
		t.Fatalf("commit and/or commitR did not successfully open")
	}

	if Open(badx, u.Neg(u), commit) || Open(badx, u.Neg(u), commitR) {
		t.Logf("x : %v --- u : %v\n", x, u)
		t.Logf("commit: %v\n", commit)
		t.Logf("commitR: %v\n", commitR)
		t.Fatalf("commit and/or commitR should not have opened properly")
	}

	fmt.Println("Passed TestPedersenCommit")

}

// TODO: make a toooooon more test cases

type etx struct {
	CM    ECPoint
	CMTok ECPoint
	ABCP  *ABCProof
}

//TODO: make a sk-pk that is consistant accross all test cases
func TestAverages_Basic(t *testing.T) {

	// remeber to change both number here...
	numTx := 100
	numTranx := big.NewInt(100)

	totalValue := big.NewInt(0)
	totalRand := big.NewInt(0)
	txn := make([]etx, numTx)
	sk, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	PK := ZKCurve.H.Mult(sk)
	var value *big.Int
	var commRand *big.Int
	var err error

	// Generate
	for ii := 0; ii < numTx; ii++ {
		value, _ = rand.Int(rand.Reader, ZKCurve.C.Params().N)
		totalValue.Add(totalValue, value)
		txn[ii].CM, commRand, err = PedCommit(value)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		totalRand.Add(totalRand, commRand)
		txn[ii].CMTok = PK.Mult(commRand)
		txn[ii].ABCP, _ = NewABCProof(txn[ii].CM, txn[ii].CMTok, value, sk, Right)
	}

	// Purely for testing purposes, usually this is computed at the end by auditor
	// actualAverage := new(big.Int).Quo(totalValue, numTranx)

	// ========= BANK PROCESS ===========

	// To calculate average we need to first show proof of knowledge
	// of the sums of both the total value of transactions and the
	// sum of the C-bit commitments
	// This process is extactly the same process described in zkLedger
	// (Neha Nerula) paper in section 4.2

	//Need to aggregate a bunch of stuff to do equivilance proofs and what not
	totalCM := Zero
	totalCMTok := Zero
	totalC := Zero
	totalCTok := Zero

	for ii := 0; ii < numTx; ii++ {
		totalCM = txn[ii].CM.Add(totalCM)
		totalCMTok = txn[ii].CMTok.Add(totalCMTok)
		totalC = txn[ii].ABCP.C.Add(totalC)
		totalCTok = txn[ii].ABCP.CToken.Add(totalCTok)
	}

	// makes the call look cleaner
	B1 := totalC.Add(ZKCurve.G.Mult(numTranx).Neg())
	R1 := totalCTok
	B2 := ZKCurve.H
	R2 := PK

	eProofNumTx, status := NewEquivalenceProof(B1, R1, B2, R2, sk)

	if status != nil {
		proofStatus(status.(*errorProof))
		t.Logf("Average Test: equivilance proof failed to generate for numTx\n")
		t.Fatalf("Averages did not gerneate correct NUMTX equivilance proof\n")
	}

	B1 = totalCM.Add(ZKCurve.G.Mult(totalValue).Neg())
	R1 = totalCMTok

	eProofValue, status1 := NewEquivalenceProof(B1, R1, B2, R2, sk)

	if status1 != nil {
		proofStatus(status1.(*errorProof))
		t.Logf("Average Test: equivilance proof failed to generate for value sum\n")
		t.Fatalf("Averages did not gerneate correct VALUE equivilance proof\n")
	}

	// ASSUME:
	// eProofs passed to auditor
	// clear text answers of total value and total number tx passed to auditor
	// auditor WILL recalculate all the totals (totalCM, totalCMTok, etc) before doing the following
	// auditor WILL recualculate the B1's as shown above
	// auditor WILL verify eProofs and then perform the final average calcualtion, shown below
	// ======== AUDITOR PROCESS ===========

	B1 = totalC.Add(ZKCurve.G.Mult(numTranx).Neg())
	R1 = totalCTok
	B2 = ZKCurve.H
	R2 = PK

	checkTx, err := eProofNumTx.Verify(B1, R1, B2, R2)

	if err != nil {
		t.Fatalf("Error while calling equivalence proof verify: %s", err.Error())
	}

	if !checkTx {
		t.Logf("Average Test: NUMTX equivilance proof did not verify\n")
		t.Fatalf("Equivilance proof of NUMTX did not verify\n")
	}

	B1 = totalCM.Add(ZKCurve.G.Mult(totalValue).Neg())
	R1 = totalCMTok

	checkVal, err := eProofValue.Verify(B1, R1, B2, R2)

	if err != nil {
		t.Fatalf("Error while calling equivalence proof verify: %s", err.Error())
	}

	if !checkVal {
		t.Logf("Average Test: SUM equivilance proof did not verify\n")
		t.Fatalf("Equivilance proof of SUM did not verify\n")
	}

	fmt.Println("Passed TestAverages")

}

// ============== BENCHMARKS =================
func BenchmarkPedCommit(b *testing.B) {
	value, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		PedCommit(value)
	}
}

func BenchmarkPedCommitR(b *testing.B) {
	value, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	randVal, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		PedCommitR(value, randVal)
	}
}

func BenchmarkOpen(b *testing.B) {
	value, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	randVal, _ := rand.Int(rand.Reader, ZKCurve.C.Params().N)
	CM := PedCommitR(value, randVal)
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		Open(value, randVal, CM)
	}
}
