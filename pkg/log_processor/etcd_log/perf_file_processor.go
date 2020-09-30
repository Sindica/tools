package etcd_log

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func ParseRangeLog(pathToFind string) {
	inputFile := "etcd.to.execute.range.log.compacted"
	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".keycount")
	fmt.Println("File " + inputFilename)
	AnalysisReadOnlyRangePerfData(inputFilename, outputFilename, "RangeOnly")
}

func ParNonRangeLog(pathToFind string) {
	inputFile := "etcd.to.execute.norange.log.compacted"
	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".keycount")
	fmt.Println("File " + inputFilename)
	AnalysisReadOnlyRangePerfData(inputFilename, outputFilename,"NonRange")
}

func AnalysisReadOnlyRangePerfData(inputFilename, outputFilename string, perfFileType string) {
	inputfileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputfileHandler.Close()

	outputFileHandler, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename, err)
		panic(err)
	}
	defer outputFileHandler.Close()

	lineReader := bufio.NewReader(inputfileHandler)
	lineCount := 0
	timeLimit1 := 100000	// 100 ms
	timeLimit2 := 10 * timeLimit1 // 1s
	timeLimit3 := 5 * timeLimit2 // 5s
	timeLimit4 := 10 * timeLimit2 // 10s
	countExceed1 := 0
	countExceed2 := 0
	countExceed3 := 0
	countExceed4 := 0

	keyCount := make(map[string]int)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		lineCount++
		var durationStr string
		var key string
		if perfFileType == "RangeOnly" {
			key, durationStr = getReadOnlyRangePerfData(line)
		} else {
			key, durationStr = getNonRangePerfData(line)
		}
		durationInMicroSec, _ := strconv.Atoi(durationStr)

		if durationInMicroSec > 0 {
			if durationInMicroSec > timeLimit1 {
				countExceed1++
				if durationInMicroSec > timeLimit2 {
					countExceed2++
					if durationInMicroSec > timeLimit3 {
						countExceed3++
						if durationInMicroSec > timeLimit4 {
							countExceed4++
						}
					}
				}
			}

			if v, isOK := keyCount[key]; isOK {
				keyCount[key] = v + 1
			} else {
				keyCount[key] = 1
			}
		}
	}

	fmt.Printf("Scanned %d lines, found %d line exceeds 100ms, %d line excceds 1s, %d line exceeds 5s, %d line exceeds 10s\n",
		lineCount, countExceed1, countExceed2, countExceed3, countExceed4)

	// output key count into file
	keyArray := make([]string, len(keyCount))
	index := 0
	for k, _ := range keyCount {
		keyArray[index] = k
		index++
	}
	sort.Strings(keyArray)
	outputFileHandler.WriteString("key, count\n")
	countTotal := 0
	for i:=0; i < index; i++ {
		v, _ := keyCount[keyArray[i]]
		outputLine := fmt.Sprintf("%s, %d\n", keyArray[i], v)
		outputFileHandler.WriteString(outputLine)
		countTotal += v
	}
	outputFileHandler.WriteString(fmt.Sprintf("Total %d keys\n", index))
	fmt.Printf("Key count file generated. Total %d keys, count total %d. Equal line total %v\n",
		index, countTotal, countTotal + 1 == lineCount)
}

func getReadOnlyRangePerfData(line string) (string, string) {
	fields := strings.Split(line, ", ")
	req := RangeOnlyRangeRequest{
		key:                fields[0],
		range_end:          fields[1],
		count_only:         fields[2],
		limit:              fields[3],
		count:              fields[4],
		size:               fields[5],
		durationInMicroSec: fields[6][:len(fields[6])-1],
	}

	return req.key, req.durationInMicroSec
}

func getNonRangePerfData(line string) (string, string) {
	fields := strings.Split(line, ", ")
	req := NoRangeRequest{
		key:                fields[0],
		method:             fields[1],
		mod_revision:       fields[2],
		success_method:     fields[3],
		success_value_size: fields[4],
		failure_method:     fields[5],
		size:               fields[6],
		durationInMicroSec: fields[7][:len(fields[7])-1],
	}

	return req.key, req.durationInMicroSec
}