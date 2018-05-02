//https://www.microsoft.com/en-us/download/details.aspx?id=13255
package main

//----- Packages -----
import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"text/template"

	"crypto/rand"
	"github.com/hornbill/color" //-- CLI Colour
	"github.com/hornbill/goApiLib"
	"github.com/hornbill/pb" //--Hornbil Clone of "github.com/cheggaaa/pb"
	"strconv"
	"strings"
	"sync"
	"time"
	//SQL Package
	"github.com/hornbill/sqlx"
	//SQL Drivers
	_ "github.com/alexbrainman/odbc"
	_ "github.com/hornbill/go-mssqldb"
	_ "github.com/hornbill/mysql"
	_ "github.com/jnewmano/mysql320" //MySQL v3.2.0 to v5 driver - Provides SWSQL (MySQL 4.0.16) support
)

//----- Constants -----
const (
	letterBytes  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	version      = "1.0.1"
	constOK      = "ok"
	updateString = "Update"
	createString = "Create"
)

var (
	SQLImportConf       SQLImportConfStruct
	xmlmcInstanceConfig xmlmcConfig
	xmlmcUsers          []userListItemStruct
	sites               []siteListStruct
	counters            counterTypeStruct
	configFileName      string
	configZone          string
	configLogPrefix     string
	configDryRun        bool
	configVersion       bool
	configWorkers       int
	configMaxRoutines   string
	BaseSQLQuery        string
	timeNow             string
	startTime           time.Time
	endTime             time.Duration
	errorCount          uint64
	noValuesToUpdate    = "There are no values to update"
	mutex               = &sync.Mutex{}
	mutexBar            = &sync.Mutex{}
	mutexCounters       = &sync.Mutex{}
	mutexCustomers      = &sync.Mutex{}
	mutexSite           = &sync.Mutex{}
	mutexSites          = &sync.Mutex{}
	mutexGroups         = &sync.Mutex{}
	mutexManagers       = &sync.Mutex{}
	logFileMutex        = &sync.Mutex{}
	bufferMutex         = &sync.Mutex{}
	worker              sync.WaitGroup
	maxGoroutines       = 6

	ContactArray = []string{
		"logon_id",
		"firstname",
		"lastname",
		"company",
		"email_1",
		"email_2",
		"tel_1",
		"tel_2",
		"jobtitle",
		"description",
		"notes",
		"country",
		"language",
		"private",
		"rights",
		"contact_status",
		"custom_1",
		"custom_2",
		"custom_3",
		"custom_4",
		"custom_5",
		"custom_6"}
)

type siteListStruct struct {
	OrgName   string
	OrgID     int
	CompanyID string
}
type xmlmcSiteListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsSiteListStruct `xml:"params"`
	State        stateStruct          `xml:"state"`
}
type paramsSiteListStruct struct {
	RowData paramsSiteRowDataListStruct `xml:"rowData"`
}
type paramsSiteRowDataListStruct struct {
	Row siteObjectStruct `xml:"row"`
}
type siteObjectStruct struct {
	OrganizationId   int    `xml:"h_organization_id"`
	OrganizationName string `xml:"h_organization_name"`
	/* SiteCountry string `xml:"h_country"` */
}
type xmlmcConfig struct {
	instance string
	zone     string
	url      string
}

type counterTypeStruct struct {
	updated        uint16
	created        uint16
	profileUpdated uint16
	updatedSkipped uint16
	createskipped  uint16
	profileSkipped uint16
}
type contactMappingStruct struct {
	login_id       string
	firstname      string
	lastname       string
	company        string
	email_1        string
	email_2        string
	tel_1          string
	tel_2          string
	jobtitle       string
	description    string
	notes          string
	country        string
	language       string
	private        string
	rights         string
	contact_status string
	custom_1       string
	custom_2       string
	custom_3       string
	custom_4       string
	custom_5       string
	custom_6       string
}

type SQLImportConfStruct struct {
	APIKey               string
	InstanceID           string
	URL                  string
	ContactAction        string
	AttachCustomerPortal bool
	UpdateContactStatus  bool
	SQLConf              sqlConfStruct
	ContactMapping       map[string]string
	SQLAttributes        []string
}
type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}
type xmlmcCheckUserResponse struct {
	MethodResult string                     `xml:"status,attr"`
	Params       paramsUserSearchListStruct `xml:"params"`
	State        stateStruct                `xml:"state"`
}

type paramsUserSearchListStruct struct {
	RowData paramsUserRowDataListStruct `xml:"rowData"`
}
type paramsUserRowDataListStruct struct {
	Row userObjectStruct `xml:"row"`
}
type userObjectStruct struct {
	PKID string `xml:"h_pk_id"`
}

type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsCheckUsersStruct struct {
	RecordExist bool `xml:"recordExist"`
}
type paramsStruct struct {
	SessionID string `xml:"sessionId"`
}
type paramsUserListStruct struct {
	UserListItem []userListItemStruct `xml:"userListItem"`
}
type userListItemStruct struct {
	ContactID string `xml:"contactId"`
	Name      string `xml:"name"`
}
type sqlConfStruct struct {
	Driver    string
	Server    string
	UserName  string
	Password  string
	Port      int
	Query     string
	Database  string
	Encrypt   bool
	ContactID string
}

type xmlmcuserSetGroupOptionsResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
}
type xmlmcprofileSetImageResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}
type xmlmcGroupListResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}

type paramsGroupListStruct struct {
	RowData paramsGroupRowDataListStruct `xml:"rowData"`
}

type paramsGroupRowDataListStruct struct {
	Row groupObjectStruct `xml:"row"`
}

type groupObjectStruct struct {
	GroupID   string `xml:"h_id"`
	GroupName string `xml:"h_name"`
}

type xmlmcPrimEntResponse struct {
	MethodResult string              `xml:"status,attr"`
	Params       paramsPrimEntStruct `xml:"params"`
	State        stateStruct         `xml:"state"`
}

type paramsPrimEntStruct struct {
	RowData paramsPrimEntRowStruct `xml:"primaryEntityData"`
}

type paramsPrimEntRowStruct struct {
	Row primEntObjectStruct `xml:"record"`
}

type primEntObjectStruct struct {
	PkID string `xml:"h_pk_id"`
}

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

//----- Main Function -----
func main() {

	//-- Initiate Variables
	initVars()

	//-- Process Flags
	procFlags()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Load Configuration File Into Struct
	SQLImportConf = loadConfig()

	//-- Validation on Configuration File
	err := validateConf()
	if err != nil {
		logger(4, fmt.Sprintf("%v", err), true)
		logger(4, "Please Check your Configuration File: "+fmt.Sprintf("%s", configFileName), true)
		return
	}

	//-- Set Instance ID
	var boolSetInstance = setInstance(configZone, SQLImportConf.InstanceID)
	if boolSetInstance != true {
		return
	}

	//-- Generate Instance XMLMC Endpoint
	SQLImportConf.URL = getInstanceURL()
	logger(1, "Instance Endpoint "+fmt.Sprintf("%v", SQLImportConf.URL), true)
	//-- Once we have loaded the config write to hornbill log file
	logged := espLogger("---- XMLMC SQL Import Utility V"+fmt.Sprintf("%v", version)+" ----", "debug")

	if !logged {
		logger(4, "Unable to Connect to Instance", true)
		return
	}

	//Set SWSQLDriver to mysql320
	if SQLImportConf.SQLConf.Driver == "swsql" {
		SQLImportConf.SQLConf.Driver = "mysql320"
	}

	//Get asset types, process accordingly
	//	BaseSQLQuery = SQLImportConf.SQLConf.Query
	//   fmt.Println(buildConnectionString())
	var boolSQLUsers, arrUsers = queryDatabase()
	if boolSQLUsers {
		processUsers(arrUsers)
	} else {
		logger(4, "No Results found", true)
		return
	}

	outputEnd()
}

func outputEnd() {
	//-- End output
	if errorCount > 0 {
		logger(4, "Error encountered please check the log file", true)
		logger(4, "Error Count: "+fmt.Sprintf("%d", errorCount), true)
		//logger(4, "Check Log File for Details", true)
	}
	logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
	logger(1, "Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), true)

	logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)
	logger(1, "Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), true)

	logger(1, "Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), true)
	logger(1, "Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), true)

	//-- Show Time Takens
	endTime = time.Now().Sub(startTime)
	logger(1, "Time Taken: "+fmt.Sprintf("%v", endTime), true)
	//-- complete
	complete()
	logger(1, "---- XMLMC SQL Import Complete ---- ", true)
}
func procFlags() {
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configZone, "zone", "eur", "Override the default Zone the instance sits in")
	flag.StringVar(&configLogPrefix, "logprefix", "", "Add prefix to the logfile")
	flag.BoolVar(&configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating users")
	flag.BoolVar(&configVersion, "version", false, "Output Version")
	flag.IntVar(&configWorkers, "workers", 1, "Number of Worker threads to use")
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

	logger(1, "Flag - Config File "+fmt.Sprintf("%s", configFileName), true)
	logger(1, "Flag - Zone "+fmt.Sprintf("%s", configZone), true)
	logger(1, "Flag - Log Prefix "+fmt.Sprintf("%s", configLogPrefix), true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", configDryRun), true)
	logger(1, "Flag - Workers "+fmt.Sprintf("%v", configWorkers), false)
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

//-- Worker Pool Function
func loggerGen(t int, s string) string {

	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
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
func loggerWriteBuffer(s string) {
	logger(0, s, false)
}

//-- Logging function
func logger(t int, s string, outputtoCLI bool) {
	//-- Curreny WD
	cwd, _ := os.Getwd()
	//-- Log Folder
	logPath := cwd + "/log"
	//-- Log File
	logFileName := logPath + "/" + configLogPrefix + "SQL_User_Import_" + timeNow + ".log"
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

//-- complete
func complete() {
	//-- End output
	espLogger("Errors: "+fmt.Sprintf("%d", errorCount), "error")
	espLogger("Updated: "+fmt.Sprintf("%d", counters.updated), "debug")
	espLogger("Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), "debug")
	espLogger("Created: "+fmt.Sprintf("%d", counters.created), "debug")
	espLogger("Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), "debug")
	espLogger("Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), "debug")
	espLogger("Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), "debug")
	espLogger("Time Taken: "+fmt.Sprintf("%v", endTime), "debug")
	espLogger("---- XMLMC SQL User Import Complete ---- ", "debug")
}

// Set Instance Id
func setInstance(strZone string, instanceID string) bool {
	//-- Set Zone
	setZone(strZone)
	//-- Check for blank instance
	if instanceID == "" {
		logger(4, "InstanceId Must be Specified in the Configuration File", true)
		return false
	}
	//-- Set Instance
	xmlmcInstanceConfig.instance = instanceID
	return true
}

// Set Instance Zone to Overide Live
func setZone(zone string) {
	xmlmcInstanceConfig.zone = zone

	return
}

//-- Log to ESP
func espLogger(message string, severity string) bool {
	espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.URL)
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

//-- Function Builds XMLMC End Point
func getInstanceURL() string {
	xmlmcInstanceConfig.url = "https://"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.zone
	xmlmcInstanceConfig.url += "api.hornbill.com/"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.instance
	xmlmcInstanceConfig.url += "/xmlmc/"

	return xmlmcInstanceConfig.url
}

//buildConnectionString -- Build the connection string for the SQL driver
func buildConnectionString() string {
	//	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" || SQLImportConf.SQLConf.Password == "" {
	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" {
		//Conf not set - log error and return empty string
		logger(4, "Database configuration not set.", true)
		return ""
	}
	logger(1, "Connecting to Database Server: "+SQLImportConf.SQLConf.Server, true)
	connectString := ""
	switch SQLImportConf.SQLConf.Driver {
	case "mssql":
		connectString = "server=" + SQLImportConf.SQLConf.Server
		connectString = connectString + ";database=" + SQLImportConf.SQLConf.Database
		connectString = connectString + ";user id=" + SQLImportConf.SQLConf.UserName
		connectString = connectString + ";password=" + SQLImportConf.SQLConf.Password
		if SQLImportConf.SQLConf.Encrypt == false {
			connectString = connectString + ";encrypt=disable"
		}
		if SQLImportConf.SQLConf.Port != 0 {
			var dbPortSetting string
			dbPortSetting = strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + ";port=" + dbPortSetting
		}
	case "mysql":
		connectString = SQLImportConf.SQLConf.UserName + ":" + SQLImportConf.SQLConf.Password
		connectString = connectString + "@tcp(" + SQLImportConf.SQLConf.Server + ":"
		if SQLImportConf.SQLConf.Port != 0 {
			var dbPortSetting string
			dbPortSetting = strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + dbPortSetting
		} else {
			connectString = connectString + "3306"
		}
		connectString = connectString + ")/" + SQLImportConf.SQLConf.Database
	case "mysql320":
		var dbPortSetting string
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting = strconv.Itoa(SQLImportConf.SQLConf.Port)
		} else {
			dbPortSetting = "3306"
		}
		connectString = "tcp:" + SQLImportConf.SQLConf.Server + ":" + dbPortSetting
		connectString = connectString + "*" + SQLImportConf.SQLConf.Database + "/" + SQLImportConf.SQLConf.UserName + "/" + SQLImportConf.SQLConf.Password
	case "csv":
		//connectString = "driver=Microsoft.Jet.OLEDB.4.0; Data Source=C:\\SPF\\Go\\work\\csvtest;Extended Properties=\"text;HDR=Yes;FMT=Delimited\""
		connectString = "Driver={Microsoft Text Driver (*.txt; *.csv)};DefaultDir=C:\\SPF\\Go\\work\\csvtest;Extensions=CSV;Extended Properties=\"text;HDR=Yes;FMT=Delimited\""
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	case "excel":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	}
	return connectString
}

//queryDatabase -- Query Asset Database for assets of current type
//-- Builds map of assets, returns true if successful
func queryDatabase() (bool, []map[string]interface{}) {
	//Clear existing Asset Map down
	ArrUserMaps := make([]map[string]interface{}, 0)
	connString := buildConnectionString()
	if connString == "" {
		return false, ArrUserMaps
	}
	//Connect to the JSON specified DB
	db, err := sqlx.Open(SQLImportConf.SQLConf.Driver, connString)
	defer db.Close()
	if err != nil {
		logger(4, " [DATABASE] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}
	//Check connection is open
	err = db.Ping()
	if err != nil {
		logger(4, " [DATABASE] [PING] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}
	logger(3, "[DATABASE] Connection Successful", true)
	logger(3, "[DATABASE] Running database query for Customers. Please wait...", true)
	//build query
	sqlQuery := SQLImportConf.SQLConf.Query //BaseSQLQuery
	logger(3, "[DATABASE] Query:"+sqlQuery, false)
	//Run Query
	rows, err := db.Queryx(sqlQuery)
	if err != nil {
		logger(4, " [DATABASE] Database Query Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}

	//Build map full of assets
	intUserCount := 0
	for rows.Next() {
		intUserCount++
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		//Stick marshalled data map in to parent slice
		ArrUserMaps = append(ArrUserMaps, results)
	}
	defer rows.Close()
	logger(3, fmt.Sprintf("[DATABASE] Found %d results", intUserCount), false)
	return true, ArrUserMaps
}

//processAssets -- Processes Assets from Asset Map
//--If asset already exists on the instance, update
//--If asset doesn't exist, create
func processUsers(arrUsers []map[string]interface{}) {
	bar := pb.StartNew(len(arrUsers))
	logger(1, "Processing Contacts", false)

	//Get the identity of the AssetID field from the config
	contactIDField := fmt.Sprintf("%v", SQLImportConf.SQLConf.ContactID)
	//-- Loop each asset
	maxGoroutinesGuard := make(chan struct{}, maxGoroutines)

	for _, customerRecord := range arrUsers {
		maxGoroutinesGuard <- struct{}{}
		worker.Add(1)
		userMap := customerRecord
		//Get the asset ID for the current record
		contactID := fmt.Sprintf("%s", userMap[contactIDField])
		logger(1, "User ID: "+contactID, false)
		if contactID != "" {
			//logger(1, "User ID: "+fmt.Sprintf("%s", contactID), false)
			espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.URL)
			espXmlmc.SetAPIKey(SQLImportConf.APIKey)
			go func() {
				defer worker.Done()
				time.Sleep(1 * time.Millisecond)
				mutexBar.Lock()
				bar.Increment()
				mutexBar.Unlock()

				var isErr = false
				foundId, err := checkUserOnInstance(contactID, espXmlmc)
				if err != nil {
					logger(4, "Unable to Search For User: "+fmt.Sprintf("%v", err), false)
					isErr = true
				}
				//logger(4, "Found Contact: "+fmt.Sprintf("%d", foundId), true)
				//-- Update or Create Asset
				if !isErr {
					if foundId > 0 && SQLImportConf.ContactAction != createString {
						logger(1, fmt.Sprintf("Update Customer: %s (%d)", contactID, foundId), false)
						_, errUpdate := updateUser(userMap, espXmlmc, foundId)
						if errUpdate != nil {
							logger(4, "Unable to Update User: "+fmt.Sprintf("%v", errUpdate), false)
						}
					} else if foundId < 0 && SQLImportConf.ContactAction != updateString {
						logger(1, "Create Customer: "+contactID, false)
						_, errorCreate := updateUser(userMap, espXmlmc, foundId)
						if errorCreate != nil {
							logger(4, "Unable to Create User: "+fmt.Sprintf("%v", errorCreate), false)
						}
					}
				}
				<-maxGoroutinesGuard
			}()
		}
	}
	worker.Wait()
	bar.FinishPrint("Processing Complete!")
}

func updateUser(u map[string]interface{}, espXmlmc *apiLib.XmlmcInstStruct, foundId int) (bool, error) {
	buf2 := bytes.NewBufferString("")
	searchResultSiteID := ""
	//searchResultCompID := ""
	//-- Do we Lookup Site
	var p map[string]string
	p = make(map[string]string)
	//fmt.Println("%v", u)
	for key, value := range u {
		p[key] = fmt.Sprintf("%s", value)
	}
	//    contactID := p[SQLImportConf.SQLConf.ContactID]
	espXmlmc.SetParam("entity", "Contact")
	espXmlmc.SetParam("returnModifiedData", "true")
	espXmlmc.SetParam("returnRawValues", "true")

	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")

	if foundId > 0 {
		espXmlmc.SetParam("h_pk_id", fmt.Sprintf("%d", foundId))
	}

	//-- Loop Through UserProfileMapping
	for key := range ContactArray {
		field := ContactArray[key]
		value := SQLImportConf.ContactMapping[field]
		//  fmt.Println(abc)
		t := template.New(field)
		t, _ = t.Parse(value)
		buf := bytes.NewBufferString("")
		t.Execute(buf, p)
		value = buf.String()
		if value == "%!s(<nil>)" {
			value = ""
		}

		if field == "contact_status" && foundId > 0 && !SQLImportConf.UpdateContactStatus {
		} else if field == "company" && value != "" {
			espXmlmc.SetParam("h_"+field, value)
			//search IF result:
			//searchResultSiteID, searchResultCompID = getSiteFromLookup(value, buf2)
			searchResultSiteID, _ = getSiteFromLookup(value, buf2)
			if searchResultSiteID != "" {
				espXmlmc.SetParam("h_organization_id", searchResultSiteID)
			}
		} else if value != "" {
			espXmlmc.SetParam("h_"+field, value)
		}
	}

	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	var XMLSTRING = espXmlmc.GetParam()
	logger(1, "User Create/Update XML "+fmt.Sprintf("%s", XMLSTRING), false)
	//-- Check for Dry Run
	if configDryRun != true {
		var XMLCreate string
		var xmlmcErr error

		if foundId > 0 {
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityUpdateRecord")
		} else {
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityAddRecord")
		}

		var xmlRespon xmlmcPrimEntResponse //xmlmcResponse
		if xmlmcErr != nil {
			errorCountInc()
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			errorCountInc()
			return false, err
		}
		if xmlRespon.MethodResult != constOK {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			return false, err

		}
		if foundId < 0 {
			foundId, err = strconv.Atoi(xmlRespon.Params.RowData.Row.PkID)
		}
		//fmt.Println(xmlRespon.Params.RowData.Row.PkID + " --- " + strconv.Itoa(foundId) + " ::: " + searchResultSiteID)
		//fmt.Sprintln("%v",xmlRespon.Params.RowData.Row)
		//        if (searchResultCompID != "" && foundId > 0){
		if searchResultSiteID != "" && foundId > 0 {
			logger(1, "Org Relation", false)
			var xmlRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("entity", "RelatedContainer")
			//espXmlmc.SetParam("returnModifiedData", false)
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_container", searchResultSiteID)
			espXmlmc.SetParam("h_element", strconv.Itoa(foundId))
			espXmlmc.SetParam("h_element_type", "Contact")
			espXmlmc.SetParam("h_rel_type", "member")
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			//            logger(1, "Org Rel XML1 "+fmt.Sprintf("%s", espXmlmc.GetParam()), false)
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityAddRecord")
			if xmlmcErr != nil {
				errorCountInc()
				return false, xmlmcErr
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlRelationResp)
			if err != nil {
				errorCountInc()
				return false, err
			}
			//logger(1, xmlRelationResp.State.ErrorRet, false)
			/* Fire and forget
			   if xmlRelationResp.MethodResult != constOK {
			       err = errors.New(xmlRelationResp.State.ErrorRet)
			       errorCountInc()
			       return false, err
			   }
			*/

			//### Addition for contact = organisation
			var xmlCompRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("entity", "OrganizationContacts")
			espXmlmc.SetParam("returnModifiedData", "false")
			espXmlmc.SetParam("formatValues", "false")
			espXmlmc.SetParam("returnRawValues", "false")
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_organization_id", searchResultSiteID)
			espXmlmc.SetParam("h_contact_id", strconv.Itoa(foundId))
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			//            logger(1, "Org Rel XML2 "+fmt.Sprintf("%s", espXmlmc.GetParam()), false)
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityAddRecord")
			if xmlmcErr != nil {
				errorCountInc()
				return false, xmlmcErr
			}
			err = xml.Unmarshal([]byte(XMLCreate), &xmlCompRelationResp)
			if err != nil {
				errorCountInc()
				return false, err
			}
			//logger(1, xmlCompRelationResp.State.ErrorRet, false)
			// Fire and Forget
		}

		if foundId > 0 {
			logger(1, "User Update Success", false)
			updateCountInc()
		} else {
			logger(1, "User Create Success", false)
			createCountInc()
		}

		if SQLImportConf.AttachCustomerPortal {
			var xmlRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("portalId", "customer")
			espXmlmc.SetParam("contactId", strconv.Itoa(foundId))
			espXmlmc.SetParam("accessStatus", "approved")
			XMLCreate, xmlmcErr = espXmlmc.Invoke("admin", "portalSetContactAccess")
			if xmlmcErr != nil {
				errorCountInc()
				return false, xmlmcErr
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlRelationResp)
			if err != nil {
				errorCountInc()
				return false, err
			}
			if xmlRespon.MethodResult != constOK {
				err = errors.New(xmlRelationResp.State.ErrorRet)
				errorCountInc()
				return false, err
			}
		}

		logger(1, buf2.String(), false)
		return true, nil
	} else {
	}
	//-- DEBUG XML TO LOG FILE
	logger(1, "User Create XML "+fmt.Sprintf("%s", espXmlmc.GetParam()), false)
	if foundId > 0 {
		updateSkippedCountInc()
	} else {
		createSkippedCountInc()
	}
	espXmlmc.ClearParam()

	return true, nil
}

func checkUserOnInstance(contactID string, espXmlmc *apiLib.XmlmcInstStruct) (int, error) {
	/*
		espXmlmc.SetParam("entity", "Contact")
		espXmlmc.SetParam("keyValue", contactID)
		XMLCheckUser, xmlmcErr := espXmlmc.Invoke("data", "entityDoesRecordExist")
		var xmlRespon xmlmcCheckUserResponse
		if xmlmcErr != nil {
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLCheckUser), &xmlRespon)
		if err != nil {
			stringError := err.Error()
			stringBody := string(XMLCheckUser)
			errWithBody := errors.New(stringError + " RESPONSE BODY: " + stringBody)
			return false, errWithBody
		}
		if xmlRespon.MethodResult != constOK {
			err := errors.New(xmlRespon.State.ErrorRet)
			return false, err
		}
		return xmlRespon.Params.RecordExist, nil
	*/
	var intReturn int
	intReturn = -1

	espXmlmc.SetParam("entity", "Contact")
	espXmlmc.SetParam("matchScope", "all")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_logon_id", contactID)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")

	XMLCheckUser, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")
	var xmlRespon xmlmcCheckUserResponse
	if xmlmcErr != nil {
		//	buffer.WriteString(loggerGen(4, "Unable to Search for Contact: "+fmt.Sprintf("%v", xmlmcErr)))
		logger(4, "Unable to Search for Contact: "+fmt.Sprintf("%v", xmlmcErr), true)
	}
	err := xml.Unmarshal([]byte(XMLCheckUser), &xmlRespon)
	if err != nil {
		//buffer.WriteString(loggerGen(4, "Unable to Search for Contact: "+fmt.Sprintf("%v", err)))
		logger(4, "Unable to Search for Contact: "+fmt.Sprintf("%v", err), true)
	} else {
		if xmlRespon.MethodResult != constOK {
			//buffer.WriteString(loggerGen(4, "Unable to Search for Contact: "+xmlRespon.State.ErrorRet))
			logger(4, "Unable to Search for Contact: "+xmlRespon.State.ErrorRet, true)
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.PKID != "" {
				intReturn, err = strconv.Atoi(xmlRespon.Params.RowData.Row.PKID)
			}
		}
	}

	return intReturn, err

}

//-- Function to search for site
func getSiteFromLookup(site string, buffer *bytes.Buffer) (string, string) {
	siteReturn := ""
	compReturn := ""
	/*	//-- Check if Site Attribute is set
		if SQLImportConf.SiteLookup.Attribute == "" {
			buffer.WriteString(loggerGen(4, "Site Lookup is Enabled but Attribute is not Defined"))
			return ""
		}
		//-- Get Value of Attribute
		buffer.WriteString(loggerGen(1, "SQL Attribute for Site Lookup: "+SQLImportConf.SiteLookup.Attribute))
	*/
	//-- Get Value of Attribute
	siteAttributeName := processComplexField(site)
	buffer.WriteString(loggerGen(1, "Looking Up Site: "+siteAttributeName))
	if siteAttributeName == "" {
		return "", ""
	}
	siteIsInCache, SiteIDCache, CompID := orgInCache(siteAttributeName)
	//-- Check if we have Cached the site already
	if siteIsInCache {
		//fmt.Println("IN CACHE")
		siteReturn = strconv.Itoa(SiteIDCache)
		compReturn = CompID
		buffer.WriteString(loggerGen(1, "Found Site in Cache: "+siteReturn+" ("+CompID+")"))
	} else {
		//fmt.Println("NO CACHE")
		siteIsOnInstance, SiteIDInstance, compIDInstance := searchOrg(siteAttributeName, buffer)
		//-- If Returned set output
		if siteIsOnInstance {
			siteReturn = strconv.Itoa(SiteIDInstance)
			compReturn = compIDInstance
		}
	}
	buffer.WriteString(loggerGen(1, "Site Lookup found ID: "+siteReturn))
	return siteReturn, compReturn
}

func processComplexField(s string) string {
	return html.UnescapeString(s)
}

//-- Function to Check if in Cache
func orgInCache(orgName string) (bool, int, string) {
	boolReturn := false
	intReturn := 0
	stringCompReturn := ""
	mutexSites.Lock()
	//-- Check if in Cache
	for _, site := range sites {
		if site.OrgName == orgName {
			boolReturn = true
			intReturn = site.OrgID
			stringCompReturn = site.CompanyID
			break
		}
	}
	mutexSites.Unlock()
	return boolReturn, intReturn, stringCompReturn
}

//-- Function to Check if site is on the instance
func searchOrg(orgName string, buffer *bytes.Buffer) (bool, int, string) {
	boolReturn := false
	intReturn := 0
	strCompReturn := ""
	//-- ESP Query for site
	espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.URL)
	espXmlmc.SetAPIKey(SQLImportConf.APIKey)
	if orgName == "" {
		return boolReturn, intReturn, ""
	}
	espXmlmc.SetParam("entity", "Organizations")
	espXmlmc.SetParam("matchScope", "all")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_organization_name", orgName)
	//	espXmlmc.SetParam("column", "h_organization_name")
	//	espXmlmc.SetParam("value", orgName)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	//fmt.Println(espXmlmc.GetParam())
	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")
	//fmt.Println(XMLSiteSearch)
	var xmlRespon xmlmcSiteListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.OrganizationName != "" {
				if strings.ToLower(xmlRespon.Params.RowData.Row.OrganizationName) == strings.ToLower(orgName) {
					intReturn = xmlRespon.Params.RowData.Row.OrganizationId
					boolReturn = true

					var xml2Resp xmlmcGroupListResponse
					espXmlmc.ClearParam()
					espXmlmc.SetParam("entity", "Container")
					espXmlmc.SetParam("matchScope", "all")
					//espXmlmc.SetParam("returnModifiedData", false)
					espXmlmc.OpenElement("searchFilter")
					espXmlmc.SetParam("h_name", orgName)
					espXmlmc.SetParam("h_type", "Organizations")
					espXmlmc.CloseElement("searchFilter")
					XMLOrgSearch, xmlmcOSErr := espXmlmc.Invoke("data", "entityBrowseRecords")
					if xmlmcOSErr != nil {
						buffer.WriteString(loggerGen(4, "Unable to Search for Container: "+fmt.Sprintf("%v", xmlmcOSErr)))
						errorCountInc()
					} else {
						err := xml.Unmarshal([]byte(XMLOrgSearch), &xml2Resp)
						if err != nil {
							buffer.WriteString(loggerGen(4, "Unable to read container response: "+fmt.Sprintf("%v", err)))
							buffer.WriteString(loggerGen(4, fmt.Sprintf("%v", XMLOrgSearch)))
							errorCountInc()
						} else if xml2Resp.MethodResult != constOK {
							buffer.WriteString(loggerGen(4, fmt.Sprintf("%v", XMLOrgSearch)))
							buffer.WriteString(loggerGen(4, "Unable to deal with container response: "+xmlRespon.State.ErrorRet))
							err = errors.New(xml2Resp.State.ErrorRet)
							errorCountInc()
						} else {
							strCompReturn = xml2Resp.Params.RowData.Row.GroupID
						}

					}
					//-- Add Site to Cache
					mutexSites.Lock()
					var newSiteForCache siteListStruct
					newSiteForCache.OrgID = intReturn
					newSiteForCache.OrgName = orgName
					newSiteForCache.CompanyID = strCompReturn
					name := []siteListStruct{newSiteForCache}
					sites = append(sites, name...)
					mutexSites.Unlock()

				}
			}
		}
	}

	return boolReturn, intReturn, strCompReturn
}

//-- Generate Password String
func generatePasswordString(n int) string {
	var arbytes = make([]byte, n)
	rand.Read(arbytes)
	for i, b := range arbytes {
		arbytes[i] = letterBytes[b%byte(len(letterBytes))]
	}
	return string(arbytes)
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
func updateSkippedCountInc() {
	mutexCounters.Lock()
	counters.updatedSkipped++
	mutexCounters.Unlock()
}
func createSkippedCountInc() {
	mutexCounters.Lock()
	counters.createskipped++
	mutexCounters.Unlock()
}
func createCountInc() {
	mutexCounters.Lock()
	counters.created++
	mutexCounters.Unlock()
}
func profileCountInc() {
	mutexCounters.Lock()
	counters.profileUpdated++
	mutexCounters.Unlock()
}
func profileSkippedCountInc() {
	mutexCounters.Lock()
	counters.profileSkipped++
	mutexCounters.Unlock()
}
