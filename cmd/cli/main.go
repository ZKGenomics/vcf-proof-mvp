package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zkgenomics/vcf-proof-mvp/internal/proofs"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		handleGenerate(os.Args[2:])
	case "verify":
		handleVerify(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleGenerate(args []string) {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	proofType := generateCmd.String("type", "", "Type of proof to generate (chromosome, eyecolor, brca1)")
	vcfPath := generateCmd.String("vcf", "", "Path to VCF file")
	outputPath := generateCmd.String("output", "", "Output path for the proof file")
	provingKeyPath := generateCmd.String("proving-key", "", "Path to existing proving key (optional)")
	outputDir := generateCmd.String("output-dir", "output", "Output directory for proof files")

	generateCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s generate [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate a zero-knowledge proof from genomic data\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		generateCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s generate -type chromosome -vcf data/genome.vcf -output-dir output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s generate -type eyecolor -vcf data/genome.vcf -output my_proof.bin\n", os.Args[0])
	}

	generateCmd.Parse(args)

	if *proofType == "" || *vcfPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -type and -vcf are required\n\n")
		generateCmd.Usage()
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Set default output path if not specified
	if *outputPath == "" {
		*outputPath = filepath.Join(*outputDir, *proofType+"_proof.bin")
	}

	proof, err := createProof(*proofType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating %s proof...\n", *proofType)
	fmt.Printf("VCF file: %s\n", *vcfPath)
	fmt.Printf("Output path: %s\n", *outputPath)
	if *provingKeyPath != "" {
		fmt.Printf("Using proving key: %s\n", *provingKeyPath)
	}

	if err := proof.Generate(*vcfPath, *provingKeyPath, *outputPath); err != nil {
		fmt.Printf("Error generating proof: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s proof at: %s\n", *proofType, *outputPath)
}

func handleVerify(args []string) {
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	proofType := verifyCmd.String("type", "", "Type of proof to verify (chromosome, eyecolor, brca1)")
	proofPath := verifyCmd.String("proof", "", "Path to proof file")
	verifyingKeyPath := verifyCmd.String("verifying-key", "", "Path to verifying key file")

	verifyCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s verify [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Verify a zero-knowledge proof\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		verifyCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s verify -type chromosome -proof output/chromosome_proof.bin -verifying-key output/chromosome_proof.bin.vk\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s verify -type eyecolor -proof my_proof.bin -verifying-key my_proof.bin.vk\n", os.Args[0])
	}

	verifyCmd.Parse(args)

	if *proofType == "" || *proofPath == "" {
		// Try to auto-detect verifying key path if not provided
		if *verifyingKeyPath == "" && *proofPath != "" {
			*verifyingKeyPath = *proofPath + ".vk"
		}

		if *proofType == "" || *proofPath == "" {
			fmt.Fprintf(os.Stderr, "Error: -type and -proof are required\n\n")
			verifyCmd.Usage()
			os.Exit(1)
		}
	}

	// Auto-detect verifying key path if not provided
	if *verifyingKeyPath == "" {
		*verifyingKeyPath = *proofPath + ".vk"
	}

	proof, err := createProof(*proofType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Verifying %s proof...\n", *proofType)
	fmt.Printf("Proof file: %s\n", *proofPath)
	fmt.Printf("Verifying key: %s\n", *verifyingKeyPath)

	verified, err := proof.Verify(*verifyingKeyPath, *proofPath)
	if err != nil {
		fmt.Printf("Error verifying proof: %v\n", err)
		os.Exit(1)
	}

	if verified {
		fmt.Printf("✓ %s proof verified successfully!\n", strings.Title(*proofType))
	} else {
		fmt.Printf("✗ %s proof verification failed!\n", strings.Title(*proofType))
		os.Exit(1)
	}
}

func createProof(proofType string) (proofs.Proof, error) {
	switch strings.ToLower(proofType) {
	case "chromosome":
		return &proofs.ChromosomeProof{}, nil
	case "eyecolor":
		return &proofs.EyeColorProof{}, nil
	case "brca1":
		return &proofs.BRCA1Proof{}, nil
	case "herc2":
		return &proofs.HERC2Proof{}, nil
	default:
		return nil, fmt.Errorf("unknown proof type: %s. Supported types: chromosome, eyecolor, brca1", proofType)
	}
}

func printUsage() {
	fmt.Printf("VCF Proof CLI - Generate and verify zero-knowledge proofs for genomic data\n\n")
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Printf("Commands:\n")
	fmt.Printf("  generate    Generate a zero-knowledge proof from VCF data\n")
	fmt.Printf("  verify      Verify a zero-knowledge proof\n")
	fmt.Printf("  help        Show this help message\n\n")
	fmt.Printf("Supported proof types:\n")
	fmt.Printf("  chromosome  Chromosome-based genomic proof\n")
	fmt.Printf("  eyecolor    Eye color trait proof\n")
	fmt.Printf("  brca1       BRCA1 gene mutation proof\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  %s generate -type chromosome -vcf data/genome.vcf\n", os.Args[0])
	fmt.Printf("  %s verify -type chromosome -proof output/chromosome_proof.bin\n", os.Args[0])
	fmt.Printf("  %s help\n\n", os.Args[0])
	fmt.Printf("For more detailed help on a specific command, use:\n")
	fmt.Printf("  %s <command> -h\n", os.Args[0])
}
