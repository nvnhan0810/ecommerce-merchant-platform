package domain

import "errors"

var (
	ErrCountryNotFound  = errors.New("country not found")
	ErrProvinceNotFound = errors.New("province not found")
	ErrWardNotFound     = errors.New("ward not found")
	ErrInvalidGeoRef    = errors.New("invalid country, province, or ward")
)

type Country struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	NameEn    string `json:"name_en"`
	IsDefault bool   `json:"is_default"`
}

type Province struct {
	Code        string   `json:"code"`
	CountryCode string   `json:"country_code"`
	Name        string   `json:"name"`
	NameEn      string   `json:"name_en"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

type Ward struct {
	Code         string   `json:"code"`
	ProvinceCode string   `json:"province_code"`
	Name         string   `json:"name"`
	NameEn       string   `json:"name_en"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

type GeoRepository interface {
	ListCountries() ([]Country, error)
	DefaultCountry() (Country, error)
	GetCountry(code string) (Country, error)
	ListProvinces(countryCode string) ([]Province, error)
	ListWards(provinceCode string) ([]Ward, error)
	GetProvince(code string) (Province, error)
	GetWard(code string) (Ward, error)
}
