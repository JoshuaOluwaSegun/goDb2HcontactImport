//https://www.microsoft.com/en-us/download/details.aspx?id=13255
package main

//----- Packages -----
import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"text/template"

	"strconv"
	"strings"
	"time"

	apiLib "github.com/hornbill/goApiLib"
	"github.com/hornbill/pb"
)

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
		logger(4, "Please Check your Configuration File: "+configFileName, true)
		return
	}

	logger(1, "Instance ID: "+SQLImportConf.InstanceID, true)
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

	boolSQLContacts, arrContacts := queryDatabase()
	if boolSQLContacts {
		processContacts(arrContacts)
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
	}
	logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
	logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)

	//-- Show Time Takens
	endTime = time.Since(startTime)
	logger(1, "Time Taken: "+fmt.Sprintf("%v", endTime), true)

	//-- End output to log
	espLogger("Errors: "+fmt.Sprintf("%d", errorCount), "error")
	espLogger("Updated: "+fmt.Sprintf("%d", counters.updated), "debug")
	espLogger("Created: "+fmt.Sprintf("%d", counters.created), "debug")
	espLogger("Time Taken: "+fmt.Sprintf("%v", endTime), "debug")
	espLogger("---- XMLMC SQL Contact Import Complete ---- ", "debug")
	logger(1, "---- XMLMC SQL Contact Import Complete ---- ", true)
}

//processContacts -- Processes contacts from contact map
//--If contact already exists on the instance, update
//--If contact doesn't exist, create
func processContacts(arrContacts []map[string]interface{}) {
	bar := pb.StartNew(len(arrContacts))
	logger(1, "Processing Contacts...\n", false)

	//Get the identity of the contactID field from the config
	contactIDField := fmt.Sprintf("%v", SQLImportConf.SQLConf.ContactID)
	//-- Loop each contact
	maxGoroutinesGuard := make(chan struct{}, maxGoroutines)

	for _, contactRecord := range arrContacts {
		maxGoroutinesGuard <- struct{}{}
		worker.Add(1)
		contactMap := contactRecord
		//Get the contact ID for the current record
		contactID := fmt.Sprintf("%s", contactMap[contactIDField])
		if contactID != "" {
			espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.InstanceID)
			espXmlmc.SetAPIKey(SQLImportConf.APIKey)
			go func() {
				defer worker.Done()
				mutexBar.Lock()
				bar.Increment()
				mutexBar.Unlock()

				var buffer bytes.Buffer
				buffer.WriteString("[DEBUG] Processing Contact [" + contactID + "]\n")
				foundID, err := checkContactOnInstance(contactID, espXmlmc, &buffer)
				if err == nil {
					if foundID > 0 && (SQLImportConf.ContactAction == "Update" || SQLImportConf.ContactAction == "Both") {
						buffer.WriteString(loggerGen(1, "Update Contact: ["+contactID+"] ["+strconv.Itoa(foundID)+"]"))
						upsertContact(contactMap, espXmlmc, foundID, &buffer)
					} else if foundID <= 0 && (SQLImportConf.ContactAction == "Create" || SQLImportConf.ContactAction == "Both") {
						buffer.WriteString(loggerGen(1, "Create Contact: ["+contactID+"]"))
						upsertContact(contactMap, espXmlmc, foundID, &buffer)
					}
				}
				loggerWriteBuffer(buffer.String())
				buffer.Reset()
				<-maxGoroutinesGuard
			}()
		}
	}
	worker.Wait()
	bar.FinishPrint("Processing Complete!")
}

func upsertContact(u map[string]interface{}, espXmlmc *apiLib.XmlmcInstStruct, foundID int, buffer *bytes.Buffer) {
	insertContact := (foundID <= 0)
	searchResultOrganisationID := ""
	searchResultCompID := ""
	p := make(map[string]string)
	for key, value := range u {
		p[key] = fmt.Sprintf("%s", value)
	}
	espXmlmc.SetParam("entity", "Contact")
	espXmlmc.SetParam("returnModifiedData", "true")
	espXmlmc.SetParam("returnRawValues", "true")

	espXmlmc.OpenElement("primaryEntityData")
	espXmlmc.OpenElement("record")

	if !insertContact {
		espXmlmc.SetParam("h_pk_id", fmt.Sprintf("%d", foundID))
	}

	//-- Loop Through ContactArray
	for key := range ContactArray {
		field := ContactArray[key]
		value := SQLImportConf.ContactMapping[field]
		t := template.New(field)
		t, _ = t.Parse(value)
		buf := bytes.NewBufferString("")
		t.Execute(buf, p)
		value = buf.String()
		if value == "%!s(<nil>)" {
			value = ""
		}

		if field == "contact_status" && !insertContact && !SQLImportConf.UpdateContactStatus {
		} else if field == "company" && value != "" {
			espXmlmc.SetParam("h_"+field, value)
			searchResultOrganisationID, searchResultCompID = getOrgFromLookup(value, buffer)
			if searchResultOrganisationID != "" {
				espXmlmc.SetParam("h_organization_id", searchResultOrganisationID)
			}
		} else if value != "" {
			espXmlmc.SetParam("h_"+field, value)
		}
	}

	espXmlmc.CloseElement("record")
	espXmlmc.CloseElement("primaryEntityData")
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "Contact Create/Update XML: "+XMLSTRING))
	//-- Check for Dry Run
	if !configDryRun {
		var XMLCreate string
		var xmlmcErr error
		method := "entityAddRecord"
		if !insertContact {
			method = "entityUpdateRecord"
		}
		XMLCreate, xmlmcErr = espXmlmc.Invoke("data", method)
		var xmlRespon xmlmcPrimEntResponse
		if xmlmcErr != nil {
			errorCountInc()
			buffer.WriteString(loggerGen(4, "Contact Create Failed. API Invoke Error from ["+method+"] : "+fmt.Sprintf("%v", xmlmcErr)))
			return
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			errorCountInc()
			buffer.WriteString(loggerGen(4, "Contact Create Failed. Unmarshall Error from ["+method+"] : "+fmt.Sprintf("%v", err)))
			return
		}
		if xmlRespon.MethodResult != constOK {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			buffer.WriteString(loggerGen(4, "Contact Create Failed. MethodResult Not OK from ["+method+"] : "+fmt.Sprintf("%v", err)))
			return

		}

		if insertContact {
			buffer.WriteString(loggerGen(1, "Contact Create Success"))
			createCountInc()
		} else {
			buffer.WriteString(loggerGen(1, "Contact Update Success"))
			updateCountInc()
		}

		if foundID <= 0 {
			foundID, _ = strconv.Atoi(xmlRespon.Params.RowData.Row.PkID)
		}
		buffer.WriteString(loggerGen(1, "Contact ID: "+strconv.Itoa(foundID)))
		if searchResultOrganisationID != "" && foundID > 0 {
			var xmlRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("entity", "RelatedContainer")
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_container", searchResultCompID)
			espXmlmc.SetParam("h_element", strconv.Itoa(foundID))
			espXmlmc.SetParam("h_element_type", "Contact")
			espXmlmc.SetParam("h_rel_type", "member")
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityAddRecord")
			if xmlmcErr != nil {
				errorCountInc()
				buffer.WriteString(loggerGen(3, "Adding Contact to Container Unsuccessful. API Invoke Error from [entityAddRecord] for entity [RelatedContainer]: "+fmt.Sprintf("%v", xmlmcErr)))
				return
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlRelationResp)
			if err != nil {
				errorCountInc()
				buffer.WriteString(loggerGen(3, "Adding Contact to Container Unsuccessful. Unmarshall Error from [entityAddRecord] for entity [RelatedContainer]: "+fmt.Sprintf("%v", err)))
				return
			}
			buffer.WriteString(loggerGen(1, "Adding Contact to Container Success"))

			//### Addition for contact = organisation
			var xmlCompRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("entity", "OrganizationContacts")
			espXmlmc.SetParam("returnModifiedData", "false")
			espXmlmc.SetParam("formatValues", "false")
			espXmlmc.SetParam("returnRawValues", "false")
			espXmlmc.OpenElement("primaryEntityData")
			espXmlmc.OpenElement("record")
			espXmlmc.SetParam("h_organization_id", searchResultOrganisationID)
			espXmlmc.SetParam("h_contact_id", strconv.Itoa(foundID))
			espXmlmc.CloseElement("record")
			espXmlmc.CloseElement("primaryEntityData")
			XMLCreate, xmlmcErr = espXmlmc.Invoke("data", "entityAddRecord")
			if xmlmcErr != nil {
				errorCountInc()
				buffer.WriteString(loggerGen(3, "Adding Contact to Organisation Unsuccessful. API Invoke Error from [entityAddRecord] for entity [OrganizationContacts]: "+fmt.Sprintf("%v", xmlmcErr)))
				return
			}
			err = xml.Unmarshal([]byte(XMLCreate), &xmlCompRelationResp)
			if err != nil {
				errorCountInc()
				buffer.WriteString(loggerGen(3, "Adding Contact to Organisation Unsuccessful. Unmarshall Error from [entityAddRecord] for entity [OrganizationContacts]: "+fmt.Sprintf("%v", err)))
				return
			}
			buffer.WriteString(loggerGen(1, "Adding Contact to Organization Success"))
		}

		if SQLImportConf.AttachCustomerPortal {
			var xmlRelationResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("portalId", "customer")
			espXmlmc.SetParam("contactId", strconv.Itoa(foundID))
			espXmlmc.SetParam("accessStatus", "approved")
			XMLCreate, xmlmcErr = espXmlmc.Invoke("admin", "portalSetContactAccess")
			if xmlmcErr != nil {
				buffer.WriteString(loggerGen(3, "Attaching Contact to Portal unsuccessful. API Invoke Error from [portalSetContactAccess]: "+fmt.Sprintf("%v", xmlmcErr)))
				errorCountInc()
				return
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlRelationResp)
			if err != nil {
				buffer.WriteString(loggerGen(3, "Attaching Contact to Portal unsuccessful. Unmarshall Error from [portalSetContactAccess]: "+fmt.Sprintf("%v", err)))
				errorCountInc()
				return
			}
			if xmlRelationResp.MethodResult != constOK {
				err = errors.New(xmlRelationResp.State.ErrorRet)
				buffer.WriteString(loggerGen(3, "Attaching Contact to Portal unsuccessful. MethodResult not OK from [portalSetContactAccess]: "+fmt.Sprintf("%v", err)))
				errorCountInc()
				return
			}
			buffer.WriteString(loggerGen(1, "Contact to Portal Attach Success"))
		}
		//Service Subscription
		if SQLImportConf.SubscribeToServiceID > 0 {
			var xmlSubScribeResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("serviceId", strconv.Itoa(SQLImportConf.SubscribeToServiceID))
			espXmlmc.SetParam("subscriberId", strconv.Itoa(foundID))
			espXmlmc.SetParam("subscriberType", "Contact")
			XMLCreate, xmlmcErr = espXmlmc.Invoke("apps/com.hornbill.servicemanager/ServiceSubscriptions", "add")
			if xmlmcErr != nil {
				buffer.WriteString(loggerGen(3, "Subscribing Contact Unsuccessful. API Invoke Error from [ServiceSubscriptions_add]: "+fmt.Sprintf("%v", xmlmcErr)))
				errorCountInc()
				return
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlSubScribeResp)
			if err != nil {
				buffer.WriteString(loggerGen(3, "Subscribing Contact Unsuccessful. Unmarshall Error from [ServiceSubscriptions_add]: "+fmt.Sprintf("%v", err)))
				return
			}
			if xmlSubScribeResp.MethodResult != constOK {
				err = errors.New(xmlSubScribeResp.State.ErrorRet)
				buffer.WriteString(loggerGen(3, "Subscribing Contact  Unsuccessful. MethodResult not OK from [ServiceSubscriptions_add]: "+fmt.Sprintf("%v", err)))
				return
			}
			buffer.WriteString(loggerGen(1, "Subscribing Contact  Success"))
		}
		//######
		if SQLImportConf.CustomerPortalOrgView {
			var xmlPortalOrgViewResp xmlmcResponse
			espXmlmc.ClearParam()
			espXmlmc.SetParam("userId", strconv.Itoa(foundID))
			if SQLImportConf.CustomerPortalOrgViewRevoke {
				espXmlmc.SetParam("level", "0")
			} else {
				espXmlmc.SetParam("level", "1")
			}
			XMLCreate, xmlmcErr = espXmlmc.Invoke("apps/com.hornbill.servicemanager/ContactOrgRequests", "changeOrgRequestSetting")
			if xmlmcErr != nil {
				buffer.WriteString(loggerGen(3, "Allowing Org View Unsuccessful. API Invoke Error from [changeOrgRequestSetting]: "+fmt.Sprintf("%v", xmlmcErr)))
				errorCountInc()
				return
			}
			err := xml.Unmarshal([]byte(XMLCreate), &xmlPortalOrgViewResp)
			if err != nil {
				buffer.WriteString(loggerGen(3, "Allowing Org View Unsuccessful. Unmarshall Error from [changeOrgRequestSetting]: "+fmt.Sprintf("%v", err)))
				return
			}
			if xmlPortalOrgViewResp.MethodResult != constOK {
				err = errors.New(xmlPortalOrgViewResp.State.ErrorRet)
				buffer.WriteString(loggerGen(3, "Allowing Org View Unsuccessful. MethodResult not OK from [changeOrgRequestSetting]: "+fmt.Sprintf("%v", err)))
				return
			}
			buffer.WriteString(loggerGen(1, "Allowing Org View Success"))
		}
	}
	espXmlmc.ClearParam()
}

func checkContactOnInstance(contactID string, espXmlmc *apiLib.XmlmcInstStruct, buffer *bytes.Buffer) (int, error) {
	intReturn := -1
	var err error
	espXmlmc.SetParam("entity", "Contact")

//	if configMatchLike {
		espXmlmc.SetParam("matchScope", "all")
		espXmlmc.OpenElement("searchFilter")
		espXmlmc.SetParam("column", SQLImportConf.SQLConf.FieldID)
		espXmlmc.SetParam("value", contactID)
		if !configMatchLike {
			espXmlmc.SetParam("matchType", "exact")
		}
		espXmlmc.CloseElement("searchFilter")
		espXmlmc.SetParam("maxResults", "1")

		XMLCheckContact, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords2")
		var xmlRespon xmlmcCheckContactResponse
		if xmlmcErr != nil {
			buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. API Invoke Error from [entityBrowseRecords2] [Contact]: "+fmt.Sprintf("%v", xmlmcErr)))
			return intReturn, xmlmcErr
		}
		err = xml.Unmarshal([]byte(XMLCheckContact), &xmlRespon)
		if err != nil {
			buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. Unmarshall Error from [entityBrowseRecords2] [Contact]: "+fmt.Sprintf("%v", err)))
		} else {
			if xmlRespon.MethodResult != constOK {
				err = errors.New(xmlRespon.State.ErrorRet)
				buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. MethodResult not OK from [entityBrowseRecords2] [Contact]: "+fmt.Sprintf("%v", err)))
			} else {
				//-- Check Response
				if xmlRespon.Params.RowData.Row.PKID != "" {
					intReturn, err = strconv.Atoi(xmlRespon.Params.RowData.Row.PKID)
					if err != nil {
						buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. Key Type Conversion Failed [entityBrowseRecords2] [Contact]: "+fmt.Sprintf("%v", err)))
						intReturn = -1
					}
				}
			}
		}

/*	} else {
		espXmlmc.SetParam("searchQuery", SQLImportConf.SQLConf.FieldID+":"+contactID)
		espXmlmc.SetParam("resultsFrom", "0")
		espXmlmc.SetParam("resultsTo", "0")

		XMLCheckContact, xmlmcErr := espXmlmc.Invoke("data", "entitySearch")
		var xmlRespon xmlmcExactContactResponse
		if xmlmcErr != nil {
			buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. API Invoke Error from [entitySearch] [Contact]: "+fmt.Sprintf("%v", xmlmcErr)))
			return intReturn, xmlmcErr
		}
		err = xml.Unmarshal([]byte(XMLCheckContact), &xmlRespon)
		if err != nil {
			buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. Unmarshall Error from [entitySearch] [Contact]: "+fmt.Sprintf("%v", err)))
		} else {
			if xmlRespon.MethodResult != constOK {
				err = errors.New(xmlRespon.State.ErrorRet)
				buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. MethodResult not OK from [entitySearch] [Contact]: "+fmt.Sprintf("%v", err)))
			} else {
				//-- Check Response
				if xmlRespon.PKID != "" {
					intReturn, err = strconv.Atoi(xmlRespon.PKID)
					if err != nil {
						buffer.WriteString(loggerGen(3, "Search for Contact Unsuccessful. Key Type Conversion Failed [entitySearch] [Contact]: "+fmt.Sprintf("%v", err)))
						intReturn = -1
					}
				}
			}
		}

	}
*/
	return intReturn, err
}

//-- Function to search for organisation
func getOrgFromLookup(organisation string, buffer *bytes.Buffer) (string, string) {
	organisationReturn := ""
	compReturn := ""
	//-- Get Value of Attribute
	organisationAttributeName := processComplexField(organisation)
	buffer.WriteString(loggerGen(1, "Looking Up Organisation: "+organisationAttributeName))
	if organisationAttributeName == "" {
		return "", ""
	}
	organisationIsInCache, OrganisationIDCache, CompID := orgInCache(organisationAttributeName)
	//-- Check if we have Cached the organisation already
	if organisationIsInCache {
		organisationReturn = strconv.Itoa(OrganisationIDCache)
		compReturn = CompID
		buffer.WriteString(loggerGen(1, "Found Organisation in Cache: "+organisationReturn+" ("+CompID+")"))
	} else {
		organisationIsOnInstance, OrganisationIDInstance, compIDInstance := searchOrg(organisationAttributeName, buffer)
		//-- If Returned set output
		if organisationIsOnInstance {
			organisationReturn = strconv.Itoa(OrganisationIDInstance)
			compReturn = compIDInstance
		}
	}
	buffer.WriteString(loggerGen(1, "Organisation Lookup found ID: "+organisationReturn))
	return organisationReturn, compReturn
}

func processComplexField(s string) string {
	return html.UnescapeString(s)
}

//-- Function to Check if in Cache
func orgInCache(orgName string) (bool, int, string) {
	boolReturn := false
	intReturn := 0
	stringCompReturn := ""
	mutexOrganisations.Lock()
	//-- Check if in Cache
	for _, organisation := range organisations {
		if organisation.OrgName == orgName {
			boolReturn = true
			intReturn = organisation.OrgID
			stringCompReturn = organisation.CompanyID
			break
		}
	}
	mutexOrganisations.Unlock()
	return boolReturn, intReturn, stringCompReturn
}

//-- Function to Check if the organisation is on the instance
func searchOrg(orgName string, buffer *bytes.Buffer) (bool, int, string) {
	boolReturn := false
	intReturn := 0
	strCompReturn := ""
	espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.InstanceID)
	espXmlmc.SetAPIKey(SQLImportConf.APIKey)
	if orgName == "" {
		return boolReturn, intReturn, ""
	}
	espXmlmc.SetParam("entity", "Organizations")
	espXmlmc.SetParam("matchScope", "all")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("column", "h_organization_name")
	espXmlmc.SetParam("value", orgName)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	XMLOrgSearch, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords2")
	var xmlRespon xmlmcOrganisationListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLOrgSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Organisation: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.OrganizationName != "" {
				if strings.EqualFold(xmlRespon.Params.RowData.Row.OrganizationName, orgName) {
					intReturn = xmlRespon.Params.RowData.Row.OrganizationID
					boolReturn = true

					var xml2Resp xmlmcGroupListResponse
					espXmlmc.ClearParam()
					espXmlmc.SetParam("entity", "Container")
					espXmlmc.SetParam("matchScope", "all")
					espXmlmc.OpenElement("searchFilter")
					espXmlmc.SetParam("column", "h_name")
					espXmlmc.SetParam("value", orgName)
					espXmlmc.CloseElement("searchFilter")

					espXmlmc.OpenElement("searchFilter")
					espXmlmc.SetParam("column", "h_type")
					espXmlmc.SetParam("value", "Organizations")
					espXmlmc.CloseElement("searchFilter")
					XMLOrgSearch, xmlmcOSErr := espXmlmc.Invoke("data", "entityBrowseRecords2")
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
							_ = errors.New(xml2Resp.State.ErrorRet)
							errorCountInc()
						} else {
							strCompReturn = xml2Resp.Params.RowData.Row.GroupID
						}

					}
					//-- Add organisation to Cache
					mutexOrganisations.Lock()
					var newOrganisationForCache organisationListStruct
					newOrganisationForCache.OrgID = intReturn
					newOrganisationForCache.OrgName = orgName
					newOrganisationForCache.CompanyID = strCompReturn
					name := []organisationListStruct{newOrganisationForCache}
					organisations = append(organisations, name...)
					mutexOrganisations.Unlock()

				}
			}
		}
	}

	return boolReturn, intReturn, strCompReturn
}
