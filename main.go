package main

import (
	"fmt"
	"github.com/brentp/vcfgo"
	"github.com/consensys/gnark/frontend"
	"os"
)

// TODO: implement this
type VariantCircuit struct {
	// NOTE: what parts of the variant should be public vs private?
	/*
		type Variant struct {
			Chromosome string
			Pos        uint64
			Id_        string
			Reference  string
			Alternate  []string
			Quality    float32
			Filter     string
			Info_      interfaces.Info
			Format     []string
			Samples    []*SampleGenotype
			// if lazy parsing, then just save the sample strings here.
			sampleString string
			Header       *Header
			LineNumber   int64
		}
	*/
}

func main() {
	f, _ := os.Open("data/genome_example.vcf")
	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		panic(err)
	}

	count := 0
	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}

		fmt.Printf("%s\t%d\t%s\t%v\n", variant.Chromosome, variant.Pos, variant.Ref(), variant.Alt())

		// only working with first 100 variants for debuging
		if count > 100 {
			break
		}
		count++
	}
}

// given example of vcf
/*
	f, _ := os.Open("examples/test.auto_dom.no_parents.vcf")
	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		panic(err)
	}
	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}
		fmt.Printf("%s\t%d\t%s\t%v\n", variant.Chromosome, variant.Pos, variant.Ref(), variant.Alt())
		dp, err := variant.Info().Get("DP")
		fmt.Printf("depth: %v\n", dp.(int))
		sample := variant.Samples[0]
		// we can get the PL field as a list (-1 is default in case of missing value)
		PL, err := variant.GetGenotypeField(sample, "PL", -1)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", PL)
		_ = sample.DP
	}
	fmt.Fprintln(os.Stderr, rdr.Error())

*/

// gnark usage example
/*
type CubicCircuit struct {
	// struct tags on a variable is optional
	// default uses variable name and secret visibility.
	X frontend.Variable `gnark:"x"`
	Y frontend.Variable `gnark:",public"`
}

// Define declares the circuit constraints
// x**3 + x + 5 == y
func (circuit *CubicCircuit) Define(api frontend.API) error {
	x3 := api.Mul(circuit.X, circuit.X, circuit.X)
	api.AssertIsEqual(circuit.Y, api.Add(x3, circuit.X, 5))
	return nil
}

func main() {
	// compiles our circuit into a R1CS
	var circuit CubicCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)

	// groth16 zkSNARK: Setup
	pk, vk, _ := groth16.Setup(ccs)

	// witness definition
	assignment := CubicCircuit{X: 3, Y: 35}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	// groth16: Prove & Verify
	proof, _ := groth16.Prove(ccs, pk, witness)
}
*/
