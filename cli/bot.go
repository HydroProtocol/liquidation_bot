package cli

import "auctionBidder/clients"

type BidderBot struct {
	bidderClient    *clients.BidderClient
	ddexClientGroup map[string]*clients.DdexClient // base address - quote address => ddexClient
	blockChannel    chan int64
}

// 给每个auction分配ddex client

func (b *BidderBot) Run() {
	for true {
		<-b.blockChannel
		allAuctions, err := b.bidderClient.GetAllAuctions()
		if err != nil {
			continue
		}
	}
}
