package main

import (
	"fmt"
	"tools/pkg/log_processor/etcd_log"
)

func main() {
	//pathToFind := "/home/yinghuang/log/processing"
	//log_processor.ExtractPodSchedulingLog(pathToFind)
	//log_processor.ExtractScheduledAndNonScheduledPod(pathToFind)
	//log_processor.GetTimeToNano(pathToFind, "wcm-7-throttle-rs.txt", "wcm-7-throttle-rs.output")

	pathToFind := "/home/yinghuang/etcd-perf/arktos-0924-760-3.4.4-perf.1"
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
