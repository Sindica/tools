package trace_log

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type Trace struct {
	traceId       string
	totalDuration string
	startTime     string
	steps         []TraceStep
	wasCompleted  bool
}

type TraceStep struct {
	traceId string
	stepMessage string
	stepDuration string
	totalDuration string
	startTime string
	isStart bool
	isEnd bool
}

func ExtractTraceLog(pathToFind string) {
	inputFile := "apiserver.Trace"
	inputFilename := path.Join(pathToFind, inputFile)
	outputFilename := path.Join(pathToFind, inputFile + ".compacted")
	errFilename := path.Join(pathToFind, inputFile + ".errortrace")

	Trace_Parser(inputFilename, outputFilename, errFilename)
}

func Trace_Parser(inputFileName, outputFileName, nonMatchingFilename string) {
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

	lineReader := bufio.NewReader(inputfileHandler)
	lineCount := 0
	traceCount := 0
	completedTrace := 0
	incompleteTrace :=0
	outputLine := fmt.Sprintf("trace_id, is_completed, total_duration, start_time, steps\n")
	outputFileHandler.WriteString(outputLine)

	otherFileHandler, err := os.Create(nonMatchingFilename)
	if err != nil {
		fmt.Printf("Error open non matching file [%s]: %v\n", nonMatchingFilename, err)
		panic(err)
	}
	defer otherFileHandler.Close()

	traces := make(map[string]*Trace)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error read file by line: %v\n", err)
			break
		}

		lineCount++
		step := ParseStep(line)
		hasError := false
		errMsg := ""
		if step.isStart {
			traceStart := TraceStep{
				traceId: step.traceId,
				startTime: step.startTime,
				stepMessage: step.stepMessage,
				isStart: true,
			}

			traces[step.traceId] = &Trace{
				traceId:       step.traceId,
				totalDuration: step.totalDuration,
				startTime:     step.startTime,
				steps:         []TraceStep{traceStart},
			}
		} else {
			currentTrace, isOK := traces[step.traceId]
			if isOK {
				if currentTrace.wasCompleted {
					hasError = true
					errMsg = "Duplicated trace end"
				} else {
					if !step.isEnd {
						traceStep := TraceStep{
							traceId: step.traceId,
							stepDuration: step.stepDuration,
							stepMessage: step.stepMessage,
						}
						currentTrace.steps = append(currentTrace.steps, traceStep)
					} else {
						currentTrace.wasCompleted = true
						if currentTrace.totalDuration != step.totalDuration {
							hasError = true
							errMsg = fmt.Sprintf("Inconsistent total duration %s/%s", currentTrace.totalDuration, step.totalDuration)
						} else {
							traceEnd := TraceStep{
								traceId: step.traceId,
								stepMessage: step.stepMessage,
								stepDuration: step.stepDuration,
								isEnd: true,
							}

							currentTrace.steps = append(currentTrace.steps, traceEnd)
							// output current trace
							outputLine = getTraceOutput(currentTrace)
							outputFileHandler.WriteString(outputLine)

							// remove trace from map
							delete(traces, step.traceId)
							traceCount++
							completedTrace++
						}
					}
				}
			} else {
				hasError = true
				errMsg = "Trace end does not have matching start"
			}
		}

		if hasError {
			otherFileHandler.WriteString(fmt.Sprintf("%s, %s\n", step.traceId, errMsg))
		}
	}

	for _, v := range traces {
		v.wasCompleted = true
		outputLine = getTraceOutput(v)
		outputFileHandler.WriteString(outputLine)
		traceCount++
		incompleteTrace++
	}
	fmt.Printf("Total line %d, traces %d, completed trace %d, incomplete trace %d\n", lineCount, traceCount, completedTrace, incompleteTrace)
}

func ParseStep(line string) TraceStep {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic parsing line [%s]. error [%v]\n", line, r)
		}
	}()

	fields := strings.Split(line, " ")
	fieldCount := len(fields)
	step := TraceStep{}
	var err error
	var traceIdPos int
	if fieldCount >= 15 && fields[0][len(fields[0])-5:len(fields[0])-4] == "I" && fields[fieldCount-3] == "(total" && fields[fieldCount-2] == "time:" {
		// is start
		step.isStart = true

		durationValue := strings.TrimSpace(fields[fieldCount-1])
		step.totalDuration = durationValue[:len(durationValue)-2]
		step.stepMessage = strings.Join(fields[11:fieldCount-9], " ")
		step.startTime = strings.Join(fields[fieldCount-8:fieldCount-6], " ")
		traceIdPos = 10
	} else if fieldCount == 4 && strings.TrimSpace(fields[3]) == "END" {
		// is end
		step.isEnd = true
		traceIdPos = 0

		durationValue := fields[2]
		step.stepDuration = durationValue[1:len(durationValue)-1]

		durationValue = fields[1]
		step.totalDuration = durationValue[1:len(durationValue)-1]
		step.stepMessage = "END"
	} else {
		step.stepMessage = strings.TrimSpace(strings.Join(fields[3:], " "))
		traceIdPos = 0

		durationValue := fields[2]
		step.stepDuration = durationValue[1:len(durationValue)-1]
	}
	step.traceId, err = getTraceId(fields[traceIdPos])
	if err != nil {
		fmt.Sprintf("Error parsing line [%s]. Error [%v]", line, err)
	}

	return step
}

func getTraceOutput(trace *Trace) string {
	message := fmt.Sprintf("%s, %t, %s, %s, ", trace.traceId, trace.wasCompleted, getDurationInMicroSecond(trace.totalDuration), trace.startTime)
	for i:=0; i<len(trace.steps); i++ {
		step := trace.steps[i]
		if step.isStart {
			message += fmt.Sprintf("%s, ", step.stepMessage)
		} else if step.isEnd {
			message += fmt.Sprintf("END, %s", getDurationInMicroSecond(step.stepDuration))
		} else {
			message += fmt.Sprintf("%s, %s, ", step.stepMessage, getDurationInMicroSecond(step.stepDuration))
		}
	}

	message += "\n"
	return message
}

func getDurationInMicroSecond(durationValue string) string {
	timeValue, err := time.ParseDuration(durationValue)
	if err != nil {
		fmt.Printf("Error getting time duration: value %s, error %v.\n", durationValue, err)
		return ""
	} else {
		return fmt.Sprintf("%f", float64(timeValue.Nanoseconds())/1000)
	}
}

func getTraceId(traceIdField string) (string, error) {
	index1 := strings.Index(traceIdField, "Trace[")
	index2 := strings.Index(traceIdField, "]:")
	if index1 == -1 || index2 == -1 || index1 > index2 {
		return "", errors.New(fmt.Sprintf("Invalid trace id field %s", traceIdField))
	}

	return traceIdField[index1+6:index2], nil
}