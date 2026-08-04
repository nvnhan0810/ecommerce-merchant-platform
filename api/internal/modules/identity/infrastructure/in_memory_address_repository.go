package infrastructure

import (
	"sort"
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type InMemoryAddressRepository struct {
	mu   sync.RWMutex
	data map[domain.AddressID]domain.UserAddress
}

func NewInMemoryAddressRepository() *InMemoryAddressRepository {
	return &InMemoryAddressRepository{
		data: make(map[domain.AddressID]domain.UserAddress),
	}
}

func (r *InMemoryAddressRepository) Save(a domain.UserAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[a.ID] = a
	return nil
}

func (r *InMemoryAddressRepository) FindByID(id domain.AddressID) (domain.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.data[id]
	if !ok {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}
	return a, nil
}

func (r *InMemoryAddressRepository) ListByUserID(userID domain.AccountID) ([]domain.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.UserAddress
	for _, a := range r.data {
		if a.UserID == userID {
			items = append(items, a)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDefault != items[j].IsDefault {
			return items[i].IsDefault
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (r *InMemoryAddressRepository) Delete(id domain.AddressID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *InMemoryAddressRepository) ClearDefault(userID domain.AccountID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, a := range r.data {
		if a.UserID == userID && a.IsDefault {
			a.IsDefault = false
			r.data[id] = a
		}
	}
	return nil
}

type InMemoryGeoRepository struct {
	mu         sync.RWMutex
	countries  []domain.Country
	provinces  map[string]domain.Province
	wards      map[string]domain.Ward
}

func NewInMemoryGeoRepository() *InMemoryGeoRepository {
	lat, lon := 21.04, 105.836
	return &InMemoryGeoRepository{
		countries: []domain.Country{{Code: "VN", Name: "Việt Nam", NameEn: "Vietnam", IsDefault: true}},
		provinces: map[string]domain.Province{
			"01": {Code: "01", CountryCode: "VN", Name: "Hà Nội", NameEn: "Hanoi", Latitude: &lat, Longitude: &lon},
		},
		wards: map[string]domain.Ward{
			"00004": {Code: "00004", ProvinceCode: "01", Name: "Ba Đình", NameEn: "Ba Dinh", Latitude: &lat, Longitude: &lon},
		},
	}
}

func (r *InMemoryGeoRepository) ListCountries() ([]domain.Country, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Country, len(r.countries))
	copy(out, r.countries)
	return out, nil
}

func (r *InMemoryGeoRepository) DefaultCountry() (domain.Country, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.countries {
		if c.IsDefault {
			return c, nil
		}
	}
	return domain.Country{}, domain.ErrCountryNotFound
}

func (r *InMemoryGeoRepository) GetCountry(code string) (domain.Country, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.countries {
		if c.Code == code {
			return c, nil
		}
	}
	return domain.Country{}, domain.ErrCountryNotFound
}

func (r *InMemoryGeoRepository) ListProvinces(countryCode string) ([]domain.Province, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.Province
	for _, p := range r.provinces {
		if p.CountryCode == countryCode {
			items = append(items, p)
		}
	}
	return items, nil
}

func (r *InMemoryGeoRepository) ListWards(provinceCode string) ([]domain.Ward, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.Ward
	for _, w := range r.wards {
		if w.ProvinceCode == provinceCode {
			items = append(items, w)
		}
	}
	return items, nil
}

func (r *InMemoryGeoRepository) GetProvince(code string) (domain.Province, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.provinces[code]
	if !ok {
		return domain.Province{}, domain.ErrProvinceNotFound
	}
	return p, nil
}

func (r *InMemoryGeoRepository) GetWard(code string) (domain.Ward, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.wards[code]
	if !ok {
		return domain.Ward{}, domain.ErrWardNotFound
	}
	return w, nil
}
