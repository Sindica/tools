package main

import (
	"flag"
	"fmt"
	apiserver_audit_log "tools/pkg/log_processor/audit_log"
	"tools/pkg/log_processor/controller_log"
	"tools/pkg/log_processor/etcd_log"
	"tools/pkg/log_processor/trace_log"
)

func main() {
	//pathToFind := "/home/yinghuang/log/processing"
	//log_processor.ExtractPodSchedulingLog(pathToFind)
	//log_processor.ExtractScheduledAndNonScheduledPod(pathToFind)
	//log_processor.GetTimeToNano(pathToFind, "wcm-7-throttle-rs.txt", "wcm-7-throttle-rs.output")

	//parseTraceFile()

	/*
	switch mode {
	case "audit":
		parseAuditLogJsonFormat()
	case "audit_compact":
		parseCompactedAuditLog()
	default:
		fmt.Println("Invalid processing mode")
	}*/

	pathToFind := "/home/yinghuang/debug"
	controller_log.ExtractPodSchedulingTime(pathToFind)
}

func parseTraceFile() {
	//pathToFind := "/home/yinghuang/apiserver-perf/gce-500"
	pathToFind := "/home/yinghuang/apiserver-perf/xiaoning.trace.10.02"
	trace_log.ExtractTraceLog(pathToFind)
}

func parseEtcdLogFile() {
	pathToFind := "/home/yinghuang/etcd-perf/arktos-0924-760-3.4.4-perf.1-sonyaperf-load"
	//pathToFind := "/home/yinghuang/etcd-perf/arktos-0612-357-communityperf-3.4.4-perf.1/density"
	//pathToFind := "/home/yinghuang/etcd-perf/arktos-analysis/load-3.4.4"
	//pathToFind := "/home/yinghuang/etcd-perf/k8s-analysis/load-3.4.4"

	steps := 2
	switch steps {
	case 1:
		etcd_log.ExtractEtcdRangeLog(pathToFind)
		etcd_log.ExtractEtcdNoRangeLog(pathToFind)
	case 2:
		etcd_log.ParseRangeLog(pathToFind)
		fmt.Println()
		etcd_log.ParNonRangeLog(pathToFind)
	}
}

var mode string
var audit_log_filename string
var compacted_audit_log_filename string
var threadhold_xl_count int

func init() {
	flag.StringVar(&mode, "process_mode", "", "audit: process audit log\naudit_compact: process compacted audit log\n")
	flag.StringVar(&audit_log_filename, "audit_file", "", "absolute path to the audit file")
	flag.StringVar(&compacted_audit_log_filename, "audit_compacted_file", "", "absolute path to the compacted audit file")
	flag.IntVar(&threadhold_xl_count, "audit_count_xl", 10000, "extract compacted audit log to extra large request file")
	flag.Parse()
}

func parseAuditLogJsonFormat() {
	outputPath := "."
	apiserver_audit_log.ExtractAuditLog(outputPath, audit_log_filename)
}

/* Sample file:
uri, verb, response_code, count, stage
/apis/arktos.futurewei.com/v1/tenants/system/networks/default, get, 404, 2, ResponseComplete
/api/v1/tenants/system/namespaces/lodkz7-testns/secrets, list, 200, 1, ResponseComplete
/api/v1/nodes/hollow-node-54fsg, get, 200, 1, ResponseComplete
*/
func parseCompactedAuditLog() {
	apiserver_audit_log.ExtractCompactedAuditLog(".", compacted_audit_log_filename, threadhold_xl_count)
}