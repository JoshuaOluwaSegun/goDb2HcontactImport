package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/hornbill/color"
)

func procFlags() {
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configLogPrefix, "logprefix", "", "Add prefix to the logfile")
	flag.BoolVar(&configMatchLike, "matchlike", false, "Allow the Import to Search Contact by LIKE")
	flag.BoolVar(&configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating Contacts")
	flag.BoolVar(&configVersion, "version", false, "Output Version")
	flag.BoolVar(&configNoColour, "nocol", false, "Set to true to remove colour from CLI output")
	flag.StringVar(&configMaxRoutines, "concurrent", "1", "Maximum number of requests to import concurrently.")

	//-- Parse Flags
	flag.Parse()

	//-- Output config
	if !configVersion {
		outputFlags()
	}

	//Check maxGoroutines for valid value
	maxRoutines, err := strconv.Atoi(configMaxRoutines)
	if err != nil {
		color.Red("Unable to convert maximum concurrency of [" + configMaxRoutines + "] to type INT for processing")
		return
	}
	maxGoroutines = maxRoutines

	if maxGoroutines < 1 || maxGoroutines > 10 {
		color.Red("The maximum concurrent requests allowed is between 1 and 10 (inclusive).\n\n")
		color.Red("You have selected " + configMaxRoutines + ". Please try again, with a valid value against ")
		color.Red("the -concurrent switch.")
		return
	}
}
func outputFlags() {
	//-- Output
	logger(1, "---- XMLMC SQL Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)

	logger(1, "Flag - Config File "+configFileName, true)
	logger(1, "Flag - Log Prefix "+configLogPrefix, true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", configDryRun), true)
}

//-- Check Latest
//-- Function to Load Configruation File
func loadConfig() SQLImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	logger(1, "Loading Config File: "+configurationFilePath, false)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		logger(4, "No Configuration File", true)
		os.Exit(102)
	}

	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true)
	}
	//-- New Decoder
	decoder := json.NewDecoder(file)

	eldapConf := SQLImportConfStruct{}

	//-- Decode JSON
	err := decoder.Decode(&eldapConf)
	//-- Error Checking
	if err != nil {
		logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true)
	}

	//-- Return New Congfig
	return eldapConf
}

func validateConf() error {

	//-- Check for API Key
	if SQLImportConf.APIKey == "" {
		err := errors.New("API Key is not set")
		return err
	}
	//-- Check for Instance ID
	if SQLImportConf.InstanceID == "" {
		err := errors.New("InstanceID is not set")
		return err
	}

	//-- Process Config File

	return nil
}
