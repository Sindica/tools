package apiserver_audit_log

import (
	"bufio"
	"fmt"
	"kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
	"os"
	"path"
	"strings"
	"tools/pkg/log_util"
)

type userAuditLog struct {
	Username string `json:"username"`
	Groups []string `json:"groups"`
}

type responseStatusAuditLog struct {
	// metadata
	Code int `json:"code"`
}

type APIServerAuditLog struct {
	Kind string `json:"kind""`
	ApiVersion string `json:"apiVersion"`
	Level string `json:"level"`
	AuditID string `json:"auditID"`
	Stage string `json:"stage"`
	RequestURI string `json:"requestURI"`
	Verb string `json:"verb"`
	User userAuditLog `json:"user"`
	SourceIPs []string `json:"sourceIPs"`
	UserAgent string `json:"userAgent"`
	//ObjectRef
	ResponseStatus responseStatusAuditLog `json:"responseStatus"`
	RequestReceivedTimeStamp string `json:"requestReceivedTimeStamp"`
	StageTimeStamp string `json:"stageTimeStamp"`
	//annotations
}

type requestCount struct {
	// RequestURI string
	Verb string
	Code int
	Count int
	Stage string
}

func readAuditLog(filePath string, errAuditFileHandler *os.File) ([]APIServerAuditLog, error) {
	inputfileHandler, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer inputfileHandler.Close()

	lineReader := bufio.NewReader(inputfileHandler)
	lineCount := 0
	apiServerAuditLog := make([]APIServerAuditLog, 0)

	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		lineCount++

		var log APIServerAuditLog
		err = json.Unmarshal([]byte(line), &log)
		if err != nil {
			errAuditFileHandler.WriteString(line)
			continue
			//fmt.Printf("Error unmarshall line [%s]: %v\n", line, err)
			//return nil, err
		}
		apiServerAuditLog = append(apiServerAuditLog, log)
	}

	fmt.Printf("Total processd line %d\n", lineCount)

	return apiServerAuditLog, nil
}

func processAuditLog(inputFilename string, outputFileHandler1, outputFileHandler2, otherFileHandler, errAuditFileHandler *os.File) {
	apiServerAuditLog, err := readAuditLog(inputFilename, errAuditFileHandler)
	if err != nil {
		fmt.Printf("Error reading audit log: %v\n", err)
		return
	}

	// map compacted requestURI -> key -> requestCount
	reqURIMap := make(map[string]map[string]*requestCount, 0)
	for _, log := range apiServerAuditLog {
		key := fmt.Sprintf("%s:%s:s", log.Verb, log.ResponseStatus.Code, log.Stage)
		compactedURI := getCompactURI(log.RequestURI)

		if reqCountMap, isOK := reqURIMap[compactedURI]; isOK {
			if reqCount, isOK := reqCountMap[key]; isOK {
				reqCount.Count++
			} else {
				reqCountMap[key] = &requestCount{
					Verb:  log.Verb,
					Code:  log.ResponseStatus.Code,
					Count: 1,
					Stage: log.Stage,
				}
			}
		} else {
			reqCountMap := make(map[string]*requestCount, 1)
			reqCountMap[key] = &requestCount{
				Verb:  log.Verb,
				Code:  log.ResponseStatus.Code,
				Count: 1,
				Stage: log.Stage,
			}

			reqURIMap[compactedURI] = reqCountMap
		}
	}

	// print out count
	header := "uri, verb, response_code, count, stage\n"
	outputFileHandler1.WriteString(header)
	outputFileHandler2.WriteString(header)
	otherFileHandler.WriteString(header)
	for uri, reqCountMap := range reqURIMap {
		for _, reqCount := range reqCountMap {
			line := fmt.Sprintf("%s, %s, %d, %d, %s\n", uri, reqCount.Verb, reqCount.Code, reqCount.Count, reqCount.Stage)
			switch reqCount.Stage {
			case "ResponseStarted":
				outputFileHandler1.WriteString(line)
			case "ResponseComplete":
				outputFileHandler2.WriteString(line)
			default:
				otherFileHandler.WriteString(line)
			}
		}
	}
}

func getCompactURI(requestURI string) string {
	stringArray := strings.Split(requestURI, "?")
	if len(stringArray) > 0 {
		return stringArray[0]
	}

	return requestURI
}

func ProcessAuditLog(inputFilename, outputFilename1, outputFilename2, otherFilename, errorAuditLogFilename string) {
	outputFileHandler1, err := os.Create(outputFilename1)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename1, err)
		return
	}
	defer outputFileHandler1.Close()

	outputFileHandler2, err := os.Create(outputFilename2)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename2, err)
		return
	}
	defer outputFileHandler2.Close()

	otherFileHandler, err := os.Create(otherFilename)
	if err != nil {
		fmt.Printf("Error open unexpected audit log file [%s]: %v\n", otherFilename, err)
		return
	}
	defer otherFileHandler.Close()

	errAuditFileHandler, err := os.Create(errorAuditLogFilename)
	if err != nil {
		fmt.Printf("Error open unparserable audit log file [%s]: %v\n", errorAuditLogFilename, err)
		return
	}
	defer errAuditFileHandler.Close()

	processAuditLog(inputFilename, outputFileHandler1, outputFileHandler2, otherFileHandler, errAuditFileHandler)
}

func ExtractAuditLog(outputPath string, inputFilename string) {
	filenameShort := log_util.GetFilenameOnly(inputFilename)
	outputFilename1 := path.Join(outputPath, "compact-start-"+filenameShort)
	outputFilename2 := path.Join(outputPath, "compact-complete-"+filenameShort)
	otherFilename := path.Join(outputPath, "compact-Unexpected-"+filenameShort)
	errorAuditLogFilename := path.Join(outputPath, "error-entry-"+filenameShort)
	ProcessAuditLog(inputFilename, outputFilename1, outputFilename2, otherFilename, errorAuditLogFilename)
}
