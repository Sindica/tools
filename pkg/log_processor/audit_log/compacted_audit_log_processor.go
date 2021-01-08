package apiserver_audit_log

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"tools/pkg/log_util"
)

func ExtractCompactedAuditLog(outputPath string, inputFilename string, threadhold int) {
	filenameShort := log_util.GetFilenameOnly(inputFilename)
	outputFilename := path.Join(outputPath, "combined-"+filenameShort)
	xlOutputFilename := path.Join(outputPath, "combined-xl-"+filenameShort)
	errorAuditLogFilename := path.Join(outputPath, "error-"+filenameShort)
	ProcessCompactedAuditLog(inputFilename, outputFilename, xlOutputFilename, errorAuditLogFilename, threadhold)
}

func ProcessCompactedAuditLog(inputFilename, outputFilename, xlOutputFilename, errorAuditLogFilename string, threadhold int) {
	inputfileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		return
	}

	defer inputfileHandler.Close()

	outputFileHandler, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename, err)
		return
	}
	defer outputFileHandler.Close()

	xlOuputFileHandler, err := os.Create(xlOutputFilename)
	if err != nil {
		fmt.Printf("Error open extra large audit output file [%s]: %v\n", xlOutputFilename, err)
		return
	}
	defer xlOuputFileHandler.Close()

	errorFileHandler, err := os.Create(errorAuditLogFilename)
	if err != nil {
		fmt.Printf("Error open error file [%s]: %v\n", errorAuditLogFilename, err)
		return
	}
	defer errorFileHandler.Close()

	lineReader := bufio.NewReader(inputfileHandler)
	lineCount := 0
	reqURIMap := make(map[string]map[string]*requestCount, 0)

	// read/parse input data
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}
		lineCount++

		uri, reqCount, isSkip := getRequestCount(line)
		if isSkip {
			errorFileHandler.WriteString(line)
			continue
		}

		key := fmt.Sprintf("%s:%s", reqCount.Verb, reqCount.Code)
		if reqCountMap, isOK := reqURIMap[uri]; isOK {
			if reqCountEntry, isOK := reqCountMap[key]; isOK {
				reqCountEntry.Count+= reqCount.Count
			} else {
				reqCountMap[key] = reqCount
			}
		} else {
			reqCountMap := make(map[string]*requestCount, 1)
			reqCountMap[key] = reqCount
			reqURIMap[uri] = reqCountMap
		}
	}

	// output
	header := "uri, verb, response_code, count, stage\n"
	outputFileHandler.WriteString(header)
	xlOuputFileHandler.WriteString(header)
	for uri, reqCountMap := range reqURIMap {
		for _, reqCount := range reqCountMap {
			line := fmt.Sprintf("%s, %s, %d, %d, %s", uri, reqCount.Verb, reqCount.Code, reqCount.Count, reqCount.Stage)
			outputFileHandler.WriteString(line)

			if reqCount.Count >= threadhold {
				xlOuputFileHandler.WriteString(line)
			}
		}
	}
}

/* Sample file:
uri, verb, response_code, count, stage
/apis/arktos.futurewei.com/v1/tenants/system/networks/default, get, 404, 2, ResponseComplete
/api/v1/tenants/system/namespaces/lodkz7-testns/secrets, list, 200, 1, ResponseComplete
/api/v1/nodes/hollow-node-54fsg, get, 200, 1, ResponseComplete
*/
func getRequestCount(input string) (string, *requestCount, bool) {
	strArray := strings.Split(input, ", ")
	if len(strArray) != 5 {
		return "", nil, true
	}
	if strArray[0] == "uri" {
		return "", nil, true
	}

	code, err := strconv.ParseInt(strArray[2], 10, 32)
	if err != nil {
		return "", nil, true
	}
	count, err := strconv.ParseInt(strArray[3], 10, 32)
	if err != nil {
		return "", nil, true
	}
	reqCount := &requestCount{
		Verb:  strArray[1],
		Code:  int(code),
		Count: int(count),
		Stage: strArray[4],
	}

	return strArray[0], reqCount, false
}