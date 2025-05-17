package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// ChromosomeCircuit defines a minimal circuit that proves
// a specific chromosome exists in the dataset without revealing
// other genomic information
type ChromosomeCircuit struct {
	// Public input - the chromosome number we want to prove exists
	TargetChromosome frontend.Variable `gnark:",public"`

	// Private inputs - chromosome data from the VCF file
	// We'll keep a fixed number for simplicity
	Chromosome1 frontend.Variable
	Chromosome2 frontend.Variable
	Chromosome3 frontend.Variable
	Chromosome4 frontend.Variable
	Chromosome5 frontend.Variable
}

// Define declares the circuit constraints
func (circuit *ChromosomeCircuit) Define(api frontend.API) error {
	// We want to prove that TargetChromosome exists in our dataset
	// without revealing which position it was found at

	// Check if chromosomes match the target by computing their differences
	diff1 := api.Sub(circuit.Chromosome1, circuit.TargetChromosome)
	diff2 := api.Sub(circuit.Chromosome2, circuit.TargetChromosome)
	diff3 := api.Sub(circuit.Chromosome3, circuit.TargetChromosome)
	diff4 := api.Sub(circuit.Chromosome4, circuit.TargetChromosome)
	diff5 := api.Sub(circuit.Chromosome5, circuit.TargetChromosome)

	// We can also compute the sum of squares of differences
	// This would be zero only if all differences are zero
	_ = api.Add(
		api.Mul(diff1, diff1),
		api.Mul(diff2, diff2),
		api.Mul(diff3, diff3),
		api.Mul(diff4, diff4),
		api.Mul(diff5, diff5),
	)

	// If all diffs are non-zero, their product will be non-zero
	product := api.Mul(diff1, diff2, diff3, diff4, diff5)
	api.AssertIsEqual(product, 0)

	return nil
}

func extractChromosomeNumbers(vcfPath string, maxCount int) ([]int, error) {
	f, err := os.Open(vcfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		return nil, err
	}

	chromosomes := make([]int, 0, maxCount)
	count := 0

	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}

		chrStr := variant.Chromosome
		chrStr = strings.TrimPrefix(chrStr, "chr")

		chrNum, err := strconv.Atoi(chrStr)
		if err == nil {
			chromosomes = append(chromosomes, chrNum)
			count++
		}

		if count >= maxCount {
			break
		}
	}

	return chromosomes, nil
}

func main() {
	vcfPath := "data/genome_example.vcf"

	fmt.Println("Reading VCF file...")
	chromosomes, err := extractChromosomeNumbers(vcfPath, 10)
	if err != nil {
		fmt.Printf("Error reading VCF: %v\n", err)
		os.Exit(1)
	}

	if len(chromosomes) == 0 {
		fmt.Println("No valid chromosome entries found in the VCF file.")
		os.Exit(1)
	}

	fmt.Printf("Found %d chromosome entries: %v\n", len(chromosomes), chromosomes)

	// For demonstration, let's prove chromosome 22 exists in our data
	targetChromosome := 22

	var circuit ChromosomeCircuit

	fmt.Println("Compiling circuit...")
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

	fmt.Println("Creating witness...")

	// Pad chromosomes to 5 items (our fixed circuit size)
	paddedChromosomes := make([]int, 5)
	for i := 0; i < 5; i++ {
		if i < len(chromosomes) {
			paddedChromosomes[i] = chromosomes[i]
		} else {
			paddedChromosomes[i] = 0 // Default value for padding
		}
	}

	witness := &ChromosomeCircuit{
		TargetChromosome: targetChromosome,
		Chromosome1:      paddedChromosomes[0],
		Chromosome2:      paddedChromosomes[1],
		Chromosome3:      paddedChromosomes[2],
		Chromosome4:      paddedChromosomes[3],
		Chromosome5:      paddedChromosomes[4],
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

	fmt.Println("✅ Proof successfully generated and verified!")
	fmt.Printf("We have proven knowledge of chromosome %d's presence in the genomic data\n", targetChromosome)
	fmt.Println("without revealing which entries contain this chromosome or any other genomic information.")
}
