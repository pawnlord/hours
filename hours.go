package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type FileStats struct {
	LastModified int64
	TotalWorked  int64
}

type FileStatsDelta struct {
	NewestMod int64
	Start     int64
	End       int64
}

func printStat(filepath string) {
	files, err := os.ReadDir(filepath)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println(file.Name())
			stat, err := file.Info()
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(filepath+"/"+stat.Name(), stat.Size(), stat.ModTime().Unix())
		} else {
			printStat(filepath + "/" + file.Name())
		}
	}
}

func updateStat(filepath string, stats map[string]FileStats) []FileStatsDelta {
	files, err := os.ReadDir(filepath)
	if err != nil {
		fmt.Println(err)
	}

	deltas := make([]FileStatsDelta, 0)

	for _, file := range files {
		path := filepath + "/" + file.Name()
		if strings.Contains(path, "git") {
			continue
		}
		if !file.IsDir() {

			stat, err := file.Info()
			if err != nil {
				fmt.Println(err)
			}

			currentTime := stat.ModTime().Unix()

			var totalWorked int64 = 0
			var start int64 = 0

			workInfo, ok := stats[path]
			if ok {
				start = workInfo.TotalWorked
				totalWorked = workInfo.TotalWorked + (currentTime - workInfo.LastModified)
			}

			deltas = append(deltas, FileStatsDelta{currentTime, start, totalWorked})

			stats[path] = FileStats{stat.ModTime().Unix(), totalWorked}

		} else {
			subdeltas := updateStat(filepath+"/"+file.Name(), stats)

			var newest int64 = 0
			var earliestStart int64 = 0
			var latestEnd int64 = 0

			for _, d := range subdeltas {
				if d.NewestMod > newest {
					newest = d.NewestMod
				}
				if d.Start < earliestStart {
					earliestStart = d.Start
				}
				if d.End > latestEnd {
					latestEnd = d.End
				}
			}

			var subtotal int64 = 0
			if workInfo, ok := stats[path]; ok {
				subtotal = workInfo.TotalWorked
			}

			stats[path] = FileStats{newest, subtotal + latestEnd - earliestStart}
		}
	}
	return deltas
}

func main() {
	stats := make(map[string]FileStats)

	body, err := os.ReadFile("hours.json")
	if err == nil {
		err = json.Unmarshal(body, &stats)
		if err != nil {
			return
		}
	}

	// printStat(".")
	updateStat(".", stats)

	j, _ := json.Marshal(stats)
	fmt.Println(string(j[:]))
	os.WriteFile("hours.json", j, 0)
}
