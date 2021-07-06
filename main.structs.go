package main

//----- Packages -----
import (
	"sync"
	"time"
)

//----- Constants -----
const (
	version = "1.6.4"
	constOK = "ok"
)

//----- Variables -----
var (
	SQLImportConf      SQLImportConfStruct
	organisations      []organisationListStruct
	counters           counterTypeStruct
	configFileName     string
	configLogPrefix    string
	configDryRun       bool
	configVersion      bool
	configMatchLike    bool
	configMatchOrgLike bool
	configMaxRoutines  string
	configNoColour     bool
	timeNow            string
	startTime          time.Time
	endTime            time.Duration
	errorCount         uint64
	mutexBar           = &sync.Mutex{}
	mutexCounters      = &sync.Mutex{}
	mutexOrganisations = &sync.Mutex{}
	logFileMutex       = &sync.Mutex{}
	worker             sync.WaitGroup
	maxGoroutines      = 6
)

//----- Structs -----
type organisationListStruct struct {
	OrgName   string
	OrgID     int
	CompanyID string
}
type xmlmcOrganisationListResponse struct {
	MethodResult string                       `xml:"status,attr"`
	Params       paramsOrganisationListStruct `xml:"params"`
	State        stateStruct                  `xml:"state"`
}
type paramsOrganisationListStruct struct {
	RowData paramsOrganisationRowDataListStruct `xml:"rowData"`
}
type paramsOrganisationRowDataListStruct struct {
	Row organisationObjectStruct `xml:"row"`
}
type organisationObjectStruct struct {
	OrganizationID   int    `xml:"h_organization_id"`
	OrganizationName string `xml:"h_organization_name"`
}

type counterTypeStruct struct {
	updated uint16
	created uint16
}

//SQLImportConfStruct - Struct that defines the import config schema
type SQLImportConfStruct struct {
	APIKey                      string
	InstanceID                  string
	ContactAction               string
	AttachCustomerPortal        bool
	CustomerPortalOrgView       bool
	CustomerPortalOrgViewRevoke bool
	UpdateContactStatus         bool
	SubscribeToServiceID        int
	SQLConf                     sqlConfStruct
	ContactMapping              map[string]string
	SQLAttributes               []string
}
type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}
type xmlmcExactContactResponse struct {
	MethodResult string      `xml:"status,attr"`
	PKID         string      `xml:"params>primaryEntityData>h_pk_id"`
	State        stateStruct `xml:"state"`
}

type xmlmcCheckContactResponse struct {
	MethodResult string                        `xml:"status,attr"`
	Params       paramsContactSearchListStruct `xml:"params"`
	State        stateStruct                   `xml:"state"`
}
type paramsContactSearchListStruct struct {
	RowData paramsContactRowDataListStruct `xml:"rowData"`
}
type paramsContactRowDataListStruct struct {
	Row contactObjectStruct `xml:"row"`
}
type contactObjectStruct struct {
	PKID string `xml:"h_pk_id"`
}
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsStruct struct {
	SessionID string `xml:"sessionId"`
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
	FieldID   string
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
