# HarperDB Mini SDK

Files contained in this repository are not intended to be treated as a 3rd party stand alone package. They should be reviewed, modified, and incorporated as desired into the developer's own code base. The goal is to take advantage of HDB's simple and elegant HTTP based interface. Code size is very small.  

I currently have api.go, tokens.go, and util.go in a package used for testing.

All features in program sample.go work.

I only recently started working with HDB. I expect to add and change features possibly breaking compatibility with code contained here. 

## File Descriptions  

## api.go - aprox 300 lines

This is the file for interacting with the HDB API. It contains 4 primary components. These components and several "convenience" functions provide access to all features in following api categories:  

* Schemas and Tables - uses AdminReq component
* NoSQL Operations - uses UpdtReq & ReadReq components
* SQL Operations - uses UpdtReq & ReadReq components
* User Operations - uses AdminReq component
  
Working on Bulk Operations component.
  
## util.go - aprox 100 lines  

This file contains several utility functions including:
* Convert HDB created/updated date-time to std Go time object 
* Monitor status of running batch job, until complete
* JSON formatter for displaying requests/responses in ez to read format

## tokens.go - aprox 50 lines  

Provides functions to manage authorization tokens. User/Password and tokens are stored/read from .json file.
I use file path of creds/harper_creds.json with dir permission of 700.
Applications needing operation token can use: harper.GetOpToken().

## harper-tokens.go - aprox 100 lines

Stand alone program that uses functions in tokens.go to create and refresh authorization tokens.
This program could run on automated schedule using tool like crontab (on Linux).
> ./harper-token -create (to create operation and refresh tokens)  
> ./harper-token -refresh (to refresh operation token using refresh token)

## harper-sample.go  
  
Stand alone program that uses types/functions in api.go.

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

**UpdtReq type**

Used for both SQL and NonSQL update requests.
This type can be used directly by creating instance with necessay attributes populated and calling .Process() method. The only parameter is pointer to Harper instance containing http.Client, Schema*, & Authorization token.  
The following convenience funcs can also be used. They use the UpdtReq type to process.

* Update - update records using "update" operation
* UpdateOne - update a single record using "update" operation
* UpdateSql - update using "sql" operation
* Insert - insert records using "insert" operation
* InsertOne - insert a single record using "insert" operation

All Update Requests return *UpdtResult.
```
type UpdtResult struct {
	Message        string   `json:"message"`
	InsertedHashes []string `json:"inserted_hashes"`
	UpdatedHashes  []string `json:"update_hashes"`
	DeletedHashes  []string `json:"deleted_hashes"`
	SkippedHashes  []string `json:"skipped_hashes"`
}
```
WARNING - Hash (id) values should be alpha-numeric, so not returned as number type

**ReadReq type**

Used for both SQL and NonSQL read requests.
This type can be used directly by creating instance with necessary attributes populated and calling .Process() method for which the only parameter is pointer to Harper instance containing http.Client, Schema*, & authorization token.    
The following convenience funcs can also be used. They use the ReadReq type to process.

* Get - get record(s) using "search_by_hash operation, all fields returned
* Select - select records using "sql" operation


**NOTE - The harper parm passed to most convenience funcs must contain the "schema" because the func parms do not specify schema. Exception is sql requests where the schema must be included in the sql.*
