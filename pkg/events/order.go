package events

const OrderCreatedTopic = "orders.created"

type OrderCreated struct {
	EventType   string `json:"event_type"`
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

func NewOrderCreated(orderID, userID, productName string, quantity int) OrderCreated {
	return OrderCreated{
		EventType:   "order.created",
		OrderID:     orderID,
		UserID:      userID,
		ProductName: productName,
		Quantity:    quantity,
	}
}
