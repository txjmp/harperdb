# HarperDB Mini SDK

Files contained in this repository are not intended to be treated as a 3rd party stand alone package. They should be reviewed, modified, and incorporated as desired into the developer's own code base. The goal is to take advantage of HDB's simple and elegant HTTP based interface. Code size is very small.  
  
## File Descriptions  

## api.go - aprox 300 lines

This is the file for interacting with the HDB API. It contains 4 primary components. These components and several "convenience" functions provide access to all features in following api categories:  

* Schemas and Tables
* NoSQL Operations
* SQL Operations
* User Operations
* Bulk Operations (limited)  
  
## util.go - aprox 100 lines  

This file contains several utility functions including:
* Convert HDB created/updated date-time to std Go time object 
* Monitor status of running batch job, until complete

## tokens.go - aprox 50 lines  

Provides functions to manage authorization tokens. 

## harper-tokens.go - aprox 100 lines

Stand alone program that uses functions in tokens.go to create and refresh authorization tokens.

# Details of api.go

These 4 components provide the core functionality:
* Harper type
* AdminReq type
* UpdtReq type
* ReadReq type

**Harper type**

Contains attributes and method for executing API Http calls

* Run method executes api calls
* Client - http client object used by Run (multiple goroutines can use)
* AuthToken - http header authorization token
* Schema - default schema used, unless specified by process requests
* Debug - boolean, indicates if debugging is on or off, formatted json request/response displayed to screen

**AdminReq type**

Used for Schema, Table, User operations

- Process method executes Admin Requests
- Attributes required for operations such as (not limited to):
    * Create Schema
    * Create Table
    * Add User
