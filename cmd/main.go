package main

import "tools/pkg/log_processor"

func main() {
	pathToFind := "/home/yinghuang/log/processing"
	//log_processor.ExtractPodSchedulingLog(pathToFind)
	log_processor.ExtractScheduledAndNonScheduledPod(pathToFind)
}
