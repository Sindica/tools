package controller_log

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"tools/pkg/log_util"
)

type podSchedulingTime struct {
	podName string
	createdByRSControllerTime string
	receivedBySchedulerTime string
	addingToQueueTime string
	deQueueTime string
	startSchedulingTime string
	startBindingTime string
	boundedTime string
	boundedDuration time.Duration	// startBindingTime -> boundedTime
	schedulingDuration time.Duration	// startSchedulingTime -> startBindingTime
	watchedDuration time.Duration	// createdByRSControllerTime -> receivedBySchedulerTime
	queuedDuration time.Duration // addingToQueueTime -> deQueueTime
}

type durationBucket struct {
	D0_32ms int
	D32_50ms int
	D50_64ms int
	D64_128ms int
	D128_256ms int
	D256_512ms int
	D512_1s int
	D1_2s int
	D2_inf int
}

func ExtractPodSchedulingTime(pathToFind string) {
	allPodsSchedulingTimes := extractPodCreateEventLog(pathToFind)
	extractPodSchedulingTime(allPodsSchedulingTimes, pathToFind)
}

// Get pod creation event from controller log
/*
I0409 23:36:27.556371       1 event.go:259] Event(v1.ObjectReference{Kind:"ReplicaSet", Namespace:"0pd8pj-testns", Name:"saturation-deployment-0-c47675f5", UID:"60f0ef4c-683d-4f4a-9278-c9cd19021e4d", APIVersion:"apps/v1", ResourceVersion:"9989", FieldPath:"", Tenant:"arktos"}): type: 'Normal' reason: 'SuccessfulCreate' Created pod: saturation-deployment-0-c47675f5-scn2w
 */
func extractPodCreateEventLog(pathToFind string) map[string]*podSchedulingTime {
	inputFilename := path.Join(pathToFind, "controller.saturation-deployment.log")
	inputFileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputFileHandler.Close()

	allPodsSchedulingTime := make(map[string]*podSchedulingTime)
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
		allPodsSchedulingTime[podname] = &podSchedulingTime{
			podName: podname,
			createdByRSControllerTime: scheduleTime,
		}
	}

	return allPodsSchedulingTime
}

var regexToFindArktosScheduling = []string{
	"Getting pod ",
	"adding pod ",
	"About to try and schedule pod",	//dequeue
	"Attempting to schedule pod",	// start scheduling
	"Attempting to bind pod:",	//start binding
	"is bound successfully on node",	//bounded
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
func extractPodSchedulingTime(allPodsSchedulingTimes map[string]*podSchedulingTime, pathToFind string) {
	inputFilename := path.Join(pathToFind, "scheduler.saturation-deployment.log")
	inputFileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputFileHandler.Close()

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

		isMatch, caseId, podName := getMatchCase(line)
		if !isMatch {
			continue
		}

		if podName == "" {
			fmt.Printf("pod name is empty. line [%s]\n", line)
			continue
		}

		entry, isOK := allPodsSchedulingTimes[podName]
		if isOK {
			logTime, err := log_util.GetTimeFromLog(line)
			if err != nil {
				fmt.Printf("Error getting time from log [%s]. error [%v]", line, err)
				continue
			}
			switch caseId {
			case 0:
				entry.receivedBySchedulerTime = logTime
			case 1:
				entry.addingToQueueTime = logTime
			case 2:
				entry.deQueueTime = logTime
			case 3:
				entry.startSchedulingTime = logTime
			case 4:
				entry.startBindingTime = logTime
			case 5:
				entry.boundedTime = logTime
			}
		} else {
			fmt.Printf("Not found pod entry in all pods map. line [%s]\n", line)
		}
	}

	boundedDurationBucket := &durationBucket{}
	schedulingDurationBucket := &durationBucket{}
	watchedDurationBucket := &durationBucket{}
	queuedDurationBucket := &durationBucket{}
	podNamesToCheck := ""

	// calculate durations
	for podname, sTime := range allPodsSchedulingTimes {
		// boundedDuration
		if sTime.startBindingTime != "" && sTime.boundedTime != "" {
			sTime.boundedDuration, err = log_util.GetTimeDiff(sTime.startBindingTime, sTime.boundedTime)
			if err != nil {
				fmt.Printf("Error getting bound duration. pod name [%s], startBindingTime [%s], boundedTime [%s], error [%v]\n",
					podname, sTime.startBindingTime, sTime.boundedTime, err)
				break
			}
		}

		// schedulingDuration
		if sTime.startSchedulingTime != "" && sTime.startBindingTime != "" {
			sTime.schedulingDuration, err = log_util.GetTimeDiff(sTime.startSchedulingTime, sTime.startBindingTime)
			if err != nil {
				fmt.Printf("Error getting scheduling duration. pod name [%s], startSchedulingTime [%s], startSchedulingTime [%s], error [%v]\n",
					podname, sTime.startSchedulingTime, sTime.startBindingTime, err)
				break
			}
		}

		// watchedDuration
		if sTime.createdByRSControllerTime != "" && sTime.receivedBySchedulerTime != "" {
			sTime.watchedDuration, err = log_util.GetTimeDiff(sTime.createdByRSControllerTime, sTime.receivedBySchedulerTime)
			if err != nil {
				fmt.Printf("Error getting watched duration. pod name [%s], createdByRSControllerTime [%s], receivedBySchedulerTime [%s], error [%v]\n",
					podname, sTime.createdByRSControllerTime, sTime.receivedBySchedulerTime, err)
				break
			}
		}

		// queuedDuration
		if sTime.addingToQueueTime != "" && sTime.deQueueTime != "" {
			sTime.queuedDuration, err = log_util.GetTimeDiff(sTime.addingToQueueTime, sTime.deQueueTime)
			if err != nil {
				fmt.Printf("Error getting queued duration. pod name [%s], addingToQueueTime [%s], deQueueTime [%s], error [%v]\n",
					podname, sTime.addingToQueueTime, sTime.deQueueTime, err)
				break
			}
		}

		// Add into duration bucket
		isInfPod_Bound := addDurationIntoBucket(boundedDurationBucket, sTime.boundedDuration)
		isInfPod_Scheduling := addDurationIntoBucket(schedulingDurationBucket, sTime.schedulingDuration)
		isInfPod_Watch := addDurationIntoBucket(watchedDurationBucket, sTime.watchedDuration)
		isInfPod_Queued := addDurationIntoBucket(queuedDurationBucket, sTime.queuedDuration)
		if isInfPod_Bound || isInfPod_Scheduling || isInfPod_Watch || isInfPod_Queued {
			podNamesToCheck = podNamesToCheck + ", " + podname
		}
	}

	// output
	outputFilename := path.Join(pathToFind, "scheduler.saturation-deployment.log.output")
	outputFileHandler, err := os.Create(outputFilename)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputFilename, err)
		return
	}
	defer outputFileHandler.Close()

	header := "pod_name, creation_time, received_time, adding_to_queue_time, dequeue_time, start_scheduling_time, start_binding_time, bounded_time, bound_duration, scheduling_duration, watch_duration, queued_duration\n"
	outputFileHandler.WriteString(header)
	for podname, sTime := range allPodsSchedulingTimes {
		line := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s\n",
			podname, sTime.createdByRSControllerTime, sTime.receivedBySchedulerTime, sTime.addingToQueueTime,
			sTime.deQueueTime, sTime.startSchedulingTime, sTime.startBindingTime, sTime.boundedTime,
			sTime.boundedDuration, sTime.schedulingDuration, sTime.watchedDuration, sTime.queuedDuration)
		outputFileHandler.WriteString(line)
	}

	// output duration bucket
	outputBucketFilename := path.Join(pathToFind, "scheduler.saturation-deployment.log.bucket")
	outputBucketFileHandler, err := os.Create(outputBucketFilename)
	if err != nil {
		fmt.Printf("Error open output file [%s]: %v\n", outputBucketFilename, err)
		return
	}
	defer outputBucketFileHandler.Close()
	header = "case, <=32ms, <=50ms, <=64ms, <=128ms, <=256ms, <=512ms, <=1s, <=2s, 2-inf\n"
	outputBucketFileHandler.WriteString(header)
	line := "Bound duration, " + getBucketOutputLine(boundedDurationBucket)
	outputBucketFileHandler.WriteString(line)
	line = "Scheduling duration, " + getBucketOutputLine(schedulingDurationBucket)
	outputBucketFileHandler.WriteString(line)
	line = "Watched duration, " + getBucketOutputLine(watchedDurationBucket)
	outputBucketFileHandler.WriteString(line)
	line = "Queued duration, " + getBucketOutputLine(queuedDurationBucket)
	outputBucketFileHandler.WriteString(line)

	// output infinity pod names
	outputBucketFileHandler.WriteString(fmt.Sprintf("infinity pods: %s\n", podNamesToCheck))
}

func getMatchCase(line string) (isMatch bool, caseId int, podName string) {
	isMatch = false
	for i, exp := range regexToFindArktosScheduling {
		isMatched, err := regexp.MatchString(exp, line)
		if err == nil && isMatched {
			podName = getPodNameFromArktosSchedulerLog(line, i)
			isMatch = true
			caseId = i
			return
		}
	}
	return
}

func getPodNameFromArktosSchedulerLog(line string, caseId int) string {
	fields := strings.Split(line, " ")

	if caseId == 0 {
		return fields[len(fields) - 4]
	}
	fullPodName := ""
	switch caseId {
	case 1:
		fullPodName = fields[len(fields) - 5]
	case 2, 3, 4:
		fullPodName = fields[len(fields) - 1]
	case 5:
		fullPodName = fields[len(fields) - 15]
	default:
		return ""
	}
	fields = strings.Split(fullPodName, "/")
	return fields[2]
}

func addDurationIntoBucket(durBucket *durationBucket, duration time.Duration) (isInf bool) {
	isInf = false

	if duration <= time.Millisecond * 32 {
		durBucket.D0_32ms++
		return
	}
	if duration <= time.Millisecond * 50 {
		durBucket.D32_50ms++
		return
	}
	if duration <= time.Millisecond * 64 {
		durBucket.D50_64ms++
		return
	}
	if duration <= time.Millisecond * 128 {
		durBucket.D64_128ms++
		return
	}
	if duration <= time.Millisecond * 256 {
		durBucket.D128_256ms++
		return
	}
	if duration <= time.Millisecond * 512 {
		durBucket.D256_512ms++
		return
	}
	if duration <= time.Second {
		durBucket.D512_1s++
		return
	}
	if duration <= time.Second * 2 {
		durBucket.D1_2s++
		return
	}
	durBucket.D2_inf++
	isInf = true
	return
}

func getBucketOutputLine(durBucket *durationBucket) string {
	return fmt.Sprintf("%d, %d, %d, %d, %d, %d, %d, %d, %d\n",
		durBucket.D0_32ms, durBucket.D32_50ms, durBucket.D50_64ms, durBucket.D64_128ms,
		durBucket.D128_256ms, durBucket.D256_512ms, durBucket.D512_1s, durBucket.D1_2s,
		durBucket.D2_inf)
}