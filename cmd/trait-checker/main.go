package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/brentp/vcfgo"
)

type TraitVariant struct {
	Trait      string `json:"trait"`
	Gene       string `json:"gene"`
	Chromosome int    `json:"chromosome"`
	Position   int    `json:"position"`
	Region     struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"region"`
	Ref string `json:"ref"`
	Alt string `json:"alt"`
}

func main() {
	vcfPath := flag.String("vcf", "", "Path to VCF file")
	traitPath := flag.String("traits", "panels_traits.json", "Path to trait panel JSON file")
	flag.Parse()

	if *vcfPath == "" {
		fmt.Println("Error: -vcf is required")
		os.Exit(1)
	}

	fmt.Printf("Loading trait panel from %s...\n", *traitPath)
	// Load trait panel
	data, err := os.ReadFile(*traitPath)
	if err != nil {
		fmt.Printf("Error reading trait panel: %v\n", err)
		os.Exit(1)
	}

	var traits []TraitVariant
	if err := json.Unmarshal(data, &traits); err != nil {
		fmt.Printf("Error parsing trait panel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d traits from panel\n", len(traits))

	// Create position lookup map
	positions := make(map[int]TraitVariant)
	fmt.Println("\nPositions to search for:")
	for _, trait := range traits {
		positions[trait.Position] = trait
		fmt.Printf("- Position %d: %s (%s)\n", trait.Position, trait.Trait, trait.Gene)
	}

	fmt.Printf("\nOpening VCF file %s...\n", *vcfPath)
	// Open VCF file
	f, err := os.Open(*vcfPath)
	if err != nil {
		fmt.Printf("Error opening VCF: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	rdr, err := vcfgo.NewReader(f, false)
	if err != nil {
		fmt.Printf("Error creating VCF reader: %v\n", err)
		os.Exit(1)
	}

	found := make(map[int]bool)
	fmt.Println("\nSearching VCF file for trait positions...")

	// Read VCF and check positions
	variantCount := 0
	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}
		variantCount++
		if variantCount%100000 == 0 {
			fmt.Printf("Processed %d variants...\n", variantCount)
		}

		if trait, exists := positions[int(variant.Pos)]; exists {
			found[int(variant.Pos)] = true
			fmt.Printf("✓ FOUND: %s (%s) at position %d\n", trait.Trait, trait.Gene, variant.Pos)
		}
	}

	// Summary
	fmt.Printf("\nSUMMARY: Found %d out of %d traits\n", len(found), len(traits))

	if len(found) == 0 {
		fmt.Println("No trait positions found in VCF file")
	} else {
		fmt.Println("\nMissing traits:")
		for pos, trait := range positions {
			if !found[pos] {
				fmt.Printf("- %s (%s) at position %d\n", trait.Trait, trait.Gene, pos)
			}
		}
	}
}
