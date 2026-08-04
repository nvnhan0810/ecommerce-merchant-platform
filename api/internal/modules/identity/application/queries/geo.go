package queries

import "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"

type ListCountriesHandler struct {
	repo domain.GeoRepository
}

func NewListCountriesHandler(repo domain.GeoRepository) *ListCountriesHandler {
	return &ListCountriesHandler{repo: repo}
}

func (h *ListCountriesHandler) Handle() ([]domain.Country, error) {
	return h.repo.ListCountries()
}

type ListProvincesHandler struct {
	repo domain.GeoRepository
}

func NewListProvincesHandler(repo domain.GeoRepository) *ListProvincesHandler {
	return &ListProvincesHandler{repo: repo}
}

func (h *ListProvincesHandler) Handle(countryCode string) ([]domain.Province, error) {
	if countryCode == "" {
		countryCode = "VN"
	}
	return h.repo.ListProvinces(countryCode)
}

type ListWardsHandler struct {
	repo domain.GeoRepository
}

func NewListWardsHandler(repo domain.GeoRepository) *ListWardsHandler {
	return &ListWardsHandler{repo: repo}
}

func (h *ListWardsHandler) Handle(provinceCode string) ([]domain.Ward, error) {
	if provinceCode == "" {
		return []domain.Ward{}, nil
	}
	return h.repo.ListWards(provinceCode)
}
