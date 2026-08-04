package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/mediaurl"
)

type AccountDTO struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Role         string `json:"role"`
	AvatarKey    string `json:"avatar_key,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	AddressLine  string `json:"address_line,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
	WardCode     string `json:"ward_code,omitempty"`
	CountryName  string `json:"country_name,omitempty"`
	ProvinceName string `json:"province_name,omitempty"`
	WardName     string `json:"ward_name,omitempty"`
}

// PublicMerchantDTO is safe for storefront responses (no email, no detailed address).
type PublicMerchantDTO struct {
	ID           string `json:"id"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
	WardCode     string `json:"ward_code,omitempty"`
	CountryName  string `json:"country_name,omitempty"`
	ProvinceName string `json:"province_name,omitempty"`
	WardName     string `json:"ward_name,omitempty"`
}

type ListAccountsHandler struct {
	repo       domain.AccountRepository
	role       domain.Role
	publicBase string
	geo        domain.GeoRepository
}

func NewListUsersHandler(repo domain.AccountRepository) *ListAccountsHandler {
	return &ListAccountsHandler{repo: repo, role: domain.RoleUser}
}

func NewListMerchantsHandler(repo domain.AccountRepository, publicBase string, geo domain.GeoRepository) *ListAccountsHandler {
	return &ListAccountsHandler{repo: repo, role: domain.RoleMerchant, publicBase: publicBase, geo: geo}
}

func (h *ListAccountsHandler) Handle(_ context.Context) ([]AccountDTO, error) {
	items, err := h.repo.List()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0, len(items))
	for _, a := range items {
		out = append(out, ToAccountDTO(a, h.role, h.publicBase, h.geo))
	}
	return out, nil
}

func ToAccountDTO(account domain.Account, role domain.Role, publicBase string, geo domain.GeoRepository) AccountDTO {
	dto := AccountDTO{
		ID:           string(account.ID),
		Email:        account.Email,
		DisplayName:  account.DisplayName,
		Role:         string(role),
		AvatarKey:    account.AvatarKey,
		AvatarURL:    mediaurl.Absolute(publicBase, account.AvatarKey),
		AddressLine:  account.AddressLine,
		CountryCode:  account.CountryCode,
		ProvinceCode: account.ProvinceCode,
		WardCode:     account.WardCode,
	}
	enrichGeoNames(&dto.CountryName, &dto.ProvinceName, &dto.WardName, account, geo)
	return dto
}

func ToPublicMerchantDTO(account domain.Account, publicBase string, geo domain.GeoRepository) PublicMerchantDTO {
	dto := PublicMerchantDTO{
		ID:           string(account.ID),
		DisplayName:  account.DisplayName,
		AvatarURL:    mediaurl.Absolute(publicBase, account.AvatarKey),
		CountryCode:  account.CountryCode,
		ProvinceCode: account.ProvinceCode,
		WardCode:     account.WardCode,
	}
	enrichGeoNames(&dto.CountryName, &dto.ProvinceName, &dto.WardName, account, geo)
	return dto
}

func enrichGeoNames(countryName, provinceName, wardName *string, account domain.Account, geo domain.GeoRepository) {
	if geo == nil {
		return
	}
	if account.CountryCode != "" {
		if c, err := geo.GetCountry(account.CountryCode); err == nil {
			*countryName = c.Name
		}
	}
	if account.ProvinceCode != "" {
		if p, err := geo.GetProvince(account.ProvinceCode); err == nil {
			*provinceName = p.Name
		}
	}
	if account.WardCode != "" {
		if w, err := geo.GetWard(account.WardCode); err == nil {
			*wardName = w.Name
		}
	}
}
