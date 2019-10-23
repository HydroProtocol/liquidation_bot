package utils

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"os"
)

func InitDb() (err error) {
	dbPath := os.Getenv("SQLITEPATH")
	if _, err = os.Stat(dbPath); os.IsNotExist(err) {
		var db *sql.DB
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		sqlStmt := `
	create table auctions (
	txHash TEXT not null primary key,
	auctionId INTEGER not null,
	debtSymbol TEXT not null,
	collateralSymbol TEXT not null,
	repayDebt TEXT not null,
	receiveCollateral TEXT not null,
	ddexOrderId TEXT not null,
	ddexSellCollateral TEXT not null,
	ddexReceiveDebt TEXT not null,
	gasCost TEXT not null
	);`
		_, err = db.Exec(sqlStmt)
		if err == nil {
			logrus.Infof("create sqlite table auctions")
		}
	}

	return
}

func InsertAuctionRes(
	txHash string,
	auctionId int,
	debtSymbol string,
	collateralSymbol string,
	repayDebt string,
	receiveCollateral string,
	ddexOrderId string,
	ddexSellCollateral string,
	ddexReceiveDebt string,
	gasCost string) (err error) {
	db, err := sql.Open("sqlite3", os.Getenv("SQLITEPATH"))
	defer db.Close()
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Commit()

	stmt, err := tx.Prepare("insert into auctions(txHash, auctionId, debtSymbol, collateralSymbol, repayDebt, receiveCollateral, ddexOrderId, ddexSellCollateral, ddexReceiveDebt, gasCost) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(txHash, auctionId, debtSymbol, collateralSymbol, repayDebt, receiveCollateral, ddexOrderId, ddexSellCollateral, ddexReceiveDebt, gasCost)

	return err
}

func InsertFailedBid(
	txHash string,
	auctionId int,
	debtSymbol string,
	collateralSymbol string,
	gasCost string,
) (err error) {
	return InsertAuctionRes(
		txHash,
		auctionId,
		debtSymbol,
		collateralSymbol,
		"0",
		"0",
		"0x0",
		"0",
		"0",
		gasCost)
}

// token symbol -> position
func QueryPosition() (position map[string]decimal.Decimal, err error) {
	position = map[string]decimal.Decimal{"ETH": decimal.Zero}
	db, err := sql.Open("sqlite3", os.Getenv("SQLITEPATH"))
	defer db.Close()
	if err != nil {
		return
	}
	rows, err := db.Query("select debtSymbol, collateralSymbol, repayDebt, receiveCollateral, ddexSellCollateral, ddexReceiveDebt, gasCost from auctions")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var debtSymbol string
		var collateralSymbol string
		var repayDebt string
		var receiveCollateral string
		var ddexSellCollateral string
		var ddexReceiveDebt string
		var gasCost string
		err = rows.Scan(&debtSymbol, &collateralSymbol, &repayDebt, &receiveCollateral, &ddexSellCollateral, &ddexReceiveDebt, &gasCost)
		if err != nil {
			continue
		}
		if _, ok := position[debtSymbol]; !ok {
			position[debtSymbol] = decimal.Zero
		}
		if _, ok := position[collateralSymbol]; !ok {
			position[collateralSymbol] = decimal.Zero
		}
		position[collateralSymbol] = position[collateralSymbol].Add(String2Decimal(receiveCollateral))
		position[collateralSymbol] = position[collateralSymbol].Sub(String2Decimal(ddexSellCollateral))
		position[debtSymbol] = position[debtSymbol].Sub(String2Decimal(repayDebt))
		position[debtSymbol] = position[debtSymbol].Add(String2Decimal(ddexReceiveDebt))
		position["ETH"] = position["ETH"].Sub(String2Decimal(gasCost))
	}
	return
}
