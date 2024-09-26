package trade

import (
	"github.com/nthskyradiated/openpricegap-analysis/internal/news"
	"github.com/nthskyradiated/openpricegap-analysis/internal/pos"
)

type Selection struct {
	Ticker string
	pos.Position
	Articles []news.Article
}

type Deliverer interface {
	Deliver(selections []Selection) error
}