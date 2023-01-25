/*
HarperDB Go API

Notes:

	Record Ids should not be numeric. Auto generated hash is alpha-numeric.
	Hash attributes are stored in string type fields. JSON unmarshal from int to string has problems.
*/
package harper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// WARNING - if program has multiple goroutines with different values of URL, additional logic must be added.
// Depending on requirements, the URL could be added to Harper type and set by each goroutine

var URL string = "http://localhost:9925"

// Harper type used to execute API calls via Run method
// In multi goroutine program such as http server, each goroutine can use unique instance.
// The Client attribute can be point to a common http.Client which is an expensive object to create
// and is safe for concurrent access by multiple goroutines.
type Harper struct {
	Client    *http.Client // can create with new(http.Client)
	AuthToken string       // Authorization Bearer token added to http header (see tokens.go/GetOpToken func)
	Schema    string       // Default schema, but requests can also specify schema
	Debug     bool         // If true, Run method will display request/response to screen in formatted JSON
}

// run harperdb api request
func (harper *Harper) Run(payload interface{}) ([]byte, error) {
	jsonContent, err := json.Marshal(&payload) // -> []byte
	if err != nil {
		log.Println("HarperDB JSON Marshal Failed:", err)
		return nil, err
	}
	if harper.Debug {
		fmt.Println("--- REQUEST ---")
		fmt.Println(FmtJSON(jsonContent))
	}
	reqBody := bytes.NewReader(jsonContent) // -> io.Reader

	req, err := http.NewRequest("POST", URL, reqBody)
	req.Header.Add("Content-Type", "application/json")
	if harper.AuthToken != "" { // Create Tokens does not require authentication in header
		req.Header.Add("Authorization", "Bearer "+harper.AuthToken)
	}
	resp, doErr := harper.Client.Do(req)
	defer func() {
		if doErr == nil {
			resp.Body.Close()
		}
	}()
	if resp.StatusCode != http.StatusOK || doErr != nil {
		log.Println("XXX -- HarperDB Request Failed, Status:", resp.StatusCode, " ", resp.Status, " - ", doErr, " --- XXX")
		log.Println(FmtJSON(jsonContent))
		return nil, doErr
	}
	result, err := ioutil.ReadAll(resp.Body) // -> []byte
	if err != nil {
		log.Println("HarperDB Read Response Failed:", err)
	}
	if harper.Debug {
		fmt.Println("--- RESPONSE ---")
		fmt.Println(FmtJSON(result))
	}
	return result, err
}

// ---------------------------------------------------------------------------------------------

// AdminReq is request body for admin operations.
// Example ops: create schema, create_table, describe_table, drop_table, add_user, user_info, etc.
// Caller loads instance of AdminReq with required values, runs Process method
// NOTE - if operation requires schema attr, it must be loaded, will not default to harper.Schema.
type AdminReq struct {
	Operation     string `json:"operation"`
	Schema        string `json:"schema,omitempty"`
	Table         string `json:"table,omitempty"`
	Attribute     string `json:"attribute,omitempty"`
	HashAttribute string `json:"hash_attribute,omitempty"`
	Role          string `json:"role,omitempty"`
	UserName      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Active        *bool  `json:"active,omitempty"` // use &IsTrue, &IsFalse or nil to omit
}

func (req *AdminReq) Process(harper *Harper) ([]byte, error) {
	return harper.Run(req)
}

// ---------------------------------------------------------------------------------------------

// result obj for data updates (insert, update, upsert, delete)
type UpdtResult struct {
	Message        string   `json:"message"`
	InsertedHashes []string `json:"inserted_hashes"` // WARNING - Rec Ids should contain alpha-numeric, so not returned as number type
	UpdatedHashes  []string `json:"update_hashes"`
	DeletedHashes  []string `json:"deleted_hashes"`
	SkippedHashes  []string `json:"skipped_hashes"`
}

// request body for data updates (insert, update, upsert, delete), sql or nosql
type UpdtReq struct {
	Operation  string      `json:"operation"`
	Schema     string      `json:"schema,omitempty"` // if empty, loaded from Harper parm
	Table      string      `json:"table,omitempty"`
	Records    interface{} `json:"records,omitempty"`
	HashValues []string    `json:"hash_values,omitempty"`
	Sql        string      `json:"sql,omitempty"`
}

func (req *UpdtReq) Process(harper *Harper) (*UpdtResult, error) {
	if req.Schema == "" {
		req.Schema = harper.Schema
	}
	response, err := harper.Run(req)
	var result UpdtResult
	json.Unmarshal(response, &result)
	if len(result.SkippedHashes) > 0 && err == nil {
		err = errors.New("skipped_hashes")
	}
	return &result, err
}

// Update records - convenience func
func Update(harper *Harper, table string, records interface{}) (*UpdtResult, error) {
	request := UpdtReq{
		Operation: "update",
		Table:     table,
		Records:   records,
	}
	response, err := request.Process(harper)
	return response, err
}

// Update single record, api requires array - convenience func
func UpdateOne(harper *Harper, table string, record interface{}) (*UpdtResult, error) {
	records := []interface{}{record}
	request := UpdtReq{
		Operation: "update",
		Table:     table,
		Records:   records,
	}
	response, err := request.Process(harper)
	return response, err
}

// Update records using sql - convenience func
func UpdateSql(harper *Harper, sql string) (*UpdtResult, error) {
	request := UpdtReq{
		Operation: "sql",
		Sql:       sql,
	}
	response, err := request.Process(harper)
	return response, err
}

// Insert records - convenience func
func Insert(harper *Harper, table string, records interface{}) (*UpdtResult, error) {
	request := UpdtReq{
		Operation: "insert",
		Table:     table,
		Records:   records,
	}
	response, err := request.Process(harper)
	return response, err
}

// Insert single record, api requires array - convenience func
func InsertOne(harper *Harper, table string, record interface{}) (*UpdtResult, error) {
	records := []interface{}{record}
	request := UpdtReq{
		Operation: "insert",
		Table:     table,
		Records:   records,
	}
	response, err := request.Process(harper)
	return response, err
}

// ---------------------------------------------------------------------------------------------

type ReadReq struct {
	Operation       string   `json:"operation"`
	Schema          string   `json:"schema,omitempty"` // if empty, loaded from Harper parm
	Table           string   `json:"table,omitempty"`
	HashValues      []string `json:"hash_values,omitempty"`
	GetAttributes   []string `json:"get_attributes,omitempty"`
	SearchAttribute []string `json:"search_attribute,omitempty"`
	SearchValue     []string `json:"search_value,omitempty"`
	Sql             string   `json:"sql,omitempty"`
}

// response is typically slice of records
func (req *ReadReq) Process(harper *Harper, result interface{}) error {
	if req.Schema == "" {
		req.Schema = harper.Schema
	}
	response, err := harper.Run(req)
	if err != nil {
		return err
	}
	err = json.Unmarshal(response, result)
	return err
}

// get record(s) by hash value (id), all fields returned - convenience func
func Get(harper *Harper, table string, result interface{}, recIds ...string) error {
	request := ReadReq{
		Operation:     "search_by_hash",
		Table:         table,
		HashValues:    recIds,
		GetAttributes: []string{"*"},
	}
	return request.Process(harper, result)
}

// select records using sql - convenience func
func Select(harper *Harper, sql string, result interface{}) error {
	request := ReadReq{
		Operation: "sql",
		Sql:       sql,
	}
	return request.Process(harper, result)
}

// ---------------------------------------------------------------------------------------------

// response for bulk operations (contains job id)
type bulkResponse struct {
	Message string `json:"message"`
}

func (br bulkResponse) getJobId() string {
	i := strings.Index(br.Message, "id")
	if i == -1 {
		return ("not found")
	}
	return br.Message[i+3:]
}

// process csv data load (insert/update) requests, returns job-id
func CsvDataLoad(harper *Harper, table, action string, data [][]string) (string, error) {
	var csvData strings.Builder // provides efficient way to append values to long string
	for _, recValues := range data {
		csvLine := CreateCSVLine(recValues)
		csvData.WriteString(csvLine)
	}
	payload := map[string]interface{}{
		"schema":    harper.Schema,
		"operation": "csv_data_load",
		"action":    action,
		"table":     table,
		"data":      csvData.String(),
	}
	response, err := harper.Run(&payload)
	if err != nil {
		return "", err
	}
	var result bulkResponse
	json.Unmarshal(response, &result)
	return result.getJobId(), nil
}
