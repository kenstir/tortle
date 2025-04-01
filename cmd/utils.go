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
func printMeasurement(measurement string, tags []string, fields []string, timestamp int64) {
	fmt.Printf("%s,%s %s %d\n",
		measurement,
		strings.Join(tags, ","),
		strings.Join(fields, ","),
		timestamp,
	)
}
