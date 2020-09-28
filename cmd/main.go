package main

import (
	"tools/pkg/log_processor/etcd_log"
)

func main() {
	//pathToFind := "/home/yinghuang/log/processing"
	//log_processor.ExtractPodSchedulingLog(pathToFind)
	//log_processor.ExtractScheduledAndNonScheduledPod(pathToFind)
	//log_processor.GetTimeToNano(pathToFind, "wcm-7-throttle-rs.txt", "wcm-7-throttle-rs.output")

	pathToFind := "/home/yinghuang/etcd-perf/k8s-analysis/load-3.4.4"
	//pathToFind := "/home/yinghuang/etcd-perf/arktos-analysis/load-3.4.4"
	etcd_log.ExtractEtcdRangeLog(pathToFind)
	etcd_log.ExtractEtcdNoRangeLog(pathToFind)
}
