package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hornbill/color"
	apiLib "github.com/hornbill/goApiLib"
)

func initVars() {
	//-- Start Time for Durration
	startTime = time.Now()
	//-- Start Time for Log File
	timeNow = time.Now().Format(time.RFC3339)
	//-- Remove :
	timeNow = strings.Replace(timeNow, ":", "-", -1)
	//-- Set Counter
	errorCount = 0
}

// =================== COUNTERS =================== //
func errorCountInc() {
	mutexCounters.Lock()
	errorCount++
	mutexCounters.Unlock()
}
func updateCountInc() {
	mutexCounters.Lock()
	counters.updated++
	mutexCounters.Unlock()
}
func createCountInc() {
	mutexCounters.Lock()
	counters.created++
	mutexCounters.Unlock()
}

func loggerWriteBuffer(s string) {
	if s != "" {
		logLines := strings.Split(s, "\n\r")
		for _, line := range logLines {
			if line != "" {
				logger(0, line, false)
			}
		}
	}
}

//-- Worker Pool Function
func loggerGen(t int, s string) string {

	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 0:
		errorLogPrefix = ""
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	currentTime := time.Now().UTC()
	time := currentTime.Format("2006/01/02 15:04:05")
	return time + " " + errorLogPrefix + s + "\n"
}

//-- Logging function
func logger(t int, s string, outputtoCLI bool) {
	//-- Curreny WD
	cwd, _ := os.Getwd()
	//-- Log Folder
	logPath := cwd + "/log"
	//-- Log File
	logFileName := logPath + "/" + configLogPrefix + "SQL_Contact_Import_" + timeNow + ".log"
	red := color.New(color.FgRed).PrintfFunc()
	orange := color.New(color.FgCyan).PrintfFunc()
	//-- If Folder Does Not Exist then create it
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0777)
		if err != nil {
			fmt.Printf("Error Creating Log Folder %q: %s \r", logPath, err)
			os.Exit(101)
		}
	}

	//-- Open Log File
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		fmt.Printf("Error Creating Log File %q: %s \n", logFileName, err)
		os.Exit(100)
	}
	// don't forget to close it
	defer f.Close()
	// assign it to the standard logger
	logFileMutex.Lock()
	log.SetOutput(f)
	logFileMutex.Unlock()
	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 0:
		errorLogPrefix = ""
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	if outputtoCLI {
		if t == 3 {
			orange(errorLogPrefix + s + "\n")
		} else if t == 4 {
			red(errorLogPrefix + s + "\n")
		} else {
			fmt.Printf(errorLogPrefix + s + "\n")
		}

	}
	log.Println(errorLogPrefix + s)
}

//-- Log to ESP
func espLogger(message string, severity string) bool {
	espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.InstanceID)
	espXmlmc.SetAPIKey(SQLImportConf.APIKey)
	espXmlmc.SetParam("fileName", "SQL_Contact_Import")
	espXmlmc.SetParam("group", "general")
	espXmlmc.SetParam("severity", severity)
	espXmlmc.SetParam("message", message)

	XMLLogger, xmlmcErr := espXmlmc.Invoke("system", "logMessage")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", xmlmcErr), true)
		return false
	}
	err := xml.Unmarshal([]byte(XMLLogger), &xmlRespon)
	if err != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", err), true)
		return false
	}
	if xmlRespon.MethodResult != constOK {
		logger(4, "Unable to write to log "+xmlRespon.State.ErrorRet, true)
		return false
	}

	return true
}
