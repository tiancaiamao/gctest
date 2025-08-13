// provide a simple command to run UpdateServiceGCSafePoint API
// This is used in compatibility test

// !!!NOTE!!! Do not upgrade go.mod file in this directory, we need to use old client-go

package main

import (
	"os"
	"math"
	"fmt"
	"context"
	"strconv"

	"github.com/tikv/pd/client/pkg/caller"
	"github.com/tikv/pd/client"
)

func main() {
	serviceName := os.Args[1]
	barrierTS, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wrong parameter")
		os.Exit(-1)
	}
	
	pdcli, err := pd.NewClient(caller.Component("test"),
		[]string{"127.0.0.1:2379"}, pd.SecurityOption{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open pd client fail??", err)
		fmt.Println("what the fuck??111")
		os.Exit(-1)
		return
	}
	defer pdcli.Close()

	fmt.Println("before api call?")
	txnSafePoint, err := pdcli.UpdateServiceGCSafePoint(context.Background(), serviceName, math.MaxInt64, barrierTS)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("what the fuck??")
		os.Exit(-1)
		return
	}
	fmt.Println("txn safe point == ", txnSafePoint)
	return
}
