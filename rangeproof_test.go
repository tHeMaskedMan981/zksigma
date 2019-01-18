package zksigma

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

// Copy-pasted from origianl apl implementation by Willy (github.com/wrv)
func TestRangeProver_Verify(t *testing.T) {

	if !*RANGE {
		fmt.Println("Skipped TestRangeProver_Verify - use -range flag to run")
		t.Skip("Skipped TestRangeProver_Verify")
	}

	value, _ := rand.Int(rand.Reader, big.NewInt(1099511627775))
	proof, rp := NewRangeProof(value)
	comm := PedCommitR(value, rp)
	if !comm.Equal(proof.ProofAggregate) {
		t.Error("Error computing the randomnesses used -- commitments did not check out when supposed to")
	} else {
		ok, err := proof.Verify(comm)
		if !ok {
			t.Errorf("** Range proof failed: %s", err)
		} else {
			fmt.Println("Passed TestRangeProver_Verify")
		}
	}
}

func TestRangeProverSerialization(t *testing.T) {

	if !*RANGE {
		fmt.Println("Skipped TestRangeProverSerialization - use -range flag to run")
		t.Skip("Skipped TestRangeProverSerialization")
	}

	value, _ := rand.Int(rand.Reader, big.NewInt(1099511627775))
	proof, rp := NewRangeProof(value)
	proof, err := NewRangeProofFromBytes(proof.Bytes())
	if err != nil {
		t.Fatalf("TestRangeProverSerialization failed to deserialize\n")
	}
	comm := PedCommitR(value, rp)
	if !comm.Equal(proof.ProofAggregate) {
		t.Error("Error computing the randomnesses used -- commitments did not check out when supposed to")
	} else {
		ok, err := proof.Verify(comm)
		if !ok {
			t.Errorf("** Range proof failed: %s", err)
		} else {
			fmt.Println("Passed TestRangeProverSerialization")
		}
	}
}

func TestOutOfRangeRangeProver_Verify(t *testing.T) {

	if !*RANGE {
		fmt.Println("Skipped TestOutOfRangeRangeProver_Verify - use -range flag to run")
		t.Skip("Skipped TestOutOfRangeRangeProver_Verify")
	}

	fmt.Println("TestOutOfRangeRangeProver_Verify: There should be an error message following this")

	min := new(big.Int).Exp(new(big.Int).SetInt64(2), new(big.Int).SetInt64(64), nil)

	value, err := rand.Int(rand.Reader, new(big.Int).Add(new(big.Int).Sub(ZKCurve.C.Params().N, min), min)) // want to make sure it's out of range
	if err != nil {
		t.Error(err)
	}

	proof, rp := NewRangeProof(value)
	if proof != nil || rp != nil {
		t.Error("Error computing the range proof; shouldn't work")
	} else {
		fmt.Println("Passed TestOutOfRangeProver_Verify")
	}
}
