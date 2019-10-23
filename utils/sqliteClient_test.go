package utils

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"testing"
)

func TestInitDb(t *testing.T) {
	os.Setenv("SQLITEPATH", "/Users/leimingda/Documents/ddex/auctionBidder/workingDir/auctionBidderSqlite")
	InitDb()
	spew.Dump(QueryPosition())
}
