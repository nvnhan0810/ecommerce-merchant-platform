package infrastructure

import (
	"fmt"

	catalogdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type demoOrderSpec struct {
	userEmail     string
	merchantEmail string
	status        domain.OrderStatus
	note          string
	qty           int
}

func demoOrderSpecs() []demoOrderSpec {
	return []demoOrderSpec{
		{userEmail: "buyer@ecomerce.local", merchantEmail: "shop@ecomerce.local", status: domain.StatusNew, note: "Demo đơn mới", qty: 1},
		{userEmail: "an@ecomerce.local", merchantEmail: "fashion@ecomerce.local", status: domain.StatusPaid, note: "Demo đã thanh toán", qty: 1},
		{userEmail: "binh@ecomerce.local", merchantEmail: "tech@ecomerce.local", status: domain.StatusConfirmed, note: "Demo đã xác nhận", qty: 2},
		{userEmail: "chi@ecomerce.local", merchantEmail: "home@ecomerce.local", status: domain.StatusShipping, note: "Demo đang giao", qty: 1},
		{userEmail: "buyer@ecomerce.local", merchantEmail: "shop@ecomerce.local", status: domain.StatusSucceeded, note: "Demo thành công", qty: 1},
		{userEmail: "an@ecomerce.local", merchantEmail: "tech@ecomerce.local", status: domain.StatusFailed, note: "Demo thất bại", qty: 1},
		{userEmail: "binh@ecomerce.local", merchantEmail: "fashion@ecomerce.local", status: domain.StatusCancelled, note: "Demo huỷ", qty: 1},
	}
}

// SeedDemoOrders inserts sample orders (idempotent when orders table already has rows).
func SeedDemoOrders(
	orders domain.OrderRepository,
	users identitydomain.AccountRepository,
	merchants identitydomain.AccountRepository,
	products catalogdomain.ProductRepository,
) error {
	n, err := orders.Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	userByEmail, err := accountByEmail(users)
	if err != nil {
		return err
	}
	merchantByEmail, err := accountByEmail(merchants)
	if err != nil {
		return err
	}

	productList, err := products.List(500, 0)
	if err != nil {
		return err
	}
	productsByMerchant := map[string][]catalogdomain.Product{}
	for _, p := range productList {
		productsByMerchant[p.MerchantID] = append(productsByMerchant[p.MerchantID], p)
	}

	for _, spec := range demoOrderSpecs() {
		user, ok := userByEmail[spec.userEmail]
		if !ok {
			return fmt.Errorf("seed order: user %q missing", spec.userEmail)
		}
		merchant, ok := merchantByEmail[spec.merchantEmail]
		if !ok {
			return fmt.Errorf("seed order: merchant %q missing", spec.merchantEmail)
		}
		merchantProducts := productsByMerchant[string(merchant.ID)]
		if len(merchantProducts) == 0 {
			return fmt.Errorf("seed order: merchant %q has no products", spec.merchantEmail)
		}
		p := merchantProducts[0]
		order, err := domain.NewOrder(string(user.ID), string(merchant.ID), p.Price.Currency, spec.note, []domain.OrderLineInput{{
			ProductID:      string(p.ID),
			ProductName:    p.Name,
			MerchantID:     p.MerchantID,
			UnitPriceCents: p.Price.AmountCents,
			Quantity:       spec.qty,
		}})
		if err != nil {
			return err
		}
		if err := order.ChangeStatus(spec.status); err != nil {
			return err
		}
		if err := orders.Save(order); err != nil {
			return err
		}
	}
	return nil
}

func accountByEmail(repo identitydomain.AccountRepository) (map[string]identitydomain.Account, error) {
	list, err := repo.List()
	if err != nil {
		return nil, err
	}
	out := make(map[string]identitydomain.Account, len(list))
	for _, a := range list {
		out[a.Email] = a
	}
	return out, nil
}
