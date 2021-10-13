package types

type OrderMap map[uint64]Order

func (m OrderMap) Backup() (orderForms []SubmitOrder) {
	for _, order := range m {
		orderForms = append(orderForms, order.Backup())
	}

	return orderForms
}

func (m OrderMap) Add(o Order) {
	m[o.OrderID] = o
}

// Update only updates the order when the order exists in the map
func (m OrderMap) Update(o Order) {
	if _, ok := m[o.OrderID]; ok {
		m[o.OrderID] = o
	}
}

func (m OrderMap) Remove(orderID uint64) {
	delete(m, orderID)
}

func (m OrderMap) IDs() (ids []uint64) {
	for id := range m {
		ids = append(ids, id)
	}

	return ids
}

func (m OrderMap) Exists(orderID uint64) bool {
	_, ok := m[orderID]
	return ok
}

func (m OrderMap) FindByStatus(status OrderStatus) (orders OrderSlice) {
	for _, o := range m {
		if o.Status == status {
			orders = append(orders, o)
		}
	}

	return orders
}

func (m OrderMap) Filled() OrderSlice {
	return m.FindByStatus(OrderStatusFilled)
}

func (m OrderMap) Canceled() OrderSlice {
	return m.FindByStatus(OrderStatusCanceled)
}

func (m OrderMap) Orders() (orders OrderSlice) {
	for _, o := range m {
		orders = append(orders, o)
	}
	return orders
}

type OrderSlice []Order

func (s OrderSlice) IDs() (ids []uint64) {
	for _, o := range s {
		ids = append(ids, o.OrderID)
	}
	return ids
}