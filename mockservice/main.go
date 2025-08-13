// Copyright 2025 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"strings"
	"testing"
	"time"
	"fmt"

	"github.com/tikv/client-go/v2/oracle"
	pd "github.com/tikv/pd/client"
	"github.com/tikv/pd/client/pkg/caller"
	"github.com/stretchr/testify/require"
)

func main() {
	t := &testing.T{}
	pdcli, err := pd.NewClient(caller.Component("test"),
		[]string{"127.0.0.1:2379"}, pd.SecurityOption{})
	if err != nil {
		fmt.Println("open pd client fail??", err)
		return
	}
	defer pdcli.Close()

	const keyspaceID = 1
	gccli := pdcli.GetGCStatesClient(uint32(keyspaceID))
	state, err := gccli.GetGCState(context.Background())
	require.NoError(t, err)
	require.Empty(t, state.GCBarriers)

	nowTS := oracle.GoTimeToTS(time.Now())
	ttl := 15 * time.Minute
	const barrierID = "demo-service"
	info, err := gccli.SetGCBarrier(context.Background(), barrierID, nowTS, ttl)
	require.NoError(t, err)
	require.Equal(t, barrierID, info.BarrierID)
	require.Equal(t, nowTS, info.BarrierTS)
	require.Equal(t, ttl, info.TTL)

	state, err = gccli.GetGCState(context.Background())
	require.NoError(t, err)
	require.Len(t, state.GCBarriers, 1)
	info = state.GCBarriers[0]
	require.Equal(t, barrierID, info.BarrierID)
	require.Equal(t, nowTS, info.BarrierTS)
	require.Equal(t, ttl, info.TTL)

	cli := pdcli.GetGCInternalController(uint32(keyspaceID))
	res, err := cli.AdvanceTxnSafePoint(context.Background(), nowTS+1)
	require.NoError(t, err)
	require.Equal(t, res.OldTxnSafePoint, state.TxnSafePoint)
	require.Equal(t, res.NewTxnSafePoint, nowTS)
	require.Equal(t, res.Target, nowTS+1)
	require.True(t, strings.Contains(res.BlockerDescription, "GCBarrier"))

	_, err = gccli.DeleteGCBarrier(context.Background(), barrierID)
	require.NoError(t, err)

	fmt.Println("test success")
}
