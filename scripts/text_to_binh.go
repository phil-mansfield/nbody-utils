package main

import (
	"fmt"
	"os"
	"github.com/phil-mansfield/nbody-utils/io/catalogue"
)

const ErrorMessage = `Correct usage: text_to_binh config_name in_name out_name
type BinhConfig struct {
    ParticleMass float64
    MinParticles int64
    Columns      int64
    HeaderLines  int64
    MassColumn   int64
    ColumnInfo   []string
    SkipColumns  []string
    Sort         bool
}`

func main() {
	fmt.Println("Running on", os.Args)

	if len(os.Args) != 4 {
		fmt.Printf(ErrorMessage);	
		os.Exit(1)
	}

	configFile, haloFile, outFile := os.Args[1], os.Args[2], os.Args[3]

	binhConfig := catalogue.ParseBinhConfig(configFile)
	textConfig := catalogue.DefaultConfig

	catalogue.TextToBinh(haloFile, outFile, binhConfig, textConfig)

	fmt.Println("Done converting!")
}
