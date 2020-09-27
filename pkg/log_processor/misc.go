package log_processor

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

// file format:
// 2.697654994s,
func GetTimeToNano(pathToFind string, inputfilename, outputfilename string) {
	inputFilename := path.Join(pathToFind, inputfilename)
	outputFilename := path.Join(pathToFind, outputfilename)

	inputFileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputFileHandler.Close()

	outputFileHandler, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("Error create output file [%s]: %v\n", outputFilename, err)
		panic(err)
	}
	defer outputFileHandler.Close()

	lineReader := bufio.NewReader(inputFileHandler)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		strValue := strings.TrimSpace(line)
		if strings.HasSuffix(strValue, ",") {
			strValue = line[0:len(strValue)-1]
		}
		//fmt.Println(strValue)
		timeValue, err := time.ParseDuration(strValue)
		if err != nil {
			fmt.Printf("Error getting time duration: value %s, error %v\n", strValue, err)
		} else {
			outputLine := fmt.Sprintf("%s, %v\n", strValue, timeValue.Nanoseconds())
			fmt.Printf(outputLine)
			outputFileHandler.WriteString(outputLine)
		}
	}
}
