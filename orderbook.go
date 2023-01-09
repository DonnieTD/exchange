package main

import (
	"fmt"
	"sort"
	"time"

	"golang.org/x/exp/slices"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFileed float64
	Price      float64
}

type Order struct {
	Size  float64
	Bid   bool
	Limit *Limit
	// Unix nano
	Timestamp int64
}

type Orders []*Order

func (o Orders) Len() int               { return len(o) }
func (o Orders) Swap(i int, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i int, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("\n{size: %.2f}\n", o.Size)
}

// A limit is a group of orders
// at a certain price level
type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

type Limits []*Limit

type ByBestAsk struct{ Limits }

// Asks Sort related fn's
func (a *ByBestAsk) Len() int               { return len(a.Limits) }
func (a *ByBestAsk) Swap(i int, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a *ByBestAsk) Less(i int, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

// Bids Sort related fn's
func (b *ByBestBid) Len() int          { return len(b.Limits) }
func (b *ByBestBid) Swap(i int, j int) { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }

// This is reversed because if we want to buy we want to buy as cheaply as possible,
// that means if we say less it should be higher ( might be wrongs xd if wrong we will come swap)
func (b *ByBestBid) Less(i int, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("{Identifier: Limit, TotalVolume: %.2f, Price: %.2f}", l.TotalVolume, l.Price)
}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func RemoveOrderAtIndex(s []*Order, index int) []*Order {
	// we decided to go with reslicing until furthre notice
	return append(s[:index], s[index+1:]...)
}

func FastDelete(s []*Order, i int) []*Order {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (l *Limit) DeleteOrder(o *Order) {

	idx := slices.IndexFunc(l.Orders, func(c *Order) bool { return c == o })
	// delete via reslicing to preserve order and not decend into loop madness
	l.Orders = FastDelete(l.Orders, idx)
	o.Limit = nil
	l.TotalVolume -= o.Size

	sort.Sort(l.Orders)
}

type OrderBook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (ob *OrderBook) PlaceOrder(price float64, o *Order) []Match {
	// Try to match the order
	// Add the rest of the orders to the books if the intial match doesnt fill the entire order
	if o.Size > 0.0 {
		ob.add(price, o)
	}
	return []Match{}
}

func (ob *OrderBook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		limit.AddOrder(o)
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	}

}
