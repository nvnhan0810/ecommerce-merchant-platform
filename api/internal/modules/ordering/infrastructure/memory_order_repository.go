package infrastructure

import (
	"strings"
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type InMemoryOrderRepository struct {
	mu    sync.RWMutex
	items map[domain.OrderID]domain.Order
	order []domain.OrderID
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		items: map[domain.OrderID]domain.Order{},
		order: []domain.OrderID{},
	}
}

func (r *InMemoryOrderRepository) Save(order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[order.ID]; !ok {
		r.order = append(r.order, order.ID)
	}
	stored := order
	stored.ClearPendingEvents()
	stored.ClearPendingDeliveryEvents()
	if strings.TrimSpace(stored.DeliveryCarrier) == "" {
		stored.DeliveryCarrier = domain.DefaultDeliveryCarrier
	}
	r.items[order.ID] = stored
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id domain.OrderID) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.items[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return cloneOrder(o), nil
}

func (r *InMemoryOrderRepository) FindByCode(code string) (domain.Order, error) {
	parsed, err := domain.ParseOrderCode(code)
	if err != nil {
		return domain.Order{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		o := r.items[id]
		if o.Code == parsed {
			return cloneOrder(o), nil
		}
	}
	return domain.Order{}, domain.ErrOrderNotFound
}

func (r *InMemoryOrderRepository) FindByDeliveryTrackingCode(code string) (domain.Order, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		o := r.items[id]
		if o.DeliveryTrackingCode == code {
			return cloneOrder(o), nil
		}
	}
	return domain.Order{}, domain.ErrOrderNotFound
}

func (r *InMemoryOrderRepository) HasDeliveryEventID(eventID string) (bool, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return false, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		o := r.items[id]
		for _, ev := range o.DeliveryEvents {
			if ev.EventID == eventID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *InMemoryOrderRepository) List(limit, offset int) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if offset >= len(r.order) {
		return []domain.Order{}, nil
	}
	end := offset + limit
	if end > len(r.order) {
		end = len(r.order)
	}
	out := make([]domain.Order, 0, end-offset)
	for _, id := range r.order[offset:end] {
		out = append(out, cloneOrder(r.items[id]))
	}
	return out, nil
}

func (r *InMemoryOrderRepository) ListByMerchant(merchantID string, limit, offset int) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Order, 0)
	for _, id := range r.order {
		o := r.items[id]
		if o.MerchantID == merchantID {
			filtered = append(filtered, cloneOrder(o))
		}
	}
	if offset >= len(filtered) {
		return []domain.Order{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *InMemoryOrderRepository) ListByUser(userID string, limit, offset int) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Order, 0)
	for _, id := range r.order {
		o := r.items[id]
		if o.UserID == userID {
			filtered = append(filtered, cloneOrder(o))
		}
	}
	if offset >= len(filtered) {
		return []domain.Order{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *InMemoryOrderRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}

func cloneOrder(o domain.Order) domain.Order {
	cp := o
	cp.Items = append([]domain.OrderItem(nil), o.Items...)
	cp.History = append([]domain.OrderEvent(nil), o.History...)
	cp.DeliveryEvents = append([]domain.DeliveryEvent(nil), o.DeliveryEvents...)
	if o.PaidAt != nil {
		t := o.PaidAt.UTC()
		cp.PaidAt = &t
	}
	if cp.PaymentMethod == "" {
		cp.PaymentMethod = domain.PaymentMethodCOD
	}
	if cp.PaymentStatus == "" {
		cp.PaymentStatus = domain.PaymentStatusUnpaid
	}
	cp.ClearPendingEvents()
	cp.ClearPendingDeliveryEvents()
	return cp
}
