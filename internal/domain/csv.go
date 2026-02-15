// ABOUTME: This file implements CSV parsing for workout data exports.
// ABOUTME: It reads Strong-app style CSV files and produces Lift structs with deterministic IDs.
package domain

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseLiftsCSV reads a Strong-app CSV export and returns Lift structs.
// Each row maps to one set with a deterministic ID for idempotent imports.
func ParseLiftsCSV(r io.Reader) ([]Lift, error) {
	reader := csv.NewReader(r)

	// Read and validate header
	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	colIndex := make(map[string]int, len(header))
	for i, col := range header {
		colIndex[strings.TrimSpace(col)] = i
	}

	// Verify required columns exist
	required := []string{"Date", "Workout Name", "Duration", "Exercise Name", "Set Order", "Weight", "Reps", "Distance", "Seconds"}
	for _, col := range required {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required CSV column: %s", col)
		}
	}

	var lifts []Lift
	// AIDEV-NOTE: Track seen IDs to disambiguate duplicate (date, exercise, setOrder) combos.
	idCount := make(map[string]int)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row: %w", err)
		}

		setOrder := strings.TrimSpace(record[colIndex["Set Order"]])

		weight, err := parseFloatField(record, colIndex, "Weight")
		if err != nil {
			return nil, err
		}

		reps, err := parseFloatField(record, colIndex, "Reps")
		if err != nil {
			return nil, err
		}

		distance, err := parseFloatField(record, colIndex, "Distance")
		if err != nil {
			return nil, err
		}

		seconds, err := parseFloatField(record, colIndex, "Seconds")
		if err != nil {
			return nil, err
		}

		var rpe float64
		if idx, ok := colIndex["RPE"]; ok && idx < len(record) {
			rpeStr := strings.TrimSpace(record[idx])
			if rpeStr != "" {
				rpe, err = strconv.ParseFloat(rpeStr, 64)
				if err != nil {
					return nil, fmt.Errorf("parse RPE %q: %w", rpeStr, err)
				}
			}
		}

		date := strings.TrimSpace(record[colIndex["Date"]])
		exerciseName := strings.TrimSpace(record[colIndex["Exercise Name"]])

		baseID := LiftID(date, exerciseName, setOrder)
		idCount[baseID]++
		id := baseID
		if idCount[baseID] > 1 {
			id = fmt.Sprintf("%s-%d", baseID, idCount[baseID])
		}

		lift := Lift{
			ID:           id,
			Date:         date,
			WorkoutName:  strings.TrimSpace(record[colIndex["Workout Name"]]),
			Duration:     strings.TrimSpace(record[colIndex["Duration"]]),
			ExerciseName: exerciseName,
			SetOrder:     setOrder,
			Weight:       weight,
			Reps:         reps,
			Distance:     distance,
			Seconds:      seconds,
			RPE:          rpe,
		}

		lifts = append(lifts, lift)
	}

	return lifts, nil
}

func parseFloatField(record []string, colIndex map[string]int, col string) (float64, error) {
	val := strings.TrimSpace(record[colIndex[col]])
	if val == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", col, val, err)
	}
	return f, nil
}
