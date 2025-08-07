package hellolocal_test

import (
	"database/sql"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/pingcap/endless/pkg/client/tidb"
	"github.com/pingcap/endless/testcase/sqlfeature/utils"
	"github.com/pingcap/test-infra/sdk/resource"
)

var _ = Describe("Real world test case with local resources", func() {
	Context("assume that you have created your testbed", func() {
		It("should pass if you have setup your MYSQL_PROXY correctly #hello_local_2#", func() {
			ctx := suiteTestCtx
			tc := ctx.Resource("tc").(resource.TiDBCloudNextGenCluster)

			// Please ensure that you have setup .env properly,
			// Setup a sql.DB instance
			var dbs []*sql.DB
			for _, ks := range []string{"ks1", "ks2", "ks3"} {
				group, err := tc.BindTiDBGroup(ks)
				Expect(err).Should(Succeed())
				db, err := group.GetDB("test")
				Expect(err).Should(Succeed())
				dbs = append(dbs, db)
			}

			_, err := dbs[1].Exec("set @@global.tidb_gc_run_interval = 17m")
			Expect(err).Should(Succeed())
			dbs[2].Exec("set @@global.tidb_gc_run_interval = 30m")
			Expect(err).Should(Succeed())

			// Checking tikv gc leader is different for those different keyspaces
			var leaderDesc []string
			for _, db := range dbs {
				rows, err := db.Query("select variable_name, variable_value from mysql.tidb where variable_name = 'tikv_gc_leader_desc'")
				Expect(err).Should(Succeed())
				var values []string
				for rows.Next() {
					var name, value string
					err = rows.Scan(&name, &value)
					Expect(err).Should(Succeed())
					values = append(values, value)
				}
				rows.Close()
				Expect(len(values)).Should(Equal(1))
				leaderDesc = append(leaderDesc, values[0])
			}
			Expect(leaderDesc[0]).Should(Equal("10m0s"))
			Expect(leaderDesc[1]).Should(Equal("17m0s"))
			Expect(leaderDesc[2]).Should(Equal("30m0s"))

			// Check they are using different configuration
			var gcLifeTimes []string
			for _, db := range dbs {
				rows, err := db.Query("select @@global.tidb_gc_life_time, @@global.tidb_gc_run_interval")
				Expect(err).Should(Succeed())
				var values []string
				for rows.Next() {
					var gcLifeTime, gcInterval string
					err = rows.Scan(&gcLifeTime, &gcInterval)
					Expect(err).Should(Succeed())
					values = append(values, gcLifeTime)
				}
				rows.Close()
				Expect(len(values)).Should(Equal(1))
				gcLifeTimes = append(gcLifeTimes, values[0])
			}
			Expect(gcLifeTimes[0]).Should(Not(Equal(gcLifeTimes[1])))
			Expect(gcLifeTimes[1]).Should(Not(Equal(gcLifeTimes[2])))
			Expect(gcLifeTimes[2]).Should(Not(Equal(gcLifeTimes[0])))

			for _, db := range dbs {
				_, err = db.Exec("create table if not exists test.t (id int primary key, v int);")
				Expect(err).Should(Succeed())

				_, err = db.Exec("insert into test.t values (1, 10), (2, 20), (3, 30);")
				Expect(err).Should(Succeed())
			}

			// 5min later
			time.Sleep(5 * time.Minute)
			fmt.Println("check data visibility after 5min")

			for _, db := range dbs {
				rows := utils.MustQuery(db, "select * from test.t")
				Expect(rows).Should(Equal("1 10\n2 20\n3 30"))
			}

			// 15min later
			time.Sleep(10 * time.Minute)
			fmt.Println("check data visibility after 15min")

			for _, db := range dbs[1:] {
				rows := utils.MustQuery(db, "select * from test.t")
				Expect(rows).Should(Equal("1 10\n2 20\n3 30"))
			}
			_, err = dbs[0].Query("select * from test.t")
			Expect(err).Error()

			// 25min later
			time.Sleep(10 * time.Minute)
			fmt.Println("check data visibility after 25min")

			for _, db := range dbs[0:2] {
				_, err = db.Query("select * from test.t")
				Expect(err).Error()
			}
			rows := utils.MustQuery(dbs[2], "select * from test.t")
			Expect(rows).Should(Equal("1 10\n2 20\n3 30"))

			fmt.Println("test succeed")
		})
	})
})
