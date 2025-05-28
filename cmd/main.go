package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zkgenomics/vcf-proof-mvp/internal/proofs"
)

func main() {
	fmt.Println("Starting VCF proof MVP...")
	fmt.Println("This program demonstrates zero-knowledge proofs for genomic data")
	fmt.Println("The Generate function creates proofs and serializes them to files")
	fmt.Println("The Verify function reads serialized proofs and verifies them")

	// Setup paths
	vcfPath := "data/genome_example.vcf"
	outputDir := "output"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Chromosome proof paths
	chromosomeProofPath := filepath.Join(outputDir, "chromosome_proof.bin")
	//chromosomePkPath := chromosomeProofPath + ".pk"  // Proving key has .pk suffix
	chromosomeVkPath := chromosomeProofPath + ".vk" // Verifying key has .vk suffix

	// Generate chromosome proof
	// - First parameter: VCF file to read genomic data from
	// - Second parameter: Path to existing proving key (empty to generate a new one)
	// - Third parameter: Output path where the proof will be saved
	// The Generate function will:
	// 1. Read genomic data from the VCF file
	// 2. Set up the proving system (or load existing keys)
	// 3. Create a witness with the circuit constraints and private/public inputs
	// 4. Generate a zk-SNARK proof
	// 5. Save the proof and public witness to the output file
	// 6. If generating new keys, save the proving key (.pk) and verifying key (.vk) files
	var chromosomeProof proofs.ChromosomeProof
	if err := chromosomeProof.Generate(vcfPath, "", chromosomeProofPath); err != nil {
		fmt.Printf("Error generating chromosome proof: %v\n", err)
		os.Exit(1)
	}

	// Verify chromosome proof
	// - First parameter: Path to the verifying key file
	// - Second parameter: Path to the proof file containing the proof and public witness
	// The Verify function will:
	// 1. Load the verifying key
	// 2. Read the proof and public witness from the proof file
	// 3. Verify that the proof is valid using Groth16.Verify
	// 4. Return true if verification succeeded, false otherwise
	verified, err := chromosomeProof.Verify(chromosomeVkPath, chromosomeProofPath)
	if err != nil {
		fmt.Printf("Error verifying chromosome proof: %v\n", err)
		os.Exit(1)
	}

	if verified {
		fmt.Println("Chromosome proof verified successfully!")
	} else {
		fmt.Println("Chromosome proof verification failed!")
	}
}
