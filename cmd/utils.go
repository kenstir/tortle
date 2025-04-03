/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"fmt"
	"strings"
	"time"
)

// convert a unix timestamp to a string for output
func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	//return t.Format(time.RFC3339) //2022-04-11T15:33:20-04:00
	return t.Format("2006-01-02 15:04:05")
}

// print a line of data in InfluxDB line protocol format
func printMeasurement(measurement string, tags []string, fields []string) {
	timestamp := time.Now().UnixNano()
	fmt.Printf("%s,%s %s %d\n",
		measurement,
		strings.Join(tags, ","),
		strings.Join(fields, ","),
		timestamp,
	)
}

func humanizeBytes(bytes int64) string {
	const (
		_ = 1 << (10 * iota)
		KB
		MB
		GB
		TB
		PB
	)
	// x := fmt.Sprintf("KB=%d MB=%d GB=%d TB=%d PB=%d", KB, MB, GB, TB, PB)
	// fmt.Println(x)
	switch {
	case bytes >= PB:
		return fmt.Sprintf("%.2f PiB", float64(bytes)/PB)
	case bytes >= TB:
		return fmt.Sprintf("%.2f TiB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KiB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
