package main

import (
	"fmt"
	"strconv"

	//SQL Package
	"github.com/hornbill/sqlx"

	//SQL Drivers
	_ "github.com/alexbrainman/odbc"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/hornbill/go-mssqldb"
)

// queryDatabase -- Query Database for contacts
// -- Builds map of contacts, returns true if successful
func queryDatabase() (bool, []map[string]interface{}) {
	//Clear existing contact Map down
	ArrContactMaps := make([]map[string]interface{}, 0)
	connString := buildConnectionString()
	if connString == "" {
		return false, ArrContactMaps
	}
	//Connect to the JSON specified DB
	db, err := sqlx.Open(SQLImportConf.SQLConf.Driver, connString)
	if err != nil {
		logger(4, " [DATABASE] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrContactMaps
	}
	defer db.Close()

	//Check connection is open
	err = db.Ping()
	if err != nil {
		logger(4, " [DATABASE] [PING] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrContactMaps
	}
	logger(0, "[DATABASE] Connection Successful", true)
	logger(0, "[DATABASE] Running database query for Customers. Please wait...", true)
	//build query
	sqlQuery := SQLImportConf.SQLConf.Query //BaseSQLQuery
	logger(0, "[DATABASE] Query:"+sqlQuery, false)
	//Run Query
	rows, err := db.Queryx(sqlQuery)
	if err != nil {
		logger(4, " [DATABASE] Database Query Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrContactMaps
	}

	//Build map full of contacts
	intContactCount := 0
	for rows.Next() {
		intContactCount++
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			logger(4, " [DATABASE] Row MapScan Error: "+fmt.Sprintf("%v", err), true)
		} else {
			//Stick marshalled data map in to parent slice
			ArrContactMaps = append(ArrContactMaps, results)
		}
	}
	defer rows.Close()
	logger(0, fmt.Sprintf("[DATABASE] Found %d results", intContactCount), false)
	return true, ArrContactMaps
}

// buildConnectionString -- Build the connection string for the SQL driver
func buildConnectionString() string {
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
		if !SQLImportConf.SQLConf.Encrypt {
			connectString = connectString + ";encrypt=disable"
		}
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + ";port=" + dbPortSetting
		}
	case "mysql":
		connectString = SQLImportConf.SQLConf.UserName + ":" + SQLImportConf.SQLConf.Password
		connectString = connectString + "@tcp(" + SQLImportConf.SQLConf.Server + ":"
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + dbPortSetting
		} else {
			connectString = connectString + "3306"
		}
		connectString = connectString + ")/" + SQLImportConf.SQLConf.Database
	case "csv":
		connectString = "Driver={Microsoft Text Driver (*.txt; *.csv)};DefaultDir=C:\\SPF\\Go\\work\\csvtest;Extensions=CSV;Extended Properties=\"text;HDR=Yes;FMT=Delimited\""
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	case "excel":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	}
	return connectString
}
