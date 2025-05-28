package proofs

import (
	"fmt"
	"os"

	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark/frontend"
)

type EyeColorCircuit struct {
	ClaimedColor frontend.Variable `gnark:",public"`
	Genotype     frontend.Variable
}

func (c *EyeColorCircuit) Define(api frontend.API) error {
	api.Sub(c.ClaimedColor, c.Genotype)

	return nil
}

// Parse rs12913832 genotype from VCF and map to integer
func extractEyeColorGenotype(vcfPath string) (int, error) {
	f, err := os.Open(vcfPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		return 0, err
	}

	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}
		if variant.Pos == 396321 {
			fmt.Println(fmt.Sprintf("Found eye color mutation at variant: %s", variant.Chromosome))
			return 1, nil // Simplified for demonstration
		}
	}
	return 0, fmt.Errorf("not found in VCF")
}

// Map genotype integer to color integer
func genotypeToColor(genotype int) int {
	switch genotype {
	case 0:
		return 1 // Brown
	case 1:
		return 2 // Hazel/Green
	case 2:
		return 3 // Blue
	default:
		return 0
	}
}

func (p EyeColorProof) Generate(vcfPath string, provingKeyPath string, outputPath string) error {
	return nil
}

func (p EyeColorProof) Verify(verifyingKeyPath string, proofPath string) (bool, error) {
	return true, nil
}
