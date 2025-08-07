package hellolocal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/pingcap/endless/pkg/client/tidb"
	httputil "github.com/pingcap/test-infra/sdk/pkg/util/http"
	"github.com/pingcap/test-infra/sdk/resource"
)

var _ = Describe("Real world test case", func() {
	Context("assume that you have created your testbed", func() {
		It("should pass if you have setup your MYSQL_PROXY correctly #hello_local_2#", func() {
			ctx := suiteTestCtx
			tc := ctx.Resource("tc").(resource.TiDBCluster)

			// Please ensure that you have setup .env properly,

			// Setup a sql.DB instance
			db, err := tc.GetDB("test")
			// Assert the err should be nil
			Expect(err).Should(Succeed())

			_, err = db.Exec("create table if not exists t (id int primary key, v int)")
			Expect(err).Should(Succeed())

			_, err = db.Exec("begin;")
			Expect(err).Should(Succeed())

			res, err := db.Query("select * from test.t where id = 1 for update;")
			Expect(err).Should(Succeed())
			Expect(res.Close()).Should(Succeed())

			rows, err := db.Query("select @@tidb_current_ts")
			Expect(err).Should(Succeed())
			var tsStr string
			for rows.Next() {
				err = rows.Scan(&tsStr)
				Expect(err).Should(Succeed())
			}
			Expect(rows.Close()).Should(Succeed())

			txnTS, err := strconv.ParseUint(tsStr, 10, 64)
			Expect(err).Should(Succeed())

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

			statusURL, err := tc.ServiceURL(resource.DBStatus)
			Expect(err).Should(Succeed())
			addr := statusURL.String()

			cli := httputil.NewHTTPClient(http.DefaultClient)
			initState, err := getGCState(cli, addr)
			Expect(err).Should(Succeed())

			var gcRunned bool
			ticker := time.NewTicker(10 * time.Second)
			for {
				select {
				case <-ticker.C:
					state, err := getGCState(cli, addr)
					Expect(err).Should(Succeed())
					Expect(state.TxnSafePoint <= txnTS).Should(BeTrue())
					if state.TxnSafePoint != initState.TxnSafePoint {
						gcRunned = true
					}
					fmt.Println("current GC state", state)
				case <-ch:
					Expect(gcRunned).Should(BeTrue())
					return
				}
			}
		})
	})
})

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
