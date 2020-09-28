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

	// print header
	outputLine := fmt.Sprintf("key, rang_end, is_count_only, limit, range_response_count, size, duration\n")
	outputFileHandler.WriteString(outputLine)

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

type NoRangeRequest struct {
	key string
	method string
	mod_revision string
	success_key string	// compare only
	success_method string
	success_value_size string
	failure_key string // compare only
	failure_method string
	size string
	durationInMicroSec string
}

func NoReadOnlyRangeRequest_Parser(inputFileName, outputFileName, nonMatchingFilename string) {
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
	errorCount := 0

	// print header
	outputLine := fmt.Sprintf("key, method, revision, success_method, success_value_size, failure_method, size, duration\n")
	outputFileHandler.WriteString(outputLine)

	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		lineCount++
		hasError, outputLine := NoReadOnlyRangeRequest(line)

		if !hasError {

			outputFileHandler.WriteString(outputLine)
		} else {
			fmt.Printf("Cannot parse [%s]\n", line)
			otherFileHandler.WriteString(line)
			errorCount++
		}
	}
	fmt.Printf("Total line %d, error line count %d\n", lineCount, errorCount)
}

func NoReadOnlyRangeRequest(line string) (bool, string) {
	fields := strings.Split(line, " ")
	req := NoRangeRequest{
	}

	hasError := false
	method, err := getMethod(fields[7])
	action_arktos, err1 := getAction(fields[8])
	action_k8s, err2 := getAction(fields[10])
	var durationInMicroSec string
	var size string

	if err == nil {
		// etcd.log:2020-09-22 18:12:05.129283 I | etcdserver: request "ID:10592876127946485506 Method:\"PUT\" Path:\"/0/members/4faa637bfd19301/attributes\" Val:\"{\\\"name\\\":\\\"etcd-ying3-kubemark-compare-10-kubemark-master\\\",\\\"clientURLs\\\":[\\\"http://127.0.0.1:2379\\\",\\\"https://10.40.0.11:2379\\\"]}\" " with result "" took (113.908µs) to execute
		req.method = method
		if len(fields) == 16 {
			// etcd.log:2020-09-22 18:12:47.108193 I | etcdserver: request "ID:10592876127946486575 Method:\"QGET\" " with result "" took (12.789µs) to execute
			durationInMicroSec = fields[13]
		} else {
			req.key, err = getPath(fields[8])
			if err != nil {
				fmt.Printf("Cannot parse path [%s] from line [%s]\n", fields[8], line)
				hasError = true
			}

			// etcd.log:2020-09-22 18:12:05.129283 I | etcdserver: request "ID:10592876127946485506 Method:\"PUT\" Path:\"/0/members/4faa637bfd19301/attributes\" Val:\"{\\\"name\\\":\\\"etcd-ying3-kubemark-compare-10-kubemark-master\\\",\\\"clientURLs\\\":[\\\"http://127.0.0.1:2379\\\",\\\"https://10.40.0.11:2379\\\"]}\" " with result "" took (113.908µs) to execute
			durationInMicroSec = fields[15]
		}
	} else if err1 == nil && action_arktos != "" {
		// arktos:
		// etcd.log:2020-09-22 18:12:26.540548 I | etcdserver: request "header:<ID:10592876127946486339 > lease_revoke:<id:130174b703f730e4>" with result "size:27" took (65.923µs) to execute
		// etcd.log:2020-09-22 18:12:11.177562 I | etcdserver: request "header:<ID:10592876127946485989 > lease_grant:<ttl:15-second id:130174b703f730e4>" with result "size:39" took (149.606µs) to execute
		// etcd.log-20200922-1600800918.gz:2020-09-22 18:21:59.761339 I | etcdserver: request "header:<ID:10592876127946500926 > compaction:<revision:1000 > " with result "size:5" took (2.337296ms) to execute

		req.method = action_arktos

		switch req.method {
		case "lease_grant":
			size = fields[12]
			durationInMicroSec = fields[14]
		case "lease_revoke":
			size = fields[11]
			durationInMicroSec = fields[13]
		case "compaction":
			size = fields[13]
			durationInMicroSec = fields[15]
		default:
			hasError = true
		}
	} else if err2 == nil && action_k8s != "" {
		// k8s:
		// etcd.log:2020-09-25 19:24:09.781338 I | etcdserver: request "header:<ID:10636223341819266000 username:\"client\" auth_revision:1 > lease_grant:<ttl:15-second id:139b74c6b8db03cf>" with result "size:40" took (124.69µs) to execute
		req.method = action_k8s

		switch req.method {
		case "lease_grant":
			size = fields[14]
			durationInMicroSec = fields[16]
		case "lease_revoke":
			size = fields[13]
			durationInMicroSec = fields[15]
		case "compaction":
			size = fields[15]
			durationInMicroSec = fields[17]
		default:
			hasError = true
		}
	} else {	// no method:
		var key string
		var mod_revision string
		var fieldSuccess string
		var success_value_size string
		var fieldFailure string

		switch len(fields) {
		case 24:
			// 24: etcd.log:2020-09-22 18:16:59.757018 I | etcdserver: request "header:<ID:10592876127946488796 > txn:<compare:<key:\"compact_rev_key\" version:0 > success:<request_put:<key:\"compact_rev_key\" value_size:1 >> failure:<request_range:<key:\"compact_rev_key\" > >>" with result "size:16" took (151.315µs) to execute
			key = fields[8]
			mod_revision = fields[9]
			fieldSuccess = fields[11]
			success_value_size = fields[12]
			fieldFailure = fields[14]
			size = fields[19]
			durationInMicroSec = fields[21]

		case 28:
			// 28-k8s:    etcd.log:2020-09-25 19:24:09.783370 I | etcdserver: request "header:<ID:10636223341819266001 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/masterleases/10.40.0.12\" mod_revision:0 > success:<request_put:<key:\"/registry/masterleases/10.40.0.12\" value_size:65 lease:1412851304964490191 >> failure:<request_range:<key:\"/registry/masterleases/10.40.0.12\" > >>" with result "size:16" took (123.553µs) to execute
			key = fields[11]
			mod_revision = fields[12]
			fieldSuccess = fields[14]
			success_value_size = fields[15]
			fieldFailure = fields[18]
			size = fields[23]
			durationInMicroSec = fields[25]
		case 29:
			// k8s: etcd.log-20200925-1601065806.gz:2020-09-25 20:14:05.657635 W | etcdserver: request "header:<ID:10636223341819382712 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/pods/test-6vwfq1-3/saturation-deployment-0-7675bfc6b9-76zbj\" mod_revision:80621 > success:<request_put:<key:\"/registry/pods/test-6vwfq1-3/saturation-deployment-0-7675bfc6b9-76zbj\" value_size:2608 >> failure:<request_range:<key:\"/registry/pods/test-6vwfq1-3/saturation-deployment-0-7675bfc6b9-76zbj\" > >>" with result "size:18" took too long (140.245474ms) to execute
			key = fields[11]
			mod_revision = fields[12]
			fieldSuccess = fields[14]
			success_value_size = fields[15]
			fieldFailure = fields[17]
			size = fields[22]
			durationInMicroSec = fields[26]

		case 26:
			// 26-arktos: etcd.log:2020-09-22 18:12:11.180111 I | etcdserver: request "header:<ID:10592876127946485990 > txn:<compare:<target:MOD key:\"/registry/masterleases/10.40.0.11\" mod_revision:0 > success:<request_put:<key:\"/registry/masterleases/10.40.0.11\" value_size:74 lease:1369504091091710180 >> failure:<request_range:<key:\"/registry/masterleases/10.40.0.11\" > >>" with result "size:16" took (1.415049ms) to execute
			// 26-k8s:    etcd.log:2020-09-25 19:29:03.898829 I | etcdserver: request "header:<ID:10636223341819269789 username:\"client\" auth_revision:1 > txn:<compare:<key:\"compact_rev_key\" version:0 > success:<request_put:<key:\"compact_rev_key\" value_size:1 >> failure:<request_range:<key:\"compact_rev_key\" > >>" with result "size:16" took (166.551µs) to execute
			key = fields[9]
			_, err := getKey(key)
			if err != nil {
				// 26-k8s:    etcd.log:2020-09-25 19:29:03.898829 I | etcdserver: request "header:<ID:10636223341819269789 username:\"client\" auth_revision:1 > txn:<compare:<key:\"compact_rev_key\" version:0 > success:<request_put:<key:\"compact_rev_key\" value_size:1 >> failure:<request_range:<key:\"compact_rev_key\" > >>" with result "size:16" took (166.551µs) to execute
				key = fields[10]
				mod_revision = fields[11]
				fieldSuccess = fields[13]
				success_value_size = fields[14]
			} else {
				// 26-arktos: etcd.log:2020-09-22 18:12:11.180111 I | etcdserver: request "header:<ID:10592876127946485990 > txn:<compare:<target:MOD key:\"/registry/masterleases/10.40.0.11\" mod_revision:0 > success:<request_put:<key:\"/registry/masterleases/10.40.0.11\" value_size:74 lease:1369504091091710180 >> failure:<request_range:<key:\"/registry/masterleases/10.40.0.11\" > >>" with result "size:16" took (1.415049ms) to execute
				mod_revision = fields[10]
				fieldSuccess = fields[12]
				success_value_size = fields[13]
			}
			fieldFailure = fields[16]
			size = fields[21]
			durationInMicroSec = fields[23]

		case 23, 25, 27:
			// 27-k8s: etcd.log:2020-09-25 19:24:07.878007 I | etcdserver: request "header:<ID:10636223341819265677 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/ranges/serviceips\" mod_revision:0 > success:<request_put:<key:\"/registry/ranges/serviceips\" value_size:68 >> failure:<request_range:<key:\"/registry/ranges/serviceips\" > >>" with result "size:14" took (205.359µs) to execute

			// 25-arktos: etcd.log:2020-09-22 18:12:05.457308 I | etcdserver: request "header:<ID:10592876127946485567 > txn:<compare:<target:MOD key:\"/registry/ranges/serviceips\" mod_revision:0 > success:<request_put:<key:\"/registry/ranges/serviceips\" value_size:74 >> failure:<request_range:<key:\"/registry/ranges/serviceips\" > >>" with result "size:14" took (362.41µs) to execute
			// 25-k8s:    etcd.log:2020-09-25 19:24:07.860705 I | etcdserver: request "header:<ID:10636223341819265670 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" mod_revision:0 > success:<request_put:<key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" value_size:495 >> failure:<>>" with result "size:14" took (305.488µs) to execute

			// 23: etcd.log:2020-09-22 18:12:08.563601 I | etcdserver: request "header:<ID:10592876127946485651 > txn:<compare:<target:MOD key:\"/registry/namespaces/kube-system\" mod_revision:0 > success:<request_put:<key:\"/registry/namespaces/kube-system\" value_size:146 >> failure:<>>" with result "size:14" took (106.823µs) to execute
			key = fields[9]
			mod_revision = fields[10]
			fieldSuccess = fields[12]
			success_value_size = fields[13]

		    caseId := 1
			if len(fields) == 25 || len(fields) == 27 {
				_, err := getKey(key)
				if err != nil {
					// 25-k8s:    etcd.log:2020-09-25 19:24:07.860705 I | etcdserver: request "header:<ID:10636223341819265670 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" mod_revision:0 > success:<request_put:<key:\"/registry/apiregistration.k8s.io/apiservices/v1.\" value_size:495 >> failure:<>>" with result "size:14" took (305.488µs) to execute
					// 27-k8s: etcd.log:2020-09-25 19:24:07.878007 I | etcdserver: request "header:<ID:10636223341819265677 username:\"client\" auth_revision:1 > txn:<compare:<target:MOD key:\"/registry/ranges/serviceips\" mod_revision:0 > success:<request_put:<key:\"/registry/ranges/serviceips\" value_size:68 >> failure:<request_range:<key:\"/registry/ranges/serviceips\" > >>" with result "size:14" took (205.359µs) to execute					key = fields[11]
					caseId = 2
					key = fields[11]
					mod_revision = fields[12]
					fieldSuccess = fields[14]
					success_value_size = fields[15]
					fieldFailure = fields[17]
				} else {
					fieldFailure = fields[15]
				}
			}

			if len(fields) == 23 {
				fieldFailure = fields[15]
			}

			if len(fields) == 23 {
				size = fields[18]
				durationInMicroSec = fields[20]
			} else if len(fields) == 25 {
				if fields[2] == "I" {
					size = fields[20]
				} else {	// W
					size = fields[18]
				}
				durationInMicroSec = fields[22]
				_, err = getValueSize(success_value_size)
				if err != nil {
					success_value_size = ""
				}
			} else if len(fields) == 27 {
				if caseId == 1 {
					size = fields[20]
				} else if caseId == 2 {
					size = fields[22]
				}
				durationInMicroSec = fields[24]
			}
		default:
			hasError = true
		}

		if !hasError {
			req.key, err = getKey(key)
			if err != nil {
				fmt.Printf("Cannot parse key [%s] from line [%s]\n", key, line)
				hasError = true
			}
			req.mod_revision, err = getModRevision(mod_revision)
			if err != nil {
				fmt.Printf("Cannot parse mode_revision [%s] from line [%s]\n", mod_revision, line)
				hasError = true
			}

			req.success_method, req.success_key, err = getSuccessMethodAndKey(fieldSuccess)
			if err != nil {
				fmt.Printf("Cannot parse sucess segment [%s] from line [%s]\n", fieldSuccess, line)
				hasError = true
			}
			req.success_value_size, err = getValueSize(success_value_size)
			if err != nil {
				fmt.Printf("Cannot parse value_size [%s] from line [%s]\n", success_value_size, line)
				hasError = true
			}
			req.failure_method, req.failure_key, err = getFailureMethodAndKey(fieldFailure)
			if err != nil {
				fmt.Printf("Cannot parse failure segment [%s] from line [%s]\n", fieldFailure, line)
				hasError = true
			}
			if req.key != req.success_key || req.failure_key != "" && req.key != req.failure_key {
				fmt.Printf("key/success_key/failure_key: %s/%s/%s not equal in line [%s]\n",
					req.key, req.success_key, req.failure_key, line)
				hasError = true
			}

			req.size, err = getSize(size)
			if err != nil {
				fmt.Printf("Cannot parse size [%s] from line [%s]\n", size, line)
				hasError = true
			}
		}
	}

	if !hasError {
		req.durationInMicroSec, err = getDurationInNano(durationInMicroSec)
		if err != nil {
			fmt.Printf("Cannot parse duration [%s] from line [%s]\n", durationInMicroSec, line)
			hasError = true
		}
	}

	if !hasError {
		outputLine := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s\n", req.key, req.method, req.mod_revision,
			req.success_method, req.success_value_size, req.failure_method, req.size, req.durationInMicroSec)
		return false, outputLine
	} else {
		return true, ""
	}
}

func ExtractEtcdRangeLog(pathToFind string) {
	inputFile := "etcd.to.execute.range.log"
	//inputFile := "etcd.to.execute.range.log.other"

	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".compacted")
	otherFilename := path.Join(pathToFind, inputFile + ".other")
	ReadOnlyRangeRequest_Parser(inputFilename, outputFilename, otherFilename)
}

func ExtractEtcdNoRangeLog(pathToFind string) {
	inputFile := "etcd.to.execute.norange.log"
	//inputFile = inputFile + ".other.other"

	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".compacted")
	otherFilename := path.Join(pathToFind, inputFile + ".other")
	NoReadOnlyRangeRequest_Parser(inputFilename, outputFilename, otherFilename)
}

func getKey(rawKey string) (string, error) {
	// rawKey: "key:\"/registry/configmaps\"
	key, err := getValue("key", trimValue(rawKey))
	if err == nil {
		return key, nil
	}

	// txn:<compare:<key:\"compact_rev_key\"
	if strings.Contains(rawKey, "<key:") {
		fields := strings.Split(rawKey, "<key:")
		if len(fields) != 2 && fields[1] != "" {
			return "", err
		}
		return trimValue(fields[1]), nil
	}
	return "", err
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
	return getValue("size", trimValue(rawSize))
}

func getLimit(rawLimit string) (string, error) {
	// raw limit: limit:500
	return getValue("limit", rawLimit)
}

func getMethod(rawMethod string) (string, error) {
	// raw method: Method:\"PUT\"
	return getValue("Method", rawMethod)
}

func getAction(rawLeaseAction string) (string, error) {
	leaseGrantAction := "lease_grant"
	leaseRevokeAction := "lease_revoke"
	compactionAction := "compaction"
	// lease_grant:<ttl:15-second id:130174b703f730e4>"
	_, err := getValue(leaseGrantAction, trimValue(rawLeaseAction))

	if err == nil {
		return leaseGrantAction, nil
	}

	// lease_revoke:<id:130174b703f730e4>"
	_, err = getValue(leaseRevokeAction, trimValue(rawLeaseAction))
	if err == nil {
		return leaseRevokeAction, nil
	}

	// compaction:<revision:1000
	_, err = getValue(compactionAction, trimValue(rawLeaseAction))
	return compactionAction, err
}

func getPath(rawPath string) (string, error) {
	// raw path: Path:\"/0/members/4faa637bfd19301/attributes\"
	return getValue("Path", rawPath)
}

func getModRevision(rawModRev string) (string, error) {
	// mod_revision:0
	value, err := getValue("mod_revision", rawModRev)

	if err == nil {
		return value, nil
	}

	// version:0
	return getValue("version", rawModRev)
}

// optional
func getValueSize(rawValueSize string) (string, error) {
	if !strings.Contains(rawValueSize, "value_size:") {
		return "", nil
	}

	// value_size:74
	return getValue("value_size", rawValueSize)
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
	value = trimValue(value)
	return value, nil
}

func trimValue(value string) string {
	if strings.HasPrefix(value,"\\\"") {
		value = value[2:len(value)]
	}
	if strings.HasPrefix(value, "\"") {
		value = value[1:]
	}
	if strings.HasSuffix(value, "\"") {
		value = value[:len(value)-1]
	}
	if strings.HasPrefix(value, "<") {
		value = value[1:]
	}
	if strings.HasSuffix(value, ">") {
		value = value[:len(value)-1]
	}

	return value
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

func getSuccessMethodAndKey(rawSucessSeg string) (string, string, error) {
	// success:<request_put:<key:\"/registry/ranges/serviceips\"
	return getMethodAndKey("success", rawSucessSeg)
}

func getFailureMethodAndKey(rawFailureSeg string) (string, string, error) {
	// failure:<request_range:<key:\"/registry/ranges/serviceips\"
	return getMethodAndKey("failure", trimValue(rawFailureSeg))
}

func getMethodAndKey(prefix, rawStr string) (string, string, error) {
	if !strings.HasPrefix(rawStr, prefix+":") {
		return "", "", errors.New(fmt.Sprintf("Invalid %s [%s]", prefix, rawStr))
	}

	if rawStr == prefix+":<>" {
		return "", "", nil
	}

	fields := strings.Split(rawStr, ":")
	if len(fields) < 4 {
		return "", "", errors.New(fmt.Sprintf("Invalid %s [%s]", prefix, rawStr))
	}

	method := trimValue(fields[1])
	key := trimValue(strings.Join(fields[3:], ":"))
	return method, key, nil
}