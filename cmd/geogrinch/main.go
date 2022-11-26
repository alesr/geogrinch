package main

import (
	"log"
	"os"

	"github.com/alesr/geogrinch/internal/dataset"
)

const datasetFilePath string = "dataset/dataset.csv"

func main() {
	file, err := os.Open(datasetFilePath)
	if err != nil {
		log.Fatalln("failed to open dataset file:", err)
	}

	ds, err := dataset.Load(file)
	if err != nil {
		log.Fatalln("failed to load dataset:", err)
	}

	if err := ds.CalculateVariances(); err != nil {
		log.Fatalln("failed to calculate variances:", err)
	}

	ds.CalculateFDistributions()

	ds.PrintDataset()
	ds.PrintVariances()
	ds.PrintFDistributions()
}
