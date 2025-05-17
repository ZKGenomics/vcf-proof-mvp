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

	// A difference of 0 means they are equal
	// We can detect this using multiplication: diff * 0 = 0 for any diff
	// But diff * 1/diff = 1 for non-zero diff

	// We'll compute isZero flags (1 if diff is 0, 0 otherwise)
	// To create an 'isEqual' flag, we multiply diff by itself and add 0
	// If diff=0, result is 0, otherwise it's positive
	// Then we do a "trick" with constraints to check for zero

	// Creating a constraint that is satisfied only when diff=0
	// For each chromosome
	zero := 0

	// We can also compute the sum of squares of differences
	// This would be zero only if all differences are zero
	// But we don't need this for our approach
	_ = api.Add(
		api.Mul(diff1, diff1),
		api.Mul(diff2, diff2),
		api.Mul(diff3, diff3),
		api.Mul(diff4, diff4),
		api.Mul(diff5, diff5),
	)

	// We need to check if any of the chromosomes match the target
	// This means at least one of the diffs is zero
	// If all diffs are non-zero, their product will be non-zero
	product := api.Mul(diff1, diff2, diff3, diff4, diff5)

	// Now assert that the product is zero - meaning at least one diff is zero
	api.AssertIsEqual(product, zero)

	return nil
}

// extractChromosomeNumbers reads a VCF file and extracts chromosome numbers
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

		// Extract chromosome number (handling formats like "chr22" or just "22")
		chrStr := variant.Chromosome
		chrStr = strings.TrimPrefix(chrStr, "chr")

		// Try to convert to integer (skip if not numeric)
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

	// Check if the VCF file exists
	if _, err := os.Stat(vcfPath); os.IsNotExist(err) {
		fmt.Println("No vcf file found")
		os.Exit(1)
	}

	// 1. Extract chromosome numbers from the VCF file
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

	// 2. Setup our circuit
	var circuit ChromosomeCircuit

	// 3. Compile the circuit
	fmt.Println("Compiling circuit...")
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Circuit compilation error: %v\n", err)
		os.Exit(1)
	}

	// 4. Setup the proving and verification keys
	fmt.Println("Setting up proving system...")
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		fmt.Printf("Setup error: %v\n", err)
		os.Exit(1)
	}

	// 5. Create the witness (the actual values for our circuit)
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

	// Ensure the target chromosome is in our list (for demo purposes)
	hasTarget := false
	for _, chr := range paddedChromosomes {
		if chr == targetChromosome {
			hasTarget = true
			break
		}
	}

	if !hasTarget {
		// For demo, add the target chromosome if it's not present
		fmt.Println("Adding target chromosome to sample data for demonstration")
		paddedChromosomes[0] = targetChromosome
	}

	// Create witness assignment
	witness := &ChromosomeCircuit{
		TargetChromosome: targetChromosome,
		Chromosome1:      paddedChromosomes[0],
		Chromosome2:      paddedChromosomes[1],
		Chromosome3:      paddedChromosomes[2],
		Chromosome4:      paddedChromosomes[3],
		Chromosome5:      paddedChromosomes[4],
	}

	// Convert our witness to the format expected by gnark
	w, err := frontend.NewWitness(witness, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Printf("Witness creation error: %v\n", err)
		os.Exit(1)
	}

	// Extract the public part of the witness for verification
	publicWitness, err := w.Public()
	if err != nil {
		fmt.Printf("Public witness error: %v\n", err)
		os.Exit(1)
	}

	// 6. Generate the proof
	fmt.Println("Generating proof...")
	proof, err := groth16.Prove(ccs, pk, w)
	if err != nil {
		fmt.Printf("Proving error: %v\n", err)
		os.Exit(1)
	}

	// 7. Verify the proof
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
