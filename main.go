package main

import (
	"fmt"
	"github.com/brentp/vcfgo"
	"os"
)

func main() {
	f, _ := os.Open("data/genome_example.vcf")
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
		dp, _ := variant.Info().Get("DP")
		fmt.Printf("depth: %v\n", dp.(int))
		sample := variant.Samples[0]
		// we can get the PL field as a list (-1 is default in case of missing value)
		/*
			PL, err := variant.GetGenotypeField(sample, "PL", -1)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%v\n", PL)
		*/
		_ = sample.DP
	}
	fmt.Fprintln(os.Stderr, rdr.Error())

}
