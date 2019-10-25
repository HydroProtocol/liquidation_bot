package cli

import (
	"auctionBidder/client"
	"auctionBidder/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"time"
)

var DefaultGui *gocui.Gui
var YellowStr = color.New(color.FgYellow).SprintFunc()
var RedStr = color.New(color.FgRed).SprintFunc()
var GreenStr = color.New(color.FgGreen).SprintFunc()
var BlueStr = color.New(color.FgCyan).SprintFunc()

var pnlStr = func(d decimal.Decimal, unit string) string {
	if d.IsPositive() {
		return color.New(color.FgGreen).Sprintf("%s%s", utils.SetPrecision(d, 5).String(), unit)
	}
	if d.IsNegative() {
		return color.New(color.FgRed).Sprintf("%s%s", utils.SetPrecision(d, 5).String(), unit)
	}
	return fmt.Sprintf("%s%s", d.String(), unit)
}

func StartGui() (err error) {
	DefaultGui, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer DefaultGui.Close()

	maxX, maxY := DefaultGui.Size()
	infoView, _ := DefaultGui.SetView("info", 0, maxY/3+1, maxX-1, maxY-1)
	auctionView, _ := DefaultGui.SetView("auction", 0, 0, maxX*2/3-1, maxY/3)
	pnlView, _ := DefaultGui.SetView("pnl", maxX/2, 0, maxX*3/4-1, maxY/3)
	inventoryView, _ := DefaultGui.SetView("inventory", maxX*3/4, 0, maxX-1, maxY/3)
	infoView.Title = "INFO"
	infoView.Autoscroll = true
	infoView.Wrap = true
	auctionView.Title = "CURRENT AUCTIONS"
	auctionView.Autoscroll = false
	auctionView.Wrap = true
	pnlView.Title = "PROFFIT & LOSS"
	pnlView.Autoscroll = false
	pnlView.Wrap = true
	inventoryView.Title = "FREE BALANCE"
	inventoryView.Autoscroll = false
	inventoryView.Wrap = true

	RegisterLogrusHooks()

	if err := DefaultGui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		DefaultGui.Close()
		logrus.Panic(err)
	}

	if err := DefaultGui.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Panic(err)
	}

	return
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func UpdateAuctionView(auctions []*client.Auction) {
	v, err := DefaultGui.View("auction")
	if err != nil {
		logrus.Error(err)
		return
	}
	v.Clear()

	if len(auctions) == 0 {
		DefaultGui.Update(func(g *gocui.Gui) error {
			fmt.Fprintln(v, YellowStr("No Auctions"))
			return nil
		})
		return
	}

	for _, auction := range auctions {
		fmt.Fprintln(v, fmt.Sprintf("%s: sell %s %s for %s with price %s",
			YellowStr("Auction #"+strconv.Itoa(int(auction.ID))),
			auction.AvailableCollateral.String(),
			auction.CollateralSymbol,
			auction.DebtSymbol,
			auction.Price.String(),
		))
	}
}

func UpdatePnlView(position map[string]decimal.Decimal, price map[string]decimal.Decimal) {
	DefaultGui.Update(func(g *gocui.Gui) error {
		v, _ := g.View("pnl")
		v.Clear()
		totalValue := decimal.Zero
		for symbol, _ := range position {
			value := position[symbol].Mul(price[symbol])
			totalValue = totalValue.Add(value)
			fmt.Fprintln(v, fmt.Sprintf("%s: %s (%s)",
				symbol,
				pnlStr(position[symbol], symbol),
				pnlStr(value, "$")))
		}
		fmt.Fprintln(v, "total:"+pnlStr(totalValue, "$"))
		return nil
	})
}

func UpdateInventoryView(ddexClient *client.DdexClient) {
	inventory, err := ddexClient.GetInventory()
	if err != nil {
		return
	}
	DefaultGui.Update(func(g *gocui.Gui) error {
		v, _ := g.View("inventory")
		v.Clear()
		fmt.Fprintln(v, fmt.Sprintf("Your address:%s", ddexClient.Address))
		symbolList := []string{}
		for symbol, _ := range inventory {
			symbolList = append(symbolList, symbol)
		}
		sort.Strings(symbolList)
		for _, symbol := range symbolList {
			fmt.Fprintln(v, fmt.Sprintf("[%s] %s", symbol, inventory[symbol].Free.StringFixed(3)))
		}
		return nil
	})
}

type viewHook struct{}

func (hook *viewHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *viewHook) Fire(entry *logrus.Entry) error {
	DefaultGui.Update(func(g *gocui.Gui) error {
		v, _ := g.View("info")
		var prefix string

		if entry.Level == logrus.InfoLevel {
			prefix = GreenStr("[info]")
		}
		if entry.Level == logrus.DebugLevel {
			prefix = BlueStr("[debug]")
		}
		if entry.Level == logrus.ErrorLevel {
			prefix = RedStr("[error]")
		}
		if entry.Level == logrus.WarnLevel {
			prefix = YellowStr("[warn]")
		}

		fmt.Fprintln(v, fmt.Sprintf("%s %s %s", prefix, time.Now().Format("15:04:05"), entry.Message))
		return nil
	})
	return nil
}

func RegisterLogrusHooks() {
	logrus.AddHook(&viewHook{})
	rl, err := rotatelogs.New(path.Join(os.Getenv("LOGPATH"), "log.%Y%m%d%H"))
	if err != nil {
		panic("log path not exist")
	}
	logrus.SetOutput(rl)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}
