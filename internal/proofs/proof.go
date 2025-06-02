package proofs

type Proof interface {
	Generate(vcfPath string, provingKeyPath string, outputPath string) error
	Verify(verifyingKeyPath string, proofPath string) (bool, error)
}

type ChromosomeProof struct {
	Proof
}

type EyeColorProof struct {
	Proof
}

type BRCA1Proof struct {
	Proof
}

type HERC2Proof struct {
	Proof
}

const HERC2Pos uint64 = 28365618
