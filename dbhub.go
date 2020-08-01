package dbhub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	com "github.com/sqlitebrowser/dbhub.io/common"
)

const (
	LibraryVersion = "0.0.1"
)

// New creates a new DBHub.io connection object.  It doesn't connect to DBHub.io to do this.
func New(key string) (Connection, error) {
	c := Connection{
		APIKey: key,
		Server: "https://api.dbhub.io",
	}
	return c, nil
}

// ChangeServer changes the address all Queries will be sent to.  Useful for testing and development.
func (c *Connection) ChangeServer(s string) {
	c.Server = s
}

// Columns returns the column information for a given table or view
func (c Connection) Columns(dbowner, dbname, table string) (columns []com.APIJSONColumn, err error) {
	// Prepare the API parameters
	data := url.Values{}
	data.Set("apikey", c.APIKey)
	data.Set("dbowner", dbowner)
	data.Set("dbname", dbname)
	data.Set("table", table)

	// Fetch the list of columns
	var resp *http.Response
	queryUrl := c.Server + "/v1/columns"
	resp, err = sendRequest(queryUrl, data)
	if err != nil {
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Convert the response into the list of columns
	err = json.NewDecoder(resp.Body).Decode(&columns)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Indexes returns the list of indexes present in the database, along with the table they belong to
func (c Connection) Indexes(dbowner, dbname string) (idx map[string]string, err error) {
	// Prepare the API parameters
	data := url.Values{}
	data.Set("apikey", c.APIKey)
	data.Set("dbowner", dbowner)
	data.Set("dbname", dbname)

	// Fetch the list of indexes
	var resp *http.Response
	queryUrl := c.Server + "/v1/indexes"
	resp, err = sendRequest(queryUrl, data)
	if err != nil {
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Convert the response into the list of indexes
	err = json.NewDecoder(resp.Body).Decode(&idx)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Query runs a SQL query (SELECT only) on the chosen database, returning the results.
// The "blobBase64" boolean specifies whether BLOB data fields should be base64 encoded in the output, or just skipped
// using an empty string as a placeholder.
func (c Connection) Query(dbowner, dbname string, blobBase64 bool, sql string) (out Results, err error) {
	// Prepare the API parameters
	data := url.Values{}
	data.Set("apikey", c.APIKey)
	data.Set("dbowner", dbowner)
	data.Set("dbname", dbname)
	data.Set("sql", base64.StdEncoding.EncodeToString([]byte(sql)))

	// Run the query on the remote database
	var resp *http.Response
	queryUrl := c.Server + "/v1/query"
	resp, err = sendRequest(queryUrl, data)
	if err != nil {
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// The query ran successfully, so prepare and return the results
	var returnedData []com.DataRow
	err = json.NewDecoder(resp.Body).Decode(&returnedData)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the result list
	for _, j := range returnedData {

		// Construct a single row
		var oneRow ResultRow
		for _, l := range j {
			switch l.Type {
			case com.Float, com.Integer, com.Text:
				// Float, integer, and text fields are added to the output
				oneRow.Fields = append(oneRow.Fields, fmt.Sprint(l.Value))
			case com.Binary:
				// BLOB data is optionally Base64 encoded, or just skipped (using an empty string as placeholder)
				if blobBase64 {
					// Safety check. Make sure we've received a string
					if _, ok := l.Value.(string); ok {
						oneRow.Fields = append(oneRow.Fields, base64.StdEncoding.EncodeToString([]byte(l.Value.(string))))
					} else {
						oneRow.Fields = append(oneRow.Fields, fmt.Sprintf("unexpected data type '%T' for returned BLOB", l.Value))
					}
				} else {
					oneRow.Fields = append(oneRow.Fields, "")
				}
			default:
				// All other value types are just output as an empty string (for now)
				oneRow.Fields = append(oneRow.Fields, "")
			}
		}
		// Add the row to the output list
		out.Rows = append(out.Rows, oneRow)
	}
	return
}

// Tables returns the list of tables in the database
func (c Connection) Tables(dbowner, dbname string) (tbl []string, err error) {
	// Prepare the API parameters
	data := url.Values{}
	data.Set("apikey", c.APIKey)
	data.Set("dbowner", dbowner)
	data.Set("dbname", dbname)

	// Fetch the list of tables
	var resp *http.Response
	queryUrl := c.Server + "/v1/tables"
	resp, err = sendRequest(queryUrl, data)
	if err != nil {
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Convert the response into the list of tables
	err = json.NewDecoder(resp.Body).Decode(&tbl)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Views returns the list of views in the database
func (c Connection) Views(dbowner, dbname string) (vws []string, err error) {
	// Prepare the API parameters
	data := url.Values{}
	data.Set("apikey", c.APIKey)
	data.Set("dbowner", dbowner)
	data.Set("dbname", dbname)

	// Fetch the list of views
	var resp *http.Response
	queryUrl := c.Server + "/v1/views"
	resp, err = sendRequest(queryUrl, data)
	if err != nil {
		return
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Convert the response into the list of views
	err = json.NewDecoder(resp.Body).Decode(&vws)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// TODO: Create function(s) for listing indexes in the remote database

// TODO: Create function to list columns in a table (or view?)

// TODO: Create function for returning a list of available databases

// TODO: Create function for downloading complete database

// TODO: Create function for uploading complete database

// TODO: Create function for retrieving database details (size, branch, commit list, whatever else is useful)

// TODO: Make a reasonable example application written in Go
