package proofs

import (
	"fmt"
	"os"

	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark/frontend"
)

type BRCA1Circuit struct {
	ClaimedColor frontend.Variable `gnark:",public"`
	Genotype     frontend.Variable
}

func (c *BRCA1Circuit) Define(api frontend.API) error {
	api.Sub(c.ClaimedColor, c.Genotype)

	return nil
}

func (p *BRCA1Proof) Generate(vcfPath string, provingKeyPath string, outputPath string) error {
	f, err := os.Open(vcfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		return err
	}

	fmt.Println("searching for BRCA1 trait...")
	for {
		variant := rdr.Read()
		if variant == nil {
			fmt.Println("Could not find position")
			break
		}

		pos := variant.Pos

		if pos == 41276045 {
			fmt.Println("Found position.")
			fmt.Printf("Variant: Chromosome: %s, Reference: %s, Alternate: %s", variant.Chromosome, variant.Reference, variant.Alternate)
			break
		}
	}

	return nil
}

func (p *BRCA1Proof) Verify(verifyingKeyPath string, proofPath string) (bool, error) {
	return true, nil
}
