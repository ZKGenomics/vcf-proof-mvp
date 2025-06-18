package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark-crypto/ecc"
)

// Define a simple circuit: prove knowledge of x such that x^3 + x + 5 == y
type CubicCircuit struct {
	X frontend.Variable `gnark:",secret"` // Private input
	Y frontend.Variable `gnark:",public"` // Public input
}

func (circuit *CubicCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	api.AssertIsEqual(circuit.Y, api.Add(x3, circuit.X, 5))
	return nil
}

// ProofSystem manages the entire zk-SNARK workflow
type ProofSystem struct {
	r1cs    constraint.ConstraintSystem
	pk      groth16.ProvingKey
	vk      groth16.VerifyingKey
	proof   groth16.Proof
	witness map[string]interface{}
}

// NewProofSystem initializes a new proof system with the cubic circuit
func NewProofSystem() (*ProofSystem, error) {
	ps := &ProofSystem{witness: make(map[string]interface{})}

	// Compile the circuit using gnark v0.12.0 API
	// frontend.Compile now requires: field, builder, circuit
	var circuit CubicCircuit
	var err error
	field := ecc.BN254.ScalarField()
	ps.r1cs, err = frontend.Compile(field, r1cs.NewBuilder, &circuit)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Groth16 setup
	ps.pk, ps.vk, err = groth16.Setup(ps.r1cs)
	if err != nil {
		return nil, fmt.Errorf("failed to perform trusted setup: %w", err)
	}

	return ps, nil
}

// GenerateProof creates a proof for given inputs
func (ps *ProofSystem) GenerateProof(x, y int) error {
	// Create witness
	assignment := CubicCircuit{
		X: x,
		Y: y,
	}
	ps.witness["x"] = x
	ps.witness["y"] = y

	// NewWitness API in v0.12.0 requires: assignment, field, options
	field := ecc.BN254.ScalarField()
	witness, err := frontend.NewWitness(&assignment, field, frontend.PublicOnly())
	if err != nil {
		return fmt.Errorf("failed to create witness: %w", err)
	}

	// Full witness (public + secret) - no options means full witness
	fullWitness, err := frontend.NewWitness(&assignment, field)
	if err != nil {
		return fmt.Errorf("failed to create full witness: %w", err)
	}

	// Generate proof
	ps.proof, err = groth16.Prove(ps.r1cs, ps.pk, fullWitness)
	if err != nil {
		return fmt.Errorf("failed to generate proof: %w", err)
	}

	// Verify proof immediately to ensure correctness
	err = groth16.Verify(ps.proof, ps.vk, witness)
	if err != nil {
		return fmt.Errorf("failed to verify proof: %w", err)
	}

	return nil
}

// ProofQuery provides a DSL for interacting with proofs
type ProofQuery struct {
	proof       groth16.Proof
	vk          groth16.VerifyingKey
	witness     map[string]interface{}
	constraints constraint.ConstraintSystem
}

// NewProofQuery creates a new query interface for a proof
func (ps *ProofSystem) NewProofQuery() *ProofQuery {
	return &ProofQuery{
		proof:       ps.proof,
		vk:          ps.vk,
		witness:     ps.witness,
		constraints: ps.r1cs,
	}
}

// ExecuteQuery runs a query against the proof
func (pq *ProofQuery) ExecuteQuery(query string) (interface{}, error) {
	switch query {
	case "proof.size":
		var buf bytes.Buffer
		_, err := pq.proof.WriteTo(&buf)
		if err != nil {
			return nil, err
		}
		return buf.Len(), nil

	case "proof.verify":
		// Create public witness for verification
		field := ecc.BN254.ScalarField()
		witness, err := frontend.NewWitness(&CubicCircuit{Y: pq.witness["y"]}, field, frontend.PublicOnly())
		if err != nil {
			return nil, err
		}
		err = groth16.Verify(pq.proof, pq.vk, witness)
		return err == nil, err

	case "circuit.constraints":
		return pq.constraints.GetNbConstraints(), nil

	case "witness.public":
		return pq.witness["y"], nil

	case "witness.private":
		return "hidden (private witness)", nil // In real use, wouldn't expose this

	default:
		return nil, fmt.Errorf("unknown query: %s", query)
	}
}

// SaveToFiles persists the proving and verifying keys
func (ps *ProofSystem) SaveToFiles(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Save proving key
	pkFile, err := os.Create(dir + "/proving.key")
	if err != nil {
		return err
	}
	defer pkFile.Close()
	if _, err := ps.pk.WriteTo(pkFile); err != nil {
		return err
	}

	// Save verifying key
	vkFile, err := os.Create(dir + "/verifying.key")
	if err != nil {
		return err
	}
	defer vkFile.Close()
	if _, err := ps.vk.WriteTo(vkFile); err != nil {
		return err
	}

	return nil
}

// LoadFromFiles loads the proving and verifying keys
func (ps *ProofSystem) LoadFromFiles(dir string) error {
	// Load proving key - API changed to use ecc.ID instead of constraint system
	pkFile, err := os.Open(dir + "/proving.key")
	if err != nil {
		return err
	}
	defer pkFile.Close()
	ps.pk = groth16.NewProvingKey(ecc.BN254)
	if _, err := ps.pk.ReadFrom(pkFile); err != nil {
		return err
	}

	// Load verifying key - API changed to use ecc.ID instead of constraint system
	vkFile, err := os.Open(dir + "/verifying.key")
	if err != nil {
		return err
	}
	defer vkFile.Close()
	ps.vk = groth16.NewVerifyingKey(ecc.BN254)
	if _, err := ps.vk.ReadFrom(vkFile); err != nil {
		return err
	}

	return nil
}

func main() {
	// Initialize proof system
	ps, err := NewProofSystem()
	if err != nil {
		log.Fatalf("Failed to initialize proof system: %v", err)
	}

	// Example: Prove knowledge of x=3 where 3^3 + 3 + 5 = 35
	x := 3
	y := 35
	if err := ps.GenerateProof(x, y); err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	// Save keys for later use (in a real application)
	if err := ps.SaveToFiles("keys"); err != nil {
		log.Printf("Warning: failed to save keys: %v", err)
	}

	// Query the proof
	query := ps.NewProofQuery()

	// Example queries
	queries := []string{
		"proof.size",
		"proof.verify",
		"circuit.constraints",
		"witness.public",
		"witness.private",
	}

	fmt.Println("Proof System MVP Results:")
	for _, q := range queries {
		result, err := query.ExecuteQuery(q)
		if err != nil {
			fmt.Printf("Query '%s' failed: %v\n", q, err)
			continue
		}
		fmt.Printf("%s: %v\n", q, result)
	}

	fmt.Println("\nSuccess! Proof generated and verified.")
}
