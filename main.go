package main

import (
	"log"
	"time"
)

func main() {
	// start time measurement and parse command line flags
	start := time.Now()

	processCommandLineArgs()

	// load the incidents
	incidents, err := importIncidents(flagVars.inputFilename)
	if err != nil {
		log.Fatalf("Error importing %s: %v", flagVars.inputFilename, err)
	}
	if flagVars.verbose {
		log.Printf("Loaded a total of %d incidents from %s\n", len(incidents), flagVars.inputFilename)
	}

	processCommandLineCommand(incidents)

	if flagVars.verbose {
		log.Printf("Total running time: %s\n", time.Since(start))
	}
}
