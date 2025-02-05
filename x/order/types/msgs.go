package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"strconv"
)

// nolint
const (
	OrderItemLimit            = 200
	MultiCancelOrderItemLimit = 200
)

// nolint
type MsgNewOrder struct {
	Sender   sdk.AccAddress `json:"sender"`   // order maker address
	Product  string         `json:"product"`  // product for trading pair in full name of the tokens
	Side     string         `json:"side"`     // BUY/SELL
	Price    sdk.Dec        `json:"price"`    // price of the order
	Quantity sdk.Dec        `json:"quantity"` // quantity of the order
}

// NewMsgNewOrder is a constructor function for MsgNewOrder
func NewMsgNewOrder(sender sdk.AccAddress, product string, side string, price string,
	quantity string) MsgNewOrders {

	return MsgNewOrders{
		Sender: sender,
		OrderItems: []OrderItem{
			{
				Product:  product,
				Side:     side,
				Price:    sdk.MustNewDecFromStr(price),
				Quantity: sdk.MustNewDecFromStr(quantity),
			},
		},
	}
}

// nolint
type MsgCancelOrder struct {
	Sender  sdk.AccAddress `json:"sender"`
	OrderID string         `json:"order_id"`
}

// NewMsgCancelOrder is a constructor function for MsgCancelOrder
func NewMsgCancelOrder(sender sdk.AccAddress, orderID string) MsgCancelOrders {
	msgCancelOrder := MsgCancelOrders{
		Sender:   sender,
		OrderIDs: []string{orderID},
	}
	return msgCancelOrder
}

//********************MsgNewOrders*************
// nolint
type MsgNewOrders struct {
	Sender     sdk.AccAddress `json:"sender"` // order maker address
	OrderItems []OrderItem    `json:"order_items"`
}

// nolint
type OrderItem struct {
	Product  string  `json:"product"`  // product for trading pair in full name of the tokens
	Side     string  `json:"side"`     // BUY/SELL
	Price    sdk.Dec `json:"price"`    // price of the order
	Quantity sdk.Dec `json:"quantity"` // quantity of the order
}

// nolint
func NewOrderItem(product string, side string, price string,
	quantity string) OrderItem {
	return OrderItem{
		Product:  product,
		Side:     side,
		Price:    sdk.MustNewDecFromStr(price),
		Quantity: sdk.MustNewDecFromStr(quantity),
	}
}

// NewMsgNewOrders is a constructor function for MsgNewOrder
func NewMsgNewOrders(sender sdk.AccAddress, orderItems []OrderItem) MsgNewOrders {
	return MsgNewOrders{
		Sender:     sender,
		OrderItems: orderItems,
	}
}

// nolint
func (msg MsgNewOrders) Route() string { return "order" }

// nolint
func (msg MsgNewOrders) Type() string { return "new" }

// ValidateBasic : Implements Msg.
func (msg MsgNewOrders) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if msg.OrderItems == nil || len(msg.OrderItems) == 0 {
		return sdk.ErrUnknownRequest("invalid OrderItems")
	}
	if len(msg.OrderItems) > OrderItemLimit {
		return sdk.ErrUnknownRequest("Numbers of NewOrderItem should not be more than " + strconv.Itoa(OrderItemLimit))
	}
	for _, item := range msg.OrderItems {
		if len(item.Product) == 0 {
			return sdk.ErrUnknownRequest("Product cannot be empty")
		}
		symbols := strings.Split(item.Product, "_")
		if len(symbols) != 2 {
			return sdk.ErrUnknownRequest("Product should be in the format of \"base_quote\"")
		}
		if symbols[0] == symbols[1] {
			return sdk.ErrUnknownRequest("invalid product")
		}
		if item.Side != BuyOrder && item.Side != SellOrder {
			return sdk.ErrUnknownRequest(
				fmt.Sprintf("Side is expected to be \"BUY\" or \"SELL\", but got \"%s\"", item.Side))
		}
		if !(item.Price.IsPositive() && item.Quantity.IsPositive()) {
			return sdk.ErrUnknownRequest("Price/Quantity must be positive")
		}
	}

	return nil
}

// GetSignBytes : encodes the message for signing
func (msg MsgNewOrders) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners defines whose signature is required
func (msg MsgNewOrders) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// Calculate customize gas
func (msg MsgNewOrders) CalculateGas(gasUnit uint64) uint64 {
	return uint64(len(msg.OrderItems)) * gasUnit
}

// nolint
type MsgCancelOrders struct {
	Sender   sdk.AccAddress `json:"sender"` // order maker address
	OrderIDs []string       `json:"order_ids"`
}

// NewMsgCancelOrders is a constructor function for MsgCancelOrder
func NewMsgCancelOrders(sender sdk.AccAddress, orderIDItems []string) MsgCancelOrders {
	msgCancelOrder := MsgCancelOrders{
		Sender:   sender,
		OrderIDs: orderIDItems,
	}
	return msgCancelOrder
}

// nolint
func (msg MsgCancelOrders) Route() string { return "order" }

// nolint
func (msg MsgCancelOrders) Type() string { return "cancel" }

// nolint
func (msg MsgCancelOrders) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if msg.OrderIDs == nil || len(msg.OrderIDs) == 0 {
		return sdk.ErrUnknownRequest("invalid OrderIDs")
	}
	if len(msg.OrderIDs) > MultiCancelOrderItemLimit {
		return sdk.ErrUnknownRequest("Numbers of CancelOrderItem should not be more than " + strconv.Itoa(OrderItemLimit))
	}
	if hasDuplicatedID(msg.OrderIDs) {
		return sdk.ErrUnknownRequest("Duplicated order ids detected")
	}
	for _, item := range msg.OrderIDs {
		if item == "" {
			return sdk.ErrUnauthorized("orderID cannot be empty")
		}
	}

	return nil
}

func hasDuplicatedID(ids []string) bool {
	idSet := make(map[string]bool)
	for _, item := range ids {
		if !idSet[item] {
			idSet[item] = true
		} else {
			return true
		}
	}
	return false
}

// GetSignBytes encodes the message for signing
func (msg MsgCancelOrders) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners defines whose signature is required
func (msg MsgCancelOrders) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// Calculate customize gas
func (msg MsgCancelOrders) CalculateGas(gasUnit uint64) uint64 {
	return uint64(len(msg.OrderIDs)) * gasUnit
}

// nolint
type OrderResult struct {
	Error   error  `json:"error"`
	Message string `json:"msg"`     // order return error message
	OrderID string `json:"orderid"` // order return orderid
}
