#GEOGRINCH

Geogrinch is a simple program that calculates the variances and F-distributions for a given dataset .csv.

## Usage

1. Install Go
2. Replace the dataset file (.csv)
3. Run `go run cmd/geogrinch/main.go`

The dataset must respect the format `Sample;X_UTM;Y_UTM;Group;Pb_ppm;As_ppm;Sb_ppm;V_ppm`, and decimals most be separated by `'.'` not `','`.
