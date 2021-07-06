# CHANGELOG

## 1.6.4 (July 6th, 2021)

Change:

- Rebuilt using latest version of goApiLib, to fix possible issue with connections via a proxy

## 1.6.3 (June 10th, 2021)

Features:

- Since the introduction of entityBrowseRecords2, the search on organisation is done on a LIKE format instead of an exact match (case-insentive). Have now made exact match default behaviour, if you wish to revert to the behaviour from v1.0.3 until now, use the -matchorglike=true command line parameter.

## 1.6.2 (April 15th, 2020)

Change:

- Updated code to support Core application and platform changes

## 1.6.1 (March 2nd, 2020)

Fixes:

- Updating Contact Organization link instead of inserting.

## 1.6.0 (October 15th, 2019)

Features:

- Added option for CLI output to be in the terminal-default colours only
- Added additional log output when Create is selected in config, and Contact already exists in Hornbill
- Removed hard-coded list of mapped columns, so now any Contact column can be populated entirely by config

## 1.5.1 (October 3rd, 2019)

Fixes:

- reworking of 1.4.0 solution to future-proof it

## 1.5.0 (September 6th, 2019)

Features:

- Allow for individual contact subscription to a Service. Requires the numeric Service ID to be obtained (eg from the URL (123 in ...hornbill.com/INSTANCENAME/servicemanager/service/view/123/) )

## 1.4.0 (July 30th, 2019)

Features:

- Since the introduction of entityBrowseRecords2, the search is done on a LIKE format instead of an exact match (case-insentive). Have now made exact match default behaviour, if you wish to revert to the behaviour from v1.0.3 until now, use the -matchlike=true command line parameter.

## 1.3.0 (March 6th, 2019)

Features:

- Removed zone command line parameter, tool no longer needs to be provided instance zone
- Added support for MySQL 8+
- Corrected logic for logging number of updated or created contacts
- Improved error logging
- Refactored code

## 1.2.2 (December 5th, 2018)

Fixes:

  - tweak to correct privilege level

## 1.2.1 (November 20th, 2018)

Fixes:

  - tweak to properly relate organisation to contact

## 1.2.0 (November 8th, 2018)

Features:

  - Added ability to allow the contacts to see organisation level call viewing via portal (CustomerPortalOrgView: true; CustomerPortalOrgViewRevoke: false).
  - Added ability to revoke those rights for imported contacts  (CustomerPortalOrgView: true; CustomerPortalOrgViewRevoke: true).
  - CustomerPortalOrgView: false will not modify the organisation level call viewing

## 1.1.1 (November 5th, 2018)

Fixes:

  - tweak to fix new search field allocation

## 1.1.0 (November 5th, 2018)

Features:

  - Added ability to modify search field (h_logon_id is not a mandatory field and not necessarily used).

## 1.0.3 (September 26th, 2018)

Fixes:

  - Recoding to use entityBrowseRecords2 instead of entityBrowseRecords.

## 1.0.2 (May 2nd, 2018)

Features:

  - Added 6 Custom fields.

## 1.0.1 (October 30th, 2017)

Features:

  - Added new way of linking to organisation.

## 1.0.0 (April 6th, 2017)

Features:

  - Initial Release
