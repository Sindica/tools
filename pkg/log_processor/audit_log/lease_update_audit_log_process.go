package apiserver_audit_log

import (
	"fmt"
	"os"
	"path"
	"strings"
	"tools/pkg/log_util"
)

func ExtractLeaseUpdateAuditLog(outputPath string, inputFilename string) {
	filenameShort := log_util.GetFilenameOnly(inputFilename)
	outputFilename := path.Join(outputPath, "lease-update-"+filenameShort)
	errorAuditLogFilename := path.Join(outputPath, "error-entry-"+filenameShort)
	ProcessLeaseUpdateAuditLog(inputFilename, outputFilename, errorAuditLogFilename)
}

func ProcessLeaseUpdateAuditLog(inputFilename, outputFilename, errorAuditLogFilename string) {
	outputFileHandler, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename, err)
		return
	}
	defer outputFileHandler.Close()

	errAuditFileHandler, err := os.Create(errorAuditLogFilename)
	if err != nil {
		fmt.Printf("Error open unparserable audit log file [%s]: %v\n", errorAuditLogFilename, err)
		return
	}
	defer errAuditFileHandler.Close()

	processLeaseUpdateAuditLog(inputFilename, outputFileHandler, errAuditFileHandler)
}

func processLeaseUpdateAuditLog(inputFilename string, outputFileHandler, errAuditFileHandler *os.File) {
	apiServerAuditLog, err := readAuditLog(inputFilename, errAuditFileHandler)
	if err != nil {
		fmt.Printf("Error reading audit log: %v\n", err)
		return
	}

	// Get only verb=update, resource=leases
	// map update leases request to time (previous to second for now) -> count
	leaseDistributedMap := make(map[string]int, 0)
	for _, log := range apiServerAuditLog {
		if log.Verb == "update" && strings.HasPrefix(log.RequestURI , "/apis/coordination.k8s.io/v1beta1/tenants/system/namespaces/kube-node-lease/leases/hollow-node-") {
			dt := log_util.GetDateTime(log.RequestReceivedTimeStamp, 19)
			if count, isOK := leaseDistributedMap[dt]; isOK {
				leaseDistributedMap[dt] = count + 1
			} else {
				leaseDistributedMap[dt] = 1
			}
		}
	}

	// print out count
	header := "datetime, count\n"
	outputFileHandler.WriteString(header)
	for dt, count := range leaseDistributedMap {
		line := fmt.Sprintf("%s, %d\n", dt, count)
		outputFileHandler.WriteString(line)
	}
}