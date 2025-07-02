package core

import (
	"flag"
)

// CheckExtractionFlags checks for route extraction flags and enables extraction if needed
// This should be called early in main() before any module registration
func CheckExtractionFlags() {
	// Check if --extract-routes flag is present
	extractRoutes := flag.Bool("extract-routes", false, "Extract route metadata and exit")
	extractOutput := flag.String("extract-output", "./generated/routes.json", "Output path for extracted route metadata")
	
	// Parse flags if not already parsed
	if !flag.Parsed() {
		flag.Parse()
	}
	
	if *extractRoutes {
		InitExtractor(*extractOutput)
	}
}

// MustCheckExtractionFlags is like CheckExtractionFlags but panics if called after flag.Parse()
func MustCheckExtractionFlags() {
	if flag.Parsed() {
		panic("MustCheckExtractionFlags must be called before flag.Parse()")
	}
	CheckExtractionFlags()
}