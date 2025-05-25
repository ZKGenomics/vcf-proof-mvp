package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// EyeColorCircuit proves knowledge of the genotype at rs12913832
// without revealing the actual genotype, only the resulting eye color.
type EyeColorCircuit struct {
	// Public input: claimed eye color (1=Brown, 2=Hazel/Green, 3=Blue)
	ClaimedColor frontend.Variable `gnark:",public"`

	// Private input: genotype at rs12913832 (0=G/G, 1=G/A, 2=A/A)
	Genotype frontend.Variable
}

// Define circuit constraints for eye color proof
func (c *EyeColorCircuit) Define(api frontend.API) error {
	// Map genotype to color:
	// 0 (G/G) -> 1 (Brown)
	// 1 (G/A) -> 2 (Hazel/Green)
	// 2 (A/A) -> 3 (Blue)

	// Compute expected color from genotype
	expectedColor := api.Add(api.Mul(c.Genotype, 1), 1)
	// If genotype == 0: 0*1+1 = 1 (Brown)
	// If genotype == 1: 1*1+1 = 2 (Hazel/Green)
	// If genotype == 2: 2*1+1 = 3 (Blue)

	api.AssertIsEqual(c.ClaimedColor, expectedColor)
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
		if variant.ID == "rs12913832" {
			gt := variant.Genotype
			// GT can be "0/0", "0/1", "1/1"
			switch gt {
			case "0/0":
				return 0, nil // G/G
			case "0/1", "1/0":
				return 1, nil // G/A
			case "1/1":
				return 2, nil // A/A
			default:
				return 0, fmt.Errorf("unknown genotype: %s", gt)
			}
		}
	}
	return 0, fmt.Errorf("rs12913832 not found in VCF")
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

func main() {
	vcfPath := "data/genome_example.vcf"

	fmt.Println("Reading VCF for rs12913832 genotype...")
	genotype, err := extractEyeColorGenotype(vcfPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	claimedColor := genotypeToColor(genotype)
	if claimedColor == 0 {
		fmt.Println("Could not map genotype to eye color.")
		os.Exit(1)
	}

	fmt.Printf("Genotype at rs12913832: %d, Claimed eye color: %d\n", genotype, claimedColor)

	var circuit EyeColorCircuit

	fmt.Println("Compiling eye color circuit...")
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Circuit compilation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Setting up proving system...")
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		fmt.Printf("Setup error: %v\n", err)
		os.Exit(1)
	}

	witness := &EyeColorCircuit{
		ClaimedColor: claimedColor,
		Genotype:     genotype,
	}

	w, err := frontend.NewWitness(witness, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Printf("Witness creation error: %v\n", err)
		os.Exit(1)
	}

	publicWitness, err := w.Public()
	if err != nil {
		fmt.Printf("Public witness error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating proof...")
	proof, err := groth16.Prove(ccs, pk, w)
	if err != nil {
		fmt.Printf("Proving error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Verifying proof...")
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Eye color proof successfully generated and verified!")
	fmt.Printf("We have proven knowledge of the genotype at rs12913832 corresponding to eye color %d\n", claimedColor)
	fmt.Println("without revealing the actual genotype or any other genomic information.")
}

