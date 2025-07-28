package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strconv"
)

// Point represents a decoded (x, y) coordinate for the polynomial.
// We use *big.Int to handle potentially very large numbers.
type Point struct {
	X *big.Int
	Y *big.Int
}

// KeyInfo holds the metadata from the "keys" object in the JSON.
type KeyInfo struct {
	N int `json:"n"`
	K int `json:"k"`
}

// RootValue represents the encoded Y value and its base from the JSON.
type RootValue struct {
	Base  string `json:"base"`
	Value string `json:"value"`
}

// solveForSecret reads a test case file, decodes the points,
// and calculates the polynomial's constant term 'c'.
func solveForSecret(filePath string) (*big.Int, error) {
	// --- 1. Read the Test Case (Input) from a separate JSON file ---
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Use a map to handle the dynamic keys ("1", "2", "3", etc.)
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json from %s: %w", filePath, err)
	}

	// Parse the 'keys' object
	var keys KeyInfo
	if err := json.Unmarshal(rawData["keys"], &keys); err != nil {
		return nil, fmt.Errorf("failed to parse 'keys' object in %s: %w", filePath, err)
	}

	// --- 2. Decode the Y Values and collect points ---
	var points []Point
	// Sort keys to ensure we get a consistent set of points if n > k
	var sortedKeys []string
	for keyStr := range rawData {
		if keyStr != "keys" {
			sortedKeys = append(sortedKeys, keyStr)
		}
	}
	sort.Strings(sortedKeys)

	// We only need 'k' points to define the polynomial
	for _, keyStr := range sortedKeys {
		if len(points) >= keys.K {
			break
		}

		// The key is the 'x' coordinate
		x, ok := new(big.Int).SetString(keyStr, 10)
		if !ok {
			return nil, fmt.Errorf("failed to parse x-coordinate '%s' to integer", keyStr)
		}

		// Decode the corresponding 'y' coordinate
		var rootVal RootValue
		if err := json.Unmarshal(rawData[keyStr], &rootVal); err != nil {
			return nil, fmt.Errorf("failed to parse root object for key '%s': %w", keyStr, err)
		}

		base, err := strconv.Atoi(rootVal.Base)
		if err != nil {
			return nil, fmt.Errorf("invalid base '%s' for key '%s'", rootVal.Base, keyStr)
		}

		y, ok := new(big.Int).SetString(rootVal.Value, base)
		if !ok {
			return nil, fmt.Errorf("failed to parse y-value '%s' in base %d for key '%s'", rootVal.Value, base, keyStr)
		}

		points = append(points, Point{X: x, Y: y})
	}

	if len(points) < keys.K {
		return nil, fmt.Errorf("not enough points provided: need %d, got %d", keys.K, len(points))
	}

	// --- 3. Find the Secret (C) using Lagrange Interpolation ---
	// The secret c is the value of the polynomial at x=0, i.e., f(0).
	// c = f(0) = Σ [y_j * L_j(0)]
	// L_j(0) = Π [x_i / (x_i - x_j)] for i != j

	// We use rational numbers (big.Rat) for calculations to avoid precision loss from division.
	totalSum := new(big.Rat) // Initializes to 0/1

	for j := 0; j < keys.K; j++ {
		xj := points[j].X
		yj := points[j].Y

		// Calculate L_j(0)
		numerator := big.NewInt(1)
		denominator := big.NewInt(1)

		for i := 0; i < keys.K; i++ {
			if i == j {
				continue
			}
			xi := points[i].X
			
			// Numerator term: x_i
			numerator.Mul(numerator, xi)

			// Denominator term: (x_i - x_j)
			diff := new(big.Int).Sub(xi, xj)
			denominator.Mul(denominator, diff)
		}

		// Now we have L_j(0) = numerator / denominator.
		// The full term for the sum is y_j * L_j(0).
		// We can multiply y_j into the numerator.
		termNumerator := new(big.Int).Mul(yj, numerator)
		
		// Create the rational number for this term
		term := new(big.Rat).SetFrac(termNumerator, denominator)
		
		// Add it to our total sum
		totalSum.Add(totalSum, term)
	}

	// The final result 'c' must be an integer, as per the problem constraints.
	if !totalSum.IsInt() {
		return nil, fmt.Errorf("fatal: final result is not an integer, something went wrong with the calculation. Result: %s", totalSum.FloatString(5))
	}

	// Return the integer part of the result.
	return totalSum.Num(), nil
}

func main() {
	testFiles := []string{"testcase1.json", "testcase2.json"}

	fmt.Println("Catalog Placements Assignment - Shamir's Secret Sharing")
	fmt.Println("======================================================")

	for _, file := range testFiles {
		secret, err := solveForSecret(file)
		if err != nil {
			log.Fatalf("Error processing %s: %v", file, err)
		}
		fmt.Printf("Secret for %s: %s\n", file, secret.String())
	}
}