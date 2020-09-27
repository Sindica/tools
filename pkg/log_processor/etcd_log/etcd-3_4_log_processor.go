package etcd_log

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type RangeOnlyRangeRequest struct {
	key string
	range_end string
	limit string
	count_only string
	count string
	size string
	durationInMicroSec string
}
func ReadOnlyRangeRequest_Parser(inputFileName, outputFileName, nonMatchingFilename string) {
	inputfileHandler, err := os.Open(inputFileName)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFileName, err)
		panic(err)
	}
	defer inputfileHandler.Close()

	outputFileHandler, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFileName, err)
		panic(err)
	}
	defer outputFileHandler.Close()

	otherFileHandler, err := os.Create(nonMatchingFilename)
	if err != nil {
		fmt.Printf("Error open non matching file [%s]: %v\n", nonMatchingFilename, err)
		panic(err)
	}
	defer otherFileHandler.Close()

	lineReader := bufio.NewReader(inputfileHandler)
	lineCount := 0
	fields18Count := 0
	fields19Count := 0
	fields20ICount := 0
	fields20WCount := 0
	fields21ICount := 0
	fields21WCount := 0
	fields22Count := 0
	fieldsHasLimitCount := 0
	fieldsNoLimitCount := 0

	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		lineCount++

		fields := strings.Split(line, " ")
		req := RangeOnlyRangeRequest{}
		hasError := false

		req.key = fields[8]
		switch len(fields) {
		case 18:
			// etcd.log:2020-09-25 19:24:07.605099 I | etcdserver: read-only range request "key:\"/registry/masterleases/10.40.0.12\" " with result "range_response_count:0 size:4" took (237.078µs) to execute
			req.count = fields[12]
			req.size = fields[13]
			req.durationInMicroSec = fields[15]

			fields18Count++
		case 19:
			// etcd.log:2020-09-25 19:24:07.823156 I | etcdserver: read-only range request "key:\"/registry/services/specs/\" range_end:\"/registry/services/specs0\" " with result "range_response_count:0 size:4" took (296.672µs) to execute
			req.range_end = fields[9]
			req.count = fields[13]
			req.size = fields[14]
			req.durationInMicroSec = fields[16]

			fields19Count++
		case 20:
			// etcd.log-3:2020-09-25 21:00:42.529027 I | etcdserver: read-only range request "key:\"/registry/cronjobs/\" range_end:\"/registry/cronjobs0\" limit:500 " with result "range_response_count:0 size:6" took (2.804989ms) to execute
			if fields[2] == "I" {
				// etcd.log:2020-09-25 19:24:04.010186 I | etcdserver: read-only range request "key:\"/registry/configmaps\" range_end:\"/registry/configmapt\" count_only:true " with result "range_response_count:0 size:4" took (126.195µs) to execute
				req.range_end = fields[9]
				req.count = fields[14]
				req.size = fields[15]
				req.durationInMicroSec = fields[17]

				if fields[10] == "count_only:true" {
					req.count_only = "true"
				} else {
					req.limit = fields[10]
					fieldsHasLimitCount++
				}

				fields20ICount++
			} else {
				// etcd.log:2020-09-25 19:58:45.805789 W | etcdserver: read-only range request "key:\"/registry/minions/hollow-node-qpknx\" " with result "range_response_count:1 size:2833" took too long (100.57173ms) to execute
				req.count = fields[12]
				req.size = fields[13]
				req.durationInMicroSec = fields[17]

				fields20WCount++
			}

		case 21:
			req.range_end = fields[9]
			req.durationInMicroSec = fields[18]
			if fields[2] == "I" {
				// etcd.log:2020-09-25 20:01:55.283926 I | etcdserver: read-only range request "key:\"/registry/minions/hollow-node-zz46z\\000\" range_end:\"/registry/minions0\" limit:500 revision:24335 " with result "range_response_count:1 size:6044" took (221.845µs) to execute
				req.count = fields[15]
				req.size = fields[16]

				req.limit = fields[10]
				fieldsHasLimitCount++
				fields21ICount++
			} else {
				// etcd.log-20200923-1600899026.gz:2020-09-23 21:40:29.312181 W | etcdserver: read-only range request "key:\"/registry/replicasets/kube-system/\" range_end:\"/registry/replicasets/kube-system0\" " with result "range_response_count:1 size:1655" took too long (129.790871ms) to execute
				req.count = fields[13]
				req.size = fields[14]
				fields21WCount++
			}
		case 22:
			// etcd.log:2020-09-25 19:59:40.588230 W | etcdserver: read-only range request "key:\"/registry/horizontalpodautoscalers\" range_end:\"/registry/horizontalpodautoscalert\" count_only:true " with result "range_response_count:0 size:5" took too long (155.864866ms) to execute
			req.range_end = fields[9]
			req.count = fields[14]
			req.size = fields[15]
			req.durationInMicroSec = fields[19]

			if fields[10] == "count_only:true" {
				req.count_only = "true"
			} else {
				req.limit = fields[10]
				fieldsHasLimitCount++
			}
			fields22Count++
		default:
			hasError = true
		}
		if !strings.Contains(line, "limit:") {
			fieldsNoLimitCount++
		} else if req.limit == "" {
			fmt.Printf("Missing limit in line [%s]. len(fields)=%d\n", line, len(fields))
		}

		if !hasError {
			key, err := getKey(req.key)
			if err != nil {
				fmt.Printf("Cannot parse key [%s] from line [%s]\n", req.key, line)
				hasError = true
			} else {
				req.key = key
			}

			if req.range_end != "" {
				range_end, err := getRangeEnd(req.range_end)
				if err != nil {
					fmt.Printf("Cannot parse range_end [%s] from line [%s]\n", req.range_end, line)
					hasError = true
				} else {
					req.range_end = range_end
				}
			}

			if req.limit != "" {
				limit, err := getLimit(req.limit)
				if err != nil {
					fmt.Printf("Cannot parse limit [%s] from line [%s]\n", req.limit, line)
					hasError = true
				} else {
					req.limit = limit
				}
			}

			count, err := getRangeResponseCount(req.count)
			if err != nil {
				fmt.Printf("Cannot parse count [%s] from line [%s]\n", req.count, line)
				hasError = true
			} else {
				req.count = count
			}

			size, err := getSize(req.size)
			if err != nil {
				fmt.Printf("Cannot parse size [%s] from line [%s]\n", req.size, line)
				hasError = true
			} else {
				req.size = size
			}

			durationInMicroSec, err := getDurationInNano(req.durationInMicroSec)
			if err != nil {
				fmt.Printf("Cannot parse duration [%s] from line [%s]\n", req.durationInMicroSec, line)
				hasError = true
			} else {
				req.durationInMicroSec = durationInMicroSec
			}
		}
		if hasError {
			fmt.Printf("Cannot parse [%s]\n", line)
			otherFileHandler.WriteString(line)
		} else {
			outputLine := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s\n", req.key, req.range_end, req.count_only,
				req.limit, req.count, req.size, req.durationInMicroSec)
			outputFileHandler.WriteString(outputLine)
		}
	}

	fmt.Printf("Total processed line %d, fieldCountAll %d\n", lineCount,
		fields18Count + fields19Count + fields20ICount + fields20WCount + fields21ICount + fields21WCount + fields22Count)
	fmt.Printf("Has limit %d, no limit %d, total %d. Equal? %v\n", fieldsHasLimitCount, fieldsNoLimitCount,
		fieldsHasLimitCount + fieldsNoLimitCount, fieldsHasLimitCount + fieldsNoLimitCount == lineCount)
}

func ExtractEtcdRangeLog(pathToFind string) {
	inputFile := "etcd.to.execute.range.log"
	//inputFile := "etcd.to.execute.range.log.other"

	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".compacted")
	otherFilename := path.Join(pathToFind, inputFile + ".other")
	ReadOnlyRangeRequest_Parser(inputFilename, outputFilename, otherFilename)
}

func getKey(rawKey string) (string, error) {
	// rawKey: "key:\"/registry/configmaps\"
	return getValue("\"key", rawKey)
}

func getRangeEnd(rawRange string) (string, error) {
	// raw range_end: range_end:\"/registry/configmapt\"
	return getValue("range_end", rawRange)
}

func getRangeResponseCount(rawRangeResponseCount string) (string, error) {
	// raw range response count: "range_response_count:0
	return getValue("\"range_response_count", rawRangeResponseCount)
}

func getSize(rawSize string) (string, error) {
	// raw size: size:4
	return getValue("size", rawSize)
}

func getLimit(rawLimit string) (string, error) {
	// raw limit: limit:500
	return getValue("limit", rawLimit)
}

func getValue(prefix, rawStr string) (string, error) {
	if !strings.HasPrefix(rawStr, prefix+":") {
		return "", errors.New(fmt.Sprintf("Invalid %s [%s]", prefix, rawStr))
	}

	fields := strings.Split(rawStr, ":")
	if len(fields) < 2 {
		return "", errors.New(fmt.Sprintf("Invalid %s [%s]", prefix, rawStr))
	}

	value := strings.Join(fields[1:], ":")
	if strings.HasPrefix(value,"\\\"") {
		value = value[2:len(value)-2]
	}
	if strings.HasPrefix(value, "\"") {
		value = value[1:]
	}
	if strings.HasSuffix(value, "\"") {
		value = value[:len(value)-1]
	}
	return value, nil
}

func getDurationInNano(rawDuration string) (string, error) {
	// (126.195µs)
	if strings.HasPrefix(rawDuration, "(") {
		rawDuration = rawDuration[1:len(rawDuration) -1]
	}
	timeValue, err := time.ParseDuration(rawDuration)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Invalid duration [%s]", rawDuration))
	}
	nanoSec := timeValue.Nanoseconds()
	return strconv.FormatInt(nanoSec, 10), nil
}
