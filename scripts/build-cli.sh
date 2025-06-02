#!/bin/bash

set -e

echo "Building VCF Proof CLI tools..."

# Navigate to the project root
cd "$(dirname "$0")"

# Build the main CLI binary
echo "Compiling CLI from cmd/cli..."
go build -o ../bin/vcf-proof-cli ../cmd/cli

# Build the trait checker binary
echo "Compiling trait checker from cmd/trait-checker..."
go build -o ../bin/trait-checker ../cmd/trait-checker

# Make the binaries executable
chmod +x ../bin/vcf-proof-cli
chmod +x ../bin/trait-checker

echo "✓ CLI tools built successfully!"
echo "Binary locations:"
echo "  - VCF Proof CLI: ./bin/vcf-proof-cli"
echo "  - Trait Checker: ./bin/trait-checker"
echo ""
echo "Usage examples:"
echo "  # Generate and verify proofs:"
echo "  ./bin/vcf-proof-cli generate -type chromosome -vcf data/genome_example.vcf"
echo "  ./bin/vcf-proof-cli verify -type chromosome -proof output/chromosome_proof.bin"
echo "  ./bin/vcf-proof-cli help"
echo ""
echo "  # Check VCF for trait variants:"
echo "  ./bin/trait-checker -vcf data/genome_example.vcf"
echo "  ./bin/trait-checker -vcf data/genome_example.vcf -found-only"
echo "  ./bin/trait-checker -vcf data/genome_example.vcf -summary"
echo ""
echo "To install globally, you can copy the binaries to your PATH:"
echo "  sudo cp bin/vcf-proof-cli /usr/local/bin/"
echo "  sudo cp bin/trait-checker /usr/local/bin/"
