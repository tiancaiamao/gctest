package isolation

import (
	"testing"
	"bytes"
	"strings"
	"database/sql"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
	_ "github.com/go-sql-driver/mysql"
)

type ksAndAddr struct {
	keyspace string
	addr string
}

var config []ksAndAddr

func init() {
	config = []ksAndAddr{
		{"admin", "127.0.0.1:4001"},
		{"user1", "127.0.0.1:4003"},
		{"user2", "127.0.0.1:4005"},
	}
}

func TestAll(t *testing.T) {
	// Setup a sql.DB instance
	var dbs []*sql.DB
	for _, ks := range config {
		addr := fmt.Sprintf("%s:%s@tcp(%s)/%s", "root", "", ks.addr, "test")
		db, err := sql.Open("mysql", addr)
		require.NoError(t, err)
		require.NoError(t, db.Ping())
		dbs = append(dbs, db)
	}

	_, err := dbs[1].Exec("set @@global.tidb_gc_run_interval = 17m")
	require.NoError(t, err)
	dbs[2].Exec("set @@global.tidb_gc_run_interval = 30m")
	require.NoError(t, err)

	// Checking tikv gc leader is different for those different keyspaces
	var leaderDesc []string
	for _, db := range dbs {
		rows, err := db.Query("select variable_name, variable_value from mysql.tidb where variable_name = 'tikv_gc_leader_desc'")
		require.NoError(t, err)
		var values []string
		for rows.Next() {
			var name, value string
			err = rows.Scan(&name, &value)
			require.NoError(t, err)
			values = append(values, value)
		}
		rows.Close()
		require.Len(t, values, 1)
		leaderDesc = append(leaderDesc, values[0])
	}
	require.Equal(t, "10m0s", leaderDesc[0])
	require.Equal(t, "17m0s", leaderDesc[1])
	require.Equal(t, "30m0s", leaderDesc[2])

	// Check they are using different configuration
	var gcLifeTimes []string
	for _, db := range dbs {
		rows, err := db.Query("select @@global.tidb_gc_life_time, @@global.tidb_gc_run_interval")
		require.NoError(t, err)
		var values []string
		for rows.Next() {
			var gcLifeTime, gcInterval string
			err = rows.Scan(&gcLifeTime, &gcInterval)
			require.NoError(t, err)
			values = append(values, gcLifeTime)
		}
		rows.Close()
		require.Len(t, values, 1)
		gcLifeTimes = append(gcLifeTimes, values[0])
	}
	require.NotEqual(t, gcLifeTimes[0], gcLifeTimes[1])
	require.NotEqual(t, gcLifeTimes[1], gcLifeTimes[2])
	require.NotEqual(t, gcLifeTimes[2], gcLifeTimes[0])

	for _, db := range dbs {
		_, err = db.Exec("create table if not exists test.t (id int primary key, v int);")
		require.NoError(t, err)

		_, err = db.Exec("insert into test.t values (1, 10), (2, 20), (3, 30);")
		require.NoError(t, err)
	}

	// 5min later
	time.Sleep(5 * time.Minute)
	fmt.Println("check data visibility after 5min")

	for _, db := range dbs {
		mustQuery(t, db, "select * from test.t").Check(Rows("1 10\n2 20\n3 30"))
	}

	// 15min later
	time.Sleep(10 * time.Minute)
	fmt.Println("check data visibility after 15min")

	for _, db := range dbs[1:] {
		mustQuery(t, db, "select * from test.t").Check(Rows("1 10\n2 20\n3 30"))
	}
	_, err = dbs[0].Query("select * from test.t")
	require.Error(t, err)

	// 25min later
	time.Sleep(10 * time.Minute)
	fmt.Println("check data visibility after 25min")

	for _, db := range dbs[0:2] {
		_, err = db.Query("select * from test.t")
		require.Error(t, err)
	}
	mustQuery(t, dbs[2], "select * from test.t").Check(Rows("1 10\n2 20\n3 30"))

	fmt.Println("test succeed")
}

func mustQuery(t *testing.T, db *sql.DB, sql string) *Result {
	rows, err := db.Query(sql)
	require.NoError(t, err)

	columns, err := rows.Columns()
	require.NoError(t, err)

	var res [][]string
	for rows.Next() {
		oneRow := make([]string, len(columns))
		args := make([]any, len(columns))
		for i:=0; i<len(columns); i++ {
			args[i] = &oneRow[i]
		}
		err = rows.Scan(args...)
		require.NoError(t, err)

		res = append(res, oneRow)
	}
	require.NoError(t, rows.Close())
	return &Result{
		rows: res,
	}
}


// Result is the result returned by MustQuery.
type Result struct {
	rows    [][]string
	comment string
	require *require.Assertions
}

// Check asserts the result equals the expected results.
func (res *Result) Check(expected [][]any) {
	resBuff := bytes.NewBufferString("")
	for _, row := range res.rows {
		_, _ = fmt.Fprintf(resBuff, "%s\n", row)
	}

	needBuff := bytes.NewBufferString("")
	for _, row := range expected {
		_, _ = fmt.Fprintf(needBuff, "%s\n", row)
	}

	res.require.Equal(needBuff.String(), resBuff.String(), res.comment)
}

// Equal check whether the result equals the expected results.
func (res *Result) Equal(expected [][]any) bool {
	resBuff := bytes.NewBufferString("")
	for _, row := range res.rows {
		_, _ = fmt.Fprintf(resBuff, "%s\n", row)
	}

	needBuff := bytes.NewBufferString("")
	for _, row := range expected {
		_, _ = fmt.Fprintf(needBuff, "%s\n", row)
	}

	return bytes.Equal(needBuff.Bytes(), resBuff.Bytes())
}

func Rows(args ...string) [][]any {
	return RowsWithSep(" ", args...)
}

// RowsWithSep is a convenient function to wrap args to a slice of []interface.
// The arg represents a row, split by sep.
func RowsWithSep(sep string, args ...string) [][]any {
	rows := make([][]any, len(args))
	for i, v := range args {
		parts := strings.Split(v, sep)
		row := make([]any, len(parts))
		for j, s := range parts {
			row[j] = s
		}
		rows[i] = row
	}
	return rows
}
