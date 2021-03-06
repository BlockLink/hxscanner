package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/blocklink/hxscanner/src/config"
	"github.com/blocklink/hxscanner/src/db"
	"github.com/blocklink/hxscanner/src/nodeservice"
	"github.com/blocklink/hxscanner/src/scanner"
	"github.com/blocklink/hxscanner/src/plugins"
	"github.com/blocklink/hxscanner/src/log"
)

func main() {
	logger := log.GetLogger()
	log.InitLogger(logger, "info")
	logger.Println("starting hxscanner")
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())

	nodeApiUrl := flag.String("node_endpoint", "ws://127.0.0.1:8090", "hx_node websocket rpc endpoint(=ws://127.0.0.1:8090)")
	callerPubKey := flag.String("caller_pubkey", "HX5jfbqSFHm1XVUEg93NCym67z28WHmeUi3hqnem3o6Ad1BYsZA9", "contract default caller pubkey(=HX5jfbqSFHm1XVUEg93NCym67z28WHmeUi3hqnem3o6Ad1BYsZA9)")
	dbHost := flag.String("db_host", "127.0.0.1", "postgresql database host(=127.0.0.1)")
	dbPort := flag.Int("db_port", 5432, "postgresql database port(=5432)")
	dbSslMode := flag.String("db_ssl", "disable", "postgresql connection ssl mode(=disable)")
	dbUser := flag.String("db_user", "postgres", "postgresql database username(=postgres)")
	dbPassword := flag.String("db_pass", "", "postgresql database password")
	dbName := flag.String("db_name", "hxscanner", "postgresql database for this application(=hxscanner)")
	scanFromBlockNumberFlag := flag.Int("scan_from", -1, "scan from block number(default last scanned)")
	flag.Parse()

	config.SystemConfig = new(config.Config)
	config.SystemConfig.NodeApiUrl = *nodeApiUrl
	config.SystemConfig.CallerPubKeyString = *callerPubKey
	config.SystemConfig.DbConnectionString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s port=%d", *dbUser, *dbPassword, *dbName, *dbSslMode, *dbHost, *dbPort)

	nodeservice.ConnectHxNode(ctx, config.SystemConfig.NodeApiUrl)
	defer nodeservice.CloseHxNodeConn()
	err := db.OpenDb(config.SystemConfig.DbConnectionString)
	if err != nil {
		logger.Fatal("open db connection error " + err.Error())
		return
	}
	defer db.CloseDb()

	scanner.AddScanPlugin(new(plugins.TransferPlugin))
	scanner.AddScanPlugin(new(plugins.AccountRegisterPlugin))
	scanner.AddScanPlugin(new(plugins.AssetMaybeChangePlugin))
	scanner.AddScanPlugin(new(plugins.TokenContractCreateScanPlugin))
	scanner.AddScanPlugin(new(plugins.TokenContractInvokeScanPlugin))

	go func() {
		lastScannedBlockNum, err := db.GetLastScannedBlockNumber()
		if err != nil {
			logger.Fatal("read last scanned block number error " + err.Error())
			return
		}
		if *scanFromBlockNumberFlag >= 0 {
			lastScannedBlockNum = uint32(*scanFromBlockNumberFlag)
		}
		scanner.ScanBlocksFrom(ctx, int(lastScannedBlockNum)+1)
		signal.Stop(stop)
	}()

	select {
	case <-stop:
		{
			logger.Println("hxscanner stopping")
			cancel()
		}
	}
}
