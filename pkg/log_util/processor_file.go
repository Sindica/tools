package log_util

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ExtractMatchingLines(inputFile, outputFile string, searchingRegex []string) {
	inputfileHandler, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFile, err)
		panic(err)
	}

	defer inputfileHandler.Close()

	outputFileHandler, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error create output file %s: %v\n", outputFile, err)
		panic(err)
	}
	defer outputFileHandler.Close()

	if len(searchingRegex) == 0 {
		fmt.Println("Empty search regex")
		return
	}

	lineReader := bufio.NewReader(inputfileHandler)

	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		for _, checker := range searchingRegex {
			isMatched, err := regexp.MatchString(checker, line)
			if err == nil && isMatched {
				outputFileHandler.WriteString(line)
			}
		}
	}
}

func GetTimeFromLog(line string) (string, error) {
	strsByEmptySpace := strings.Split(line, " ")
	if len(strsByEmptySpace) < 2 {
		return "", fmt.Errorf("Cannot get time from log [%s]", line)
	}

	return strsByEmptySpace[1], nil
}

// 01:24:22.406258
func GetTimeDiff(time1, time2 string) (time.Duration, error) {
	h1, min1, sec1, ms1, err1 := parseTime(time1)
	if err1 != nil {
		return 0, err1
	}

	h2, min2, sec2, ms2, err2 := parseTime(time2)
	if err2 != nil {
		return 0, err2
	}

	loc, _ := time.LoadLocation("Local")

	resultTime1 := time.Date(2020, 1, 1, h1, min1, sec1, ms1, loc)
	resultTime2 := time.Date(2020, 1, 1, h2, min2, sec2, ms2, loc)
	if resultTime2.Before(resultTime1) {
		resultTime2 = time.Date(2020, 1, 2, h2, min2, sec2, ms2, loc)
	}

	timeDiff := resultTime2.Sub(resultTime1)
	fmt.Printf("Start time %s, completed time %s, diff %v\n", time1, time2, timeDiff)
	return timeDiff, nil
}

func parseTime(time1 string) (int, int, int, int, error) {
	if time1 == "" {
		return 0, 0, 0, 0, fmt.Errorf("Missing time. [%s]", time1)
	}

	timeSplit1 := strings.Split(time1, ":")
	if len(timeSplit1) != 3 {
		return 0, 0, 0, 0, fmt.Errorf("Time has invalid format %s", time1)
	}

	h, err1 := strconv.Atoi(timeSplit1[0])
	m, err2 := strconv.Atoi(timeSplit1[1])
	secondFull := timeSplit1[2]
	s, err3 := strconv.Atoi(secondFull[0:2])
	ms, err4 := strconv.Atoi(secondFull[3:])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, 0, 0, 0, fmt.Errorf("Time has invalid format %s", time1)
	}
	return h, m, s, ms, nil
}
