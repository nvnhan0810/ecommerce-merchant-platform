package infrastructure

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type InMemoryPaymentRepository struct {
	mu        sync.RWMutex
	items     map[domain.PaymentID]domain.Payment
	byRef     map[string]domain.PaymentID
	callbacks map[domain.PaymentCallbackEventID]domain.PaymentCallbackEvent
	settings  domain.PaymentSettings
}

func NewInMemoryPaymentRepository() *InMemoryPaymentRepository {
	return &InMemoryPaymentRepository{
		items:     map[domain.PaymentID]domain.Payment{},
		byRef:     map[string]domain.PaymentID{},
		callbacks: map[domain.PaymentCallbackEventID]domain.PaymentCallbackEvent{},
		settings: domain.PaymentSettings{
			OnePayDomestic: domain.OnePayGatewaySettings{
				PaymentURL: DefaultOnePayDomesticPaymentURL,
			},
			OnePayInternational: domain.OnePayGatewaySettings{
				PaymentURL: DefaultOnePayInternationalPaymentURL,
			},
			UpdatedAt: time.Now().UTC(),
		},
	}
}

func (r *InMemoryPaymentRepository) Save(payment domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := payment
	cp.OrderIDs = append([]domain.OrderID(nil), payment.OrderIDs...)
	r.items[payment.ID] = cp
	r.byRef[payment.MerchTxnRef] = payment.ID
	return nil
}

func (r *InMemoryPaymentRepository) FindByID(id domain.PaymentID) (domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	cp := p
	cp.OrderIDs = append([]domain.OrderID(nil), p.OrderIDs...)
	return cp, nil
}

func (r *InMemoryPaymentRepository) FindByMerchTxnRef(ref string) (domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byRef[strings.TrimSpace(ref)]
	if !ok {
		return domain.Payment{}, domain.ErrPaymentNotFound
	}
	p := r.items[id]
	cp := p
	cp.OrderIDs = append([]domain.OrderID(nil), p.OrderIDs...)
	return cp, nil
}

func (r *InMemoryPaymentRepository) GetSettings() (domain.PaymentSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.settings, nil
}

func (r *InMemoryPaymentRepository) SaveSettings(settings domain.PaymentSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	settings.UpdatedAt = time.Now().UTC()
	if strings.TrimSpace(settings.OnePayDomestic.PaymentURL) == "" {
		settings.OnePayDomestic.PaymentURL = DefaultOnePayDomesticPaymentURL
	}
	if strings.TrimSpace(settings.OnePayInternational.PaymentURL) == "" {
		settings.OnePayInternational.PaymentURL = DefaultOnePayInternationalPaymentURL
	}
	r.settings = settings
	return nil
}

func (r *InMemoryPaymentRepository) SaveCallbackEvent(event domain.PaymentCallbackEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.callbacks == nil {
		r.callbacks = map[domain.PaymentCallbackEventID]domain.PaymentCallbackEvent{}
	}
	cp := event
	if len(cp.RawPayload) > 0 {
		cp.RawPayload = append(json.RawMessage(nil), event.RawPayload...)
	}
	r.callbacks[event.ID] = cp
	return nil
}

func (r *InMemoryPaymentRepository) FindCallbackEventByID(id domain.PaymentCallbackEventID) (domain.PaymentCallbackEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ev, ok := r.callbacks[id]
	if !ok {
		return domain.PaymentCallbackEvent{}, domain.ErrPaymentCallbackNotFound
	}
	cp := ev
	if len(cp.RawPayload) > 0 {
		cp.RawPayload = append(json.RawMessage(nil), ev.RawPayload...)
	}
	return cp, nil
}

func (r *InMemoryPaymentRepository) ListCallbackEvents(filter domain.PaymentCallbackListFilter) ([]domain.PaymentCallbackEvent, error) {
	filter.Normalize()
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.PaymentCallbackEvent, 0, len(r.callbacks))
	for _, ev := range r.callbacks {
		if filter.Provider != "" && ev.Provider != filter.Provider {
			continue
		}
		if filter.Channel != "" && string(ev.Channel) != filter.Channel {
			continue
		}
		if filter.MerchTxnRef != "" && !strings.Contains(strings.ToLower(ev.MerchTxnRef), strings.ToLower(filter.MerchTxnRef)) {
			continue
		}
		cp := ev
		if len(cp.RawPayload) > 0 {
			cp.RawPayload = append(json.RawMessage(nil), ev.RawPayload...)
		}
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if filter.Offset >= len(out) {
		return []domain.PaymentCallbackEvent{}, nil
	}
	end := filter.Offset + filter.Limit
	if end > len(out) {
		end = len(out)
	}
	return out[filter.Offset:end], nil
}
