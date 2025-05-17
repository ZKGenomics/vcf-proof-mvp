# Genome Encryption with Zero-Knowledge Proofs

This project demonstrates how to create zero-knowledge proofs for genomic data stored in VCF (Variant Call Format) files.

## Overview

Zero-knowledge proofs allow one party (the prover) to prove to another party (the verifier) that they know a value without revealing any information about the value itself. In the context of genomic data, this enables privacy-preserving validation of genetic information.

This implementation provides a simple proof-of-concept that demonstrates:

1. Reading chromosome data from VCF files
2. Creating a zero-knowledge circuit that proves a specific chromosome exists in the dataset
3. Generating and verifying proofs without revealing which entries contain the chromosome

## Background

### VCF File Format

The Variant Call Format (VCF) is a text file format used in bioinformatics for storing gene sequence variations. Our implementation focuses on the chromosome field, which identifies the chromosome where a genetic variant is located.

### Zero-Knowledge Proofs

Zero-knowledge proofs allow proving knowledge of information without revealing the information itself. This project uses the gnark library to create zk-SNARKs (Zero-Knowledge Succinct Non-Interactive Argument of Knowledge).

## Implementation

The main components of this project are:

1. **VCF Parser**: Extracts chromosome numbers from VCF files
2. **ZK Circuit**: Defines a circuit that proves a specific chromosome exists in the dataset
3. **Prover**: Generates a proof that the target chromosome exists
4. **Verifier**: Verifies the proof without learning which specific entries contain the chromosome

### Circuit Design

The circuit takes:
- A public input: the target chromosome we want to prove exists
- Private inputs: the actual chromosome values from the VCF file

The circuit checks if any of the private chromosome values match the target and generates a proof if a match is found.

## Usage

### Prerequisites

- Go 1.18 or higher
- VCF file with genomic data

### Running the Example

1. Clone the repository
2. Ensure you have a VCF file in the `data` directory or run the example to create a sample file
3. Run the program:

```bash
go run main.go
```

The program will:
1. Read chromosome data from your VCF file
2. Compile a zero-knowledge circuit
3. Generate a proof that a specific chromosome exists in your data
4. Verify the proof

## Future Enhancements

This proof-of-concept could be extended in several ways:

1. **More Complex Proofs**: Prove properties of the genome without revealing the genome itself
2. **Privacy-Preserving Genomic Analysis**: Implement ZK proofs for common genomic analyses
3. **Integration with Genomic Databases**: Create privacy-preserving queries to genomic databases
4. **Optimizations**: Improve performance for large genomic datasets
5. **Advanced Properties**: Prove more complex properties like ancestry, disease risk, etc.

## References

- [gnark Library](https://github.com/ConsenSys/gnark)
- [VCF File Format Specification](https://samtools.github.io/hts-specs/VCFv4.2.pdf)
- [Zero-Knowledge Proofs: An Illustrated Primer](https://blog.cryptographyengineering.com/2014/11/27/zero-knowledge-proofs-illustrated-primer/)

## License

This project is licensed under the MIT License - see the LICENSE file for details.