package main

import "strings"

func FindRunnableQueryRegions(lines *[]string) (result map[int]int) {
	result = make(map[int]int)
	region := 1
	findNext := false

	for t, line := range *lines {
		empty := len(strings.TrimSpace(line)) == 0
		if empty {
			if !findNext {
				result[t+1] = 0
				region++
				findNext = true
			}

		} else {
			result[t+1] = region
			findNext = false
		}
	}

	return
}
