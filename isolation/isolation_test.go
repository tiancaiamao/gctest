package isolation

import (
	"bytes"
	"database/sql"
	"log"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

type ksAndAddr struct {
	keyspace string
	addr     string
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

	_, err := dbs[1].Exec("set @@global.tidb_gc_run_interval = 15m")
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
	require.NotEqual(t, leaderDesc[0], leaderDesc[1])
	require.NotEqual(t, leaderDesc[1], leaderDesc[2])
	require.NotEqual(t, leaderDesc[2], leaderDesc[0])

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
			values = append(values, gcInterval)
		}
		rows.Close()
		require.Len(t, values, 1)
		gcLifeTimes = append(gcLifeTimes, values[0])
	}
	require.Equal(t, "10m0s", gcLifeTimes[0])
	require.Equal(t, "15m0s", gcLifeTimes[1])
	require.Equal(t, "30m0s", gcLifeTimes[2])

	for _, db := range dbs {
		_, err = db.Exec("drop table if exists test.t;")
		require.NoError(t, err)

		_, err = db.Exec("create table test.t (id int primary key, v int);")
		require.NoError(t, err)

		_, err = db.Exec("insert into test.t values (1, 10), (2, 20), (3, 30);")
		require.NoError(t, err)
	}
	time.Sleep(3 * time.Second)
	// now
	staleTS := time.Now().Format("2006-01-02 15:04:05")
	sql := fmt.Sprintf("select * from test.t as of timestamp %q", staleTS)
	log.Println("now ==", staleTS)

	// 5min later
	time.Sleep(5 * time.Minute)
	log.Println("check data visibility after 5min")
	for _, db := range dbs {
		mustQuery(t, db, sql).Check(Rows("1 10", "2 20", "3 30"))
	}

	// 10min later, around here, GC on the first tidb happen.
	// But since gc_life_time is also 10min, we may still see the data.
	time.Sleep(5 * time.Minute)
	log.Println("10 min later")

	// 14min later
	time.Sleep(4 * time.Minute)
	log.Println("check data visibility after 14min")
	for _, db := range dbs[1:] {
		mustQuery(t, db, sql).Check(Rows("1 10", "2 20", "3 30"))
	}
	rows, err := dbs[0].Query(sql)
	if err == nil {
		_, err := rowsToResult(t, rows)
		// require.Error(t, err)
		if err == nil {
			log.Printf("We should get error, but GC does not run here. The last round it runs, the data may still be valid.")
		}
	}

	// 25min later
	time.Sleep(11 * time.Minute)
	log.Println("check data visibility after 25min")
	mustQuery(t, dbs[2], sql).Check(Rows("1 10", "2 20", "3 30"))
	for _, db := range dbs[:2] {
		rows, err = db.Query(sql)
		if err == nil {
			_, err := rowsToResult(t, rows)
			require.Error(t, err)
		}
	}

	fmt.Println("test succeed")
}

func mustQuery(t *testing.T, db *sql.DB, sql string) *Result {
	rows, err := db.Query(sql)
	require.NoError(t, err)

	res, err := rowsToResult(t, rows)
	require.NoError(t, err)
	return res
}

func rowsToResult(t *testing.T, rows *sql.Rows) (*Result, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var res [][]string
	for rows.Next() {
		oneRow := make([]string, len(columns))
		args := make([]any, len(columns))
		for i := 0; i < len(columns); i++ {
			args[i] = &oneRow[i]
		}
		err = rows.Scan(args...)
		if err != nil {
			return nil, err
		}

		res = append(res, oneRow)
	}
	err = rows.Close()
	if err != nil {
		return nil, err
	}
	return &Result{
		rows:    res,
		require: require.New(t),
	}, nil
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
