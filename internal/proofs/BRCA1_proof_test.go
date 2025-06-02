package proofs

import (
	"os"
	"testing"
)

func TestBRCA1Proof_Generate(t *testing.T) {
	// Create a temporary VCF file for testing
	vcfContent := `##fileformat=VCFv4.2
##INFO=<ID=DP,Number=1,Type=Integer,Description="Approximate read depth">
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO
17	41276045	.	A	G	60	PASS	DP=30
`

	tmpFile, err := os.CreateTemp("", "test*.vcf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(vcfContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	proof := &BRCA1Proof{}
	err = proof.Generate(tmpFile.Name(), "", "")
	if err != nil {
		t.Errorf("Generate should not return error: %v", err)
	}
}

func TestBRCA1Proof_GenerateWithMissingPosition(t *testing.T) {
	// Create a temporary VCF file without the target position
	vcfContent := `##fileformat=VCFv4.2
##INFO=<ID=DP,Number=1,Type=Integer,Description="Approximate read depth">
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO
17	12345678	.	A	G	60	PASS	DP=30
`

	tmpFile, err := os.CreateTemp("", "test*.vcf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(vcfContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	proof := &BRCA1Proof{}
	err = proof.Generate(tmpFile.Name(), "", "")
	if err != nil {
		t.Errorf("Generate should not return error: %v", err)
	}
}

func TestBRCA1Proof_Verify(t *testing.T) {
	proof := &BRCA1Proof{}
	result, err := proof.Verify("", "")
	if err != nil {
		t.Errorf("Verify should not return error: %v", err)
	}
	if !result {
		t.Errorf("Verify should return true")
	}
}
