package longtxn

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap/errors"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	// Setup a sql.DB instance
	addr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", "root", "", "127.0.0.1", "4000", "test")
	statusURL := "http://127.0.0.1:10080"

	db, err := sql.Open("mysql", addr)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	_, err = db.Exec("create table if not exists t (id int primary key, v int)")
	require.NoError(t, err)

	_, err = db.Exec("begin;")
	require.NoError(t, err)

	res, err := db.Query("select * from test.t where id = 1 for update;")
	require.NoError(t, err)
	require.NoError(t, res.Close())

	rows, err := db.Query("select @@tidb_current_ts")
	require.NoError(t, err)
	var tsStr string
	for rows.Next() {
		err = rows.Scan(&tsStr)
		require.NoError(t, err)
	}
	require.NoError(t, rows.Close())

	txnTS, err := strconv.ParseUint(tsStr, 10, 64)
	require.NoError(t, err)

	// Check something else in background goroutine
	// We need to observe:
	// 1. gc is running
	// 2. gc is blocked by the transaction
	ch := make(chan error)
	go func() {
		defer close(ch)
		res, err = db.Query("select sleep(30*60);") // Wait for 30 minutes in the transaction
		if err != nil {
			ch <- err
			return
		}

		_, err = db.Exec("update test.t set v = v + 1 where id = 1;")
		if err != nil {
			ch <- err
			return
		}

		_, err = db.Exec("commit;")
		if err != nil {
			ch <- err
			return
		}
	}()

	cli := newHTTPClient(statusURL)
	initState, err := cli.getGCState()
	require.NoError(t, err)

	var gcRunned bool
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			state, err := cli.getGCState()
			require.NoError(t, err)
			require.True(t, state.TxnSafePoint <= txnTS)
			if state.TxnSafePoint != initState.TxnSafePoint {
				gcRunned = true
			}
			fmt.Println("current GC state", state)
		case <-ch:
			require.True(t, gcRunned)
			return
		}
	}
}

type Client struct {
	*http.Client
	url string
}

func newHTTPClient(statusURL string) *Client {
	return &Client{
		Client: http.DefaultClient,
		url:    statusURL,
	}
}

func (c *Client) getGCState() (GCState, error) {
	var state GCState
	data, err := c.Get(fmt.Sprintf("%s/txn-gc-states", c.url))
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

// Get sends a HTTP GET request to the specified URL.
func (c *Client) Get(url string) ([]byte, error) {
	resp, err := c.httpRequest(url, http.MethodGet, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Read GET response failed")
	}
	if !isHTTPSuccess(resp.StatusCode) {
		return nil, errors.New(fmt.Sprintf("GET request \"%s\", got %v %s", url, resp.StatusCode, string(res)))
	}
	return res, nil
}

func isHTTPSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func (c *Client) httpRequest(url string, method string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	if bodyType != "" {
		req.Header.Set("Content-Type", bodyType)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.Annotate(err, fmt.Sprintf("do http request failed, [%s] %s", method, url))
	}
	return resp, nil
}

type GCBarrier struct {
	BarrierID string
	BarrierTS uint64
	// Nil means never expiring.
	ExpirationTime *time.Time
}

type GCState struct {
	KeyspaceID      uint32
	IsKeyspaceLevel bool
	TxnSafePoint    uint64
	GCSafePoint     uint64
	GCBarriers      []*GCBarrier
}
