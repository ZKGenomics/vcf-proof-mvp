#!/bin/bash

set -e

echo "Building VCF Proof CLI..."

# Navigate to the project root
cd "$(dirname "$0")"

# Build the CLI binary
echo "Compiling CLI from cmd/cli..."
go build -o ../bin/vcf-proof-cli ../cmd/cli

# Make the binary executable
chmod +x ../bin/vcf-proof-cli

echo "✓ CLI built successfully!"
echo "Binary location: ./bin/vcf-proof-cli"
echo ""
echo "Usage examples:"
echo "  ./bin/vcf-proof-cli generate -type chromosome -vcf data/genome_example.vcf"
echo "  ./bin/vcf-proof-cli verify -type chromosome -proof output/chromosome_proof.bin"
echo "  ./bin/vcf-proof-cli help"
echo ""
echo "To install globally, you can copy the binary to your PATH:"
echo "  sudo cp bin/vcf-proof-cli /usr/local/bin/"
