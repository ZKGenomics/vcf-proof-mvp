package proofs

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

// ChromosomeCircuit defines a minimal circuit that proves
// a specific chromosome exists in the genome without revealing
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

var circuit ChromosomeCircuit

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

func (p ChromosomeProof) Generate(vcfPath string, provingKeyPath string, outputPath string) error {
	fmt.Println("Reading VCF file...")
	chromosomes, err := extractChromosomeNumbers(vcfPath, 10)
	if err != nil {
		return fmt.Errorf("error reading VCF: %w", err)
	}

	if len(chromosomes) == 0 {
		return fmt.Errorf("no valid chromosome entries found in the VCF file")
	}

	fmt.Printf("Found %d chromosome entries: %v\n", len(chromosomes), chromosomes)

	// For demonstration, let's prove chromosome 22 exists in our data
	targetChromosome := 22

	fmt.Println("Compiling circuit...")
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return fmt.Errorf("circuit compilation error: %w", err)
	}

	// If proving key path is empty, set up a new one
	var pk groth16.ProvingKey
	var vk groth16.VerifyingKey

	if provingKeyPath == "" {
		fmt.Println("Setting up new proving system...")
		pk, vk, err = groth16.Setup(cs)
		if err != nil {
			return fmt.Errorf("setup error: %w", err)
		}

		// Save the proving key
		pkFile, err := os.Create(outputPath + ".pk")
		if err != nil {
			return fmt.Errorf("creating proving key file: %w", err)
		}
		defer pkFile.Close()

		_, err = pk.WriteTo(pkFile)
		if err != nil {
			return fmt.Errorf("writing proving key: %w", err)
		}

		// Save the verifying key
		vkFile, err := os.Create(outputPath + ".vk")
		if err != nil {
			return fmt.Errorf("creating verifying key file: %w", err)
		}
		defer vkFile.Close()

		_, err = vk.WriteTo(vkFile)
		if err != nil {
			return fmt.Errorf("writing verifying key: %w", err)
		}

		fmt.Printf("Keys saved to: %s.pk and %s.vk\n", outputPath, outputPath)
	} else {
		// Load the proving key
		fmt.Println("Loading existing proving key...")
		pkFile, err := os.Open(provingKeyPath)
		if err != nil {
			return fmt.Errorf("opening proving key file: %w", err)
		}
		defer pkFile.Close()

		pk = groth16.NewProvingKey(ecc.BN254)
		_, err = pk.ReadFrom(pkFile)
		if err != nil {
			return fmt.Errorf("reading proving key: %w", err)
		}
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
		return fmt.Errorf("witness creation error: %w", err)
	}

	publicWitness, err := w.Public()
	if err != nil {
		return fmt.Errorf("public witness error: %w", err)
	}

	fmt.Println("Generating proof...")
	proof, err := groth16.Prove(cs, pk, w)
	if err != nil {
		return fmt.Errorf("proving error: %w", err)
	}

	// Create output file and write data
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	// Write proof to file (with point compression)
	_, err = proof.WriteTo(outFile)
	if err != nil {
		return fmt.Errorf("writing proof: %w", err)
	}

	// Write public witness to file
	publicWitnessData, err := publicWitness.MarshalBinary()
	if err != nil {
		return fmt.Errorf("serializing public witness: %w", err)
	}

	// Write the size of the public witness data first
	witnessSize := uint32(len(publicWitnessData))
	if err := binary.Write(outFile, binary.BigEndian, witnessSize); err != nil {
		return fmt.Errorf("writing witness size: %w", err)
	}

	// Write the actual witness data
	if _, err := outFile.Write(publicWitnessData); err != nil {
		return fmt.Errorf("writing public witness: %w", err)
	}

	fmt.Println("✅ Proof successfully generated!")
	fmt.Printf("We have proven knowledge of chromosome %d's presence in the genomic data\n", targetChromosome)
	fmt.Println("without revealing which entries contain this chromosome or any other genomic information.")
	fmt.Printf("Proof saved to: %s\n", outputPath)

	return nil
}

func (*ChromosomeProof) Verify(verifyingKeyPath string, proofPath string) (bool, error) {
	fmt.Println("Compiling circuit...")
	_, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		return false, fmt.Errorf("compiling circuit: %w", err)
	}

	// Load the verifying key
	vkFile, err := os.Open(verifyingKeyPath)
	if err != nil {
		return false, fmt.Errorf("opening verifying key file: %w", err)
	}
	defer vkFile.Close()

	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(vkFile)
	if err != nil {
		return false, fmt.Errorf("reading verifying key: %w", err)
	}

	// Open proof file
	proofFile, err := os.Open(proofPath)
	if err != nil {
		return false, fmt.Errorf("opening proof file: %w", err)
	}
	defer proofFile.Close()

	// Read proof
	proof := groth16.NewProof(ecc.BN254)
	_, err = proof.ReadFrom(proofFile)
	if err != nil {
		return false, fmt.Errorf("reading proof: %w", err)
	}

	// Read public witness size
	var witnessSize uint32
	if err := binary.Read(proofFile, binary.BigEndian, &witnessSize); err != nil {
		return false, fmt.Errorf("reading witness size: %w", err)
	}

	// Read public witness data
	publicWitnessData := make([]byte, witnessSize)
	if _, err := io.ReadFull(proofFile, publicWitnessData); err != nil {
		return false, fmt.Errorf("reading public witness data: %w", err)
	}

	// Create public witness
	publicWitness, err := witness.New(ecc.BN254.ScalarField())
	if err != nil {
		return false, fmt.Errorf("creating witness: %w", err)
	}

	if err := publicWitness.UnmarshalBinary(publicWitnessData); err != nil {
		return false, fmt.Errorf("unmarshalling public witness: %w", err)
	}

	fmt.Println("Verifying proof...")
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return false, fmt.Errorf("verification failed: %w", err)
	}

	fmt.Println("✅ Proof successfully verified!")
	return true, nil
}
