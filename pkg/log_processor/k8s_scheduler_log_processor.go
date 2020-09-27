package log_processor

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

var regexToFindScheduling = []string{
	logTrySchedulePod,
	"Attempting to schedule pod",
	"AssumePodVolumes for pod",
	"Attempting to bind",
	longBoundPod,
}

const logTrySchedulePod = "About to try and schedule pod"
const longBoundPod = "is bound successfully on node"

func ExtractPodSchedulingLog(pathToFind string) {
	inputFilename := path.Join(pathToFind, "kube-scheduler.log")
	outputFilename := path.Join(pathToFind, "scheduler.scheduling.pod.output")
	log_util.ExtractMatchingLines(inputFilename, outputFilename, regexToFindScheduling)
}

type schedulingTime struct {
	startTime string
	duration  time.Duration
}

func ExtractScheduledAndNonScheduledPod(pathToFind string) {
	inputFilename := path.Join(pathToFind, "scheduler.scheduling.pod.output")
	scheduledFilename := path.Join(pathToFind, "scheduler.scheduled.output")
	nonScheduledFilename := path.Join(pathToFind, "scheduler.nonscheduled.output")
	latencyScheduleFilename := path.Join(pathToFind, "scheduler.scheduled.latency.output")
	latencyToWatch := time.Duration(100 * time.Microsecond)

	inputFileHandler, err := os.Open(inputFilename)
	if err != nil {
		fmt.Printf("Error open input file [%s]: %v\n", inputFilename, err)
		panic(err)
	}
	defer inputFileHandler.Close()

	scheduledFileHandler, err := os.Create(scheduledFilename)
	if err != nil {
		fmt.Printf("Error create scheduled file [%s]: %v\n", scheduledFilename, err)
		panic(err)
	}
	defer scheduledFileHandler.Close()

	nonScheduledFileHandler, err := os.Create(nonScheduledFilename)
	if err != nil {
		fmt.Printf("Error create scheduled file [%s]: %v\n", nonScheduledFilename, err)
		panic(err)
	}
	defer nonScheduledFileHandler.Close()

	latencyScheduledFileHandler, err := os.Create(latencyScheduleFilename)
	if err != nil {
		fmt.Printf("Error create scheduled file [%s]: %v\n", latencyScheduledFileHandler, err)
		panic(err)
	}
	defer latencyScheduledFileHandler.Close()

	lineReader := bufio.NewReader(inputFileHandler)

	podToSchedule := make(map[string]*schedulingTime, 0)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		isMatched, err := regexp.MatchString(logTrySchedulePod, line)
		// I0709 01:14:22.399540       1 scheduling_queue.go:817] About to try and schedule pod system/kube-system/kubernetes-dashboard-79896fd99c-xrvq5
		// I0709 01:24:21.904119       1 scheduling_queue.go:817] About to try and schedule pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-wgt2l
		if err == nil && isMatched {
			podName := getPodFullNameFromTryScheduleLog(line)
			if podName == "" {
				fmt.Printf("Failed to get pod name from try line [%s]\n", line)
				break
			}

			_, isOK := podToSchedule[podName]
			if isOK {
				fmt.Printf("Encountered multiple try and schedule for pod. log [%s]\n", line)
				continue
			}

			scheduleTime, err := log_util.GetTimeFromLog(line)
			if err != nil {
				fmt.Printf("Error getting time from log [%s]. error [%v]", line, err)
				panic(err)
			}
			podToSchedule[podName] = &schedulingTime{
				startTime: scheduleTime,
			}
			continue
		}

		isMatched, err = regexp.MatchString(longBoundPod, line)
		// I0709 01:24:22.406258       1 scheduler.go:594] pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-dwr6h is bound successfully on node hollow-node-n8jw4, 230 nodes evaluated, 230 nodes were found feasible
		if err == nil && isMatched {
			podName := getPodFullNameFromBoundLog(line)
			if podName == "" {
				fmt.Printf("Failed to get pod name from bound line [%s]", line)
				break
			}

			record, isOK := podToSchedule[podName]
			if !isOK {
				fmt.Printf("No matching try to schedule but got bound [%s]\n", podName)
				continue
			}

			completeTime, err := log_util.GetTimeFromLog(line)
			if err != nil {
				fmt.Printf("Error getting time from log [%s]. error [%v]", line, err)
				panic(err)
			}

			duration, err := log_util.GetTimeDiff(record.startTime, completeTime)
			if err != nil {
				fmt.Printf("Error getting time difference from startTime [%s], endTime []%s. error [%v]", record.startTime, completeTime, err)
				panic(err)
			}
			podToSchedule[podName].duration = duration
		}
	}

	fmt.Printf("len of podToSchedule %d", len(podToSchedule))
	// output to files
	for podName, scheduling := range podToSchedule {
		if scheduling.duration > 0 {
			scheduledFileHandler.WriteString(fmt.Sprintf("%s, %v, Start at %v\n", podName, scheduling.duration.Nanoseconds(), scheduling.startTime))
			if scheduling.duration > latencyToWatch {
				latencyScheduledFileHandler.WriteString(fmt.Sprintf("%s, %v, Start at %v\n", podName, scheduling.duration, scheduling.startTime))
			}
		} else {
			nonScheduledFileHandler.WriteString(fmt.Sprintf("%s, Start at %v\n", podName, scheduling.startTime))
		}
	}
}

// I0709 01:24:22.406258       1 scheduler.go:594] pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-dwr6h is bound successfully on node hollow-node-n8jw4, 230 nodes evaluated, 230 nodes were found feasible
func getPodFullNameFromBoundLog(line string) string {
	strsByEmptySpace := strings.Split(line, " ")
	if len(strsByEmptySpace) <= 11 {
		return ""
	}

	return strsByEmptySpace[11]
}

// I0709 01:24:21.904119       1 scheduling_queue.go:817] About to try and schedule pod system/4m3obq-testns/saturation-deployment-0-5c568bc7fc-wgt2l
func getPodFullNameFromTryScheduleLog(line string) string {
	strsByEmptySpace := strings.Split(line, " ")
	if len(strsByEmptySpace) <= 3 {
		return ""
	}

	return strings.TrimSpace(strsByEmptySpace[len(strsByEmptySpace)-1])
}
