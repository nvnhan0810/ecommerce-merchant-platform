package commands

import (
	"context"
	"fmt"
	"strings"

	catalogdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderCommand struct {
	UserID          string
	Note            string
	ShippingName    string
	ShippingPhone   string
	ShippingAddress string
	Items           []CreateOrderItemInput
	Actor           domain.Actor
}

type CreateOrderHandler struct {
	orders   domain.OrderRepository
	products catalogdomain.ProductRepository
}

func NewCreateOrderHandler(orders domain.OrderRepository, products catalogdomain.ProductRepository) *CreateOrderHandler {
	return &CreateOrderHandler{orders: orders, products: products}
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) ([]queries.OrderDTO, error) {
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		return nil, domain.ErrUserRequired
	}
	if len(cmd.Items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}

	type lineAgg struct {
		product  catalogdomain.Product
		quantity int
	}
	byProduct := map[catalogdomain.ProductID]*lineAgg{}
	for _, item := range cmd.Items {
		pid, err := catalogdomain.ParseProductID(item.ProductID)
		if err != nil {
			return nil, err
		}
		if item.Quantity <= 0 {
			return nil, domain.ErrInvalidOrderQuantity
		}
		if existing, ok := byProduct[pid]; ok {
			existing.quantity += item.Quantity
			continue
		}
		product, err := h.products.FindByID(pid)
		if err != nil {
			return nil, err
		}
		byProduct[pid] = &lineAgg{product: product, quantity: item.Quantity}
	}

	byMerchant := map[string][]domain.OrderLineInput{}
	merchantCurrency := map[string]string{}
	for _, agg := range byProduct {
		if err := agg.product.Reserve(agg.quantity); err != nil {
			return nil, err
		}
		mid := agg.product.MerchantID
		byMerchant[mid] = append(byMerchant[mid], domain.OrderLineInput{
			ProductID:      string(agg.product.ID),
			ProductName:    agg.product.Name,
			MerchantID:     mid,
			UnitPriceCents: agg.product.Price.AmountCents,
			Quantity:       agg.quantity,
		})
		merchantCurrency[mid] = agg.product.Price.Currency
	}

	for _, agg := range byProduct {
		if err := h.products.Save(agg.product); err != nil {
			return nil, fmt.Errorf("save product stock: %w", err)
		}
	}

	out := make([]queries.OrderDTO, 0, len(byMerchant))
	actor := cmd.Actor
	if strings.TrimSpace(actor.Role) == "" {
		actor.Role = "user"
	}
	if strings.TrimSpace(actor.DisplayName) == "" {
		actor.DisplayName = actor.Email
	}

	for merchantID, lines := range byMerchant {
		order, err := domain.NewOrder(userID, merchantID, merchantCurrency[merchantID], cmd.Note, cmd.ShippingName, cmd.ShippingPhone, cmd.ShippingAddress, lines)
		if err != nil {
			return nil, err
		}
		order.RecordCreated(actor)
		if err := h.orders.Save(order); err != nil {
			return nil, err
		}
		if marker, ok := h.products.(interface {
			MarkOrdered(id catalogdomain.ProductID)
		}); ok {
			for _, line := range lines {
				marker.MarkOrdered(catalogdomain.ProductID(line.ProductID))
			}
		}
		saved, err := h.orders.FindByID(order.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, queries.ToDTOWithHistory(saved, true))
	}

	return out, nil
}
