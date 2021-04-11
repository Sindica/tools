package controller_log

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"tools/pkg/log_util"
)

type podSchedulingTime struct {
	podName string
	createdByRSControllerTime string
	receivedBySchedulerTime string
	addedToQueueTime string
	deQueueTime string
	boundedTime string
	boundedDuration time.Duration
}

func ExtractPodSchedulingTime(pathToFind string) {
	allPodsSchedulingTimes := extractPodCreateEventLog(pathToFind)
	extractPodSchedulingTime(allPodsSchedulingTimes, pathToFind)
}

// Get pod creation event from controller log
/*
I0409 23:36:27.556371       1 event.go:259] Event(v1.ObjectReference{Kind:"ReplicaSet", Namespace:"0pd8pj-testns", Name:"saturation-deployment-0-c47675f5", UID:"60f0ef4c-683d-4f4a-9278-c9cd19021e4d", APIVersion:"apps/v1", ResourceVersion:"9989", FieldPath:"", Tenant:"arktos"}): type: 'Normal' reason: 'SuccessfulCreate' Created pod: saturation-deployment-0-c47675f5-scn2w
 */
func extractPodCreateEventLog(pathToFind string) map[string]podSchedulingTime {
	inputFilename := path.Join(pathToFind, "controller.saturation-deployment.log")
	inputFileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputFileHandler.Close()

	allPodsSchedulingTime := make(map[string]podSchedulingTime)
	lineReader := bufio.NewReader(inputFileHandler)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			break
		}

		scheduleTime, err := log_util.GetTimeFromLog(line)
		if err != nil {
			fmt.Printf("Error getting time from log [%s]. error [%v]", line, err)
			panic(err)
		}

		//get pod name
		fields := strings.Split(line, " ")
		podname := fields[len(fields)-1]
		allPodsSchedulingTime[podname] = podSchedulingTime{
			podName: podname,
			createdByRSControllerTime: scheduleTime,
		}
	}

	return allPodsSchedulingTime
}

// Get pod scheduling time frame from customized scheduler log
/*
I0409 22:32:35.827427       1 eventhandlers.go:164] Getting pod saturation-deployment-0-c47675f5-xf258 from API server
I0409 22:32:35.827433       1 scheduling_queue.go:210] adding pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258 to the scheduling queue.
I0409 22:32:36.209513       1 scheduling_queue.go:819] About to try and schedule pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.209529       1 scheduler.go:458] Attempting to schedule pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.212748       1 generic_scheduler.go:211] DEBUG: Compute predicates pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.228494       1 generic_scheduler.go:231] DEBUG: Prioritizing pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.245901       1 generic_scheduler.go:255] DEBUG: Selecting host pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.246120       1 scheduler.go:417] Attempting to bind pod: arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258
I0409 22:32:36.248275       1 scheduler.go:596] pod arktos/1ea47i-testns/saturation-deployment-0-c47675f5-xf258 is bound successfully on node hollow-node-1-btv5d, 500 nodes evaluated, 500 nodes were found feasible
 */
func extractPodSchedulingTime(allPodsSchedulingTimes map[string]podSchedulingTime, pathToFind string) {
	//inputFilename := path.Join(pathToFind, "scheduler.saturation-deployment.log")
	for podname, sTime := range allPodsSchedulingTimes {
		fmt.Printf("pod name: %s, created time %s\n", podname, sTime.createdByRSControllerTime)
	}
}