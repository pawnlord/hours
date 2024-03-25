package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
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
		if strings.Contains(path, "git") || path == filepath+"/hours.json" {
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
			var delta int64 = 0

			workInfo, ok := stats[path]
			if ok {
				delta = currentTime - workInfo.LastModified
				// If the distance between edits is more than 10 minutes, we trim it to a smaller amount
				// to be maximally accurate to how long we worked before the 10 minutes was over
				if delta > 60*10 {
					delta = 60 * 5
				}

				start = workInfo.TotalWorked
				totalWorked = workInfo.TotalWorked + delta
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

func updateAndSave(stats map[string]FileStats) {
	// printStat(".")
	updateStat(".", stats)

	j, _ := json.Marshal(stats)
	// fmt.Println(string(j[:]))
	os.WriteFile("hours.json", j, 0644)
}

func main() {
	stats := make(map[string]FileStats)

	body, err := os.ReadFile("hours.json")
	if err == nil {
		err = json.Unmarshal(body, &stats)
		if err != nil {
			return
		}
	} else {
		fmt.Println(err)
	}

	for name, stat := range stats {
		fmt.Println(name, stat)
		stats[name] = FileStats{
			stat.LastModified,
			stat.TotalWorked,
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	var t time.Time
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case t = <-ticker.C:
				updateAndSave(stats)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	defer close(quit)

	for {
		fmt.Println("Choose an option")
		fmt.Println("    0) Exit")
		fmt.Println("    1) Print json")
		fmt.Println("    2) Print time")
		var choice int
		fmt.Scanln(&choice)
		switch choice {
		case 0:
			return
		case 1:
			j, _ := json.Marshal(stats)
			fmt.Println(string(j[:]))
		case 2:
			fmt.Println(t.Clock())
		}
	}

}
