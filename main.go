package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DonnieTD/Exchange/orderbook"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	ex := NewExchange()

	e.GET("book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.Start(":3000")
	fmt.Println("Hello World")
}

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)

	// Create our eth market order book
	orderbooks[MarketETH] = orderbook.NewOrderBook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	Type   OrderType // limit or market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

// convenient struct to dump data out of
type OrderBookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "Market not found",
		})
	}

	orderBookData := OrderBookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := &Order{
				ID:        order.ID,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := &Order{
				ID:        order.ID,
				Price:     order.Limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Bids, o)
		}
	}

	return c.JSON(http.StatusOK, orderBookData)
}

// HOMEWORK ( make this better \TIME TO SHINE\)
func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// unhardcode the market TODO
	ob := ex.orderbooks[MarketETH]
	orderCanceled := false

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			if order.ID == int64(id) {
				ob.CancelOrder(order)
				orderCanceled = true
			}

			if orderCanceled {
				return c.JSON(200, map[string]any{
					"msg": "Order Canceled Succesfully",
				})
			}
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			if order.ID == int64(id) {
				ob.CancelOrder(order)
				orderCanceled = true
			}

			if orderCanceled {
				return c.JSON(200, map[string]any{
					"msg": "Order Canceled Succesfully",
				})
			}
		}
	}

	return nil
}

// Curl test for handlePlaceOrder
// curl --location --request POST 'localhost:3000/order' \
// --header 'Content-Type: application/json' \
//
//	--data-raw '{
//	    "bid":true,
//	    "size": 10,
//	    "price": 10000,
//	    "type" : "LIMIT",
//	    "market": "ETH"
//	}'
func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		fmt.Println("handlePlaceOrderError:", err)
		return err
	}

	// Convert string to market type
	market := Market(placeOrderData.Market)

	// get the appropriate orderbook from the exchange
	ob := ex.orderbooks[market]

	// create the order
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	// place the order
	if placeOrderData.Type == MarketOrder {
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(200, map[string]any{
			"matches": len(matches),
		})
	} else {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(200, map[string]any{
			"msg": "Limit order placed succesfully",
		})
	}

}
