package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
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

type IgnoreInfo struct {
	Pattern string
	IsNeg   bool
}

func updateStat(filepath string, stats map[string]FileStats, hoursignore []IgnoreInfo) FileStatsDelta {
	files, err := os.ReadDir(filepath)
	if err != nil {
		fmt.Println(err)
	}

	deltas := make([]FileStatsDelta, 0)

	for _, file := range files {
		pathString := filepath + "/" + file.Name()
		if strings.Contains(pathString, ".git/") || pathString == filepath+"/hours.json" {
			continue
		}
		match := false
		for _, ignore := range hoursignore {
			matched, _ := path.Match(ignore.Pattern, pathString)
			match = match || matched
			if ignore.IsNeg && matched {
				match = false
				break
			}
		}

		if match {
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

			workInfo, ok := stats[pathString]
			if ok {
				delta = currentTime - workInfo.LastModified
				// If the distance between edits is more than 10 minutes, we trim it to a smaller amount
				// to be somewhat accurate to how long we worked before the 10 minutes was over

				start = workInfo.TotalWorked
				if delta > 60*10 {
					delta = 60
					start = currentTime - 60
				}

				totalWorked = workInfo.TotalWorked + delta
			}

			deltas = append(deltas, FileStatsDelta{currentTime, start, totalWorked})

			stats[pathString] = FileStats{stat.ModTime().Unix(), totalWorked}

		} else {
			deltas = append(deltas, updateStat(filepath+"/"+file.Name(), stats, hoursignore))
		}
	}

	var newest int64 = 0
	var earliestStart int64 = math.MaxInt64
	var latestEnd int64 = 0

	var subtotal int64 = 0
	workInfo, ok := stats[filepath]

	if ok {
		subtotal = workInfo.TotalWorked
	} else {
		workInfo = FileStats{0, 0}
	}

	for _, d := range deltas {
		if d.NewestMod > newest {
			newest = d.NewestMod
		}
		if d.Start < earliestStart && d.NewestMod >= workInfo.LastModified {
			earliestStart = d.Start
		}
		if d.End > latestEnd && d.NewestMod > workInfo.LastModified {
			latestEnd = d.End
		}
	}

	if latestEnd < earliestStart {
		latestEnd = earliestStart
	}

	stats[filepath] = FileStats{newest, subtotal + latestEnd - earliestStart}

	return FileStatsDelta{newest, earliestStart, latestEnd}
}

func updateAndSave(stats map[string]FileStats, hoursignore []IgnoreInfo) {
	updateStat(".", stats, hoursignore)

	j, _ := json.Marshal(stats)
	os.WriteFile("hours.json", j, 0644)
}

func jsonRead(filename string, v any) {
	body, err := os.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(body, v)
		if err != nil {
			return
		}
	} else {
		fmt.Println(err)
	}
}

func main() {
	stats := make(map[string]FileStats)
	hoursignore := make([]IgnoreInfo, 0)

	jsonRead("hours.json", &stats)
	jsonRead(".hoursignore.json", &hoursignore)

	for name, stat := range stats {
		stats[name] = FileStats{
			stat.LastModified,
			stat.TotalWorked,
		}
	}

	ticker := time.NewTicker(20 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				updateAndSave(stats, hoursignore)
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
			j, _ := json.MarshalIndent(stats, "", "    ")
			fmt.Println(string(j[:]))
		case 2:
			rootinfo, ok := stats["."]
			if ok {
				fmt.Println("You have spent", (rootinfo.TotalWorked/60)/60, "hours,",
					(rootinfo.TotalWorked/60)%60, "minutes,",
					rootinfo.TotalWorked%60, "seconds rootinfo working on this repo")
			} else {
				fmt.Println("Script has not scraped time yet for this repository")
			}
		}
	}

}
