package proofs

type Proof interface {
	Generate(vcfPath string, provingKeyPath string, outputPath string) error
	Verify(verifyingKeyPath string, proofPath string) (bool, error)
}

type ChromosomeProof struct{}
type EyeColorProof struct{}
type BRCA1Proof struct{}
