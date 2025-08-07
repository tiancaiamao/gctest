package longtxn

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"testing"
	"time"

	httputil "github.com/pingcap/test-infra/sdk/pkg/util/http"
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

	cli := httputil.NewHTTPClient(http.DefaultClient)
	initState, err := getGCState(cli, statusURL)
	require.NoError(t, err)

	var gcRunned bool
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			state, err := getGCState(cli, statusURL)
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

func getGCState(cli *httputil.Client, addr string) (GCState, error) {
	var state GCState
	data, err := cli.Get(fmt.Sprintf("%s/txn-gc-states", addr))
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
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
