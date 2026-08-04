package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GeoRepository struct {
	pool *pgxpool.Pool
}

func NewGeoRepository(pool *pgxpool.Pool) *GeoRepository {
	return &GeoRepository{pool: pool}
}

func (r *GeoRepository) ListCountries() ([]domain.Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT code, name, name_en, is_default
		FROM countries
		ORDER BY is_default DESC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Country
	for rows.Next() {
		var c domain.Country
		if err := rows.Scan(&c.Code, &c.Name, &c.NameEn, &c.IsDefault); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *GeoRepository) DefaultCountry() (domain.Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var c domain.Country
	err := r.pool.QueryRow(ctx, `
		SELECT code, name, name_en, is_default
		FROM countries
		WHERE is_default = true
		ORDER BY code
		LIMIT 1
	`).Scan(&c.Code, &c.Name, &c.NameEn, &c.IsDefault)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Country{}, domain.ErrCountryNotFound
	}
	return c, err
}

func (r *GeoRepository) ListProvinces(countryCode string) ([]domain.Province, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT code, country_code, name, name_en, latitude, longitude
		FROM provinces
		WHERE country_code = $1
		ORDER BY name ASC
	`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Province
	for rows.Next() {
		var p domain.Province
		if err := rows.Scan(&p.Code, &p.CountryCode, &p.Name, &p.NameEn, &p.Latitude, &p.Longitude); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *GeoRepository) ListWards(provinceCode string) ([]domain.Ward, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT code, province_code, name, name_en, latitude, longitude
		FROM wards
		WHERE province_code = $1
		ORDER BY name ASC
	`, provinceCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Ward
	for rows.Next() {
		var w domain.Ward
		if err := rows.Scan(&w.Code, &w.ProvinceCode, &w.Name, &w.NameEn, &w.Latitude, &w.Longitude); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *GeoRepository) GetProvince(code string) (domain.Province, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p domain.Province
	err := r.pool.QueryRow(ctx, `
		SELECT code, country_code, name, name_en, latitude, longitude
		FROM provinces WHERE code = $1
	`, code).Scan(&p.Code, &p.CountryCode, &p.Name, &p.NameEn, &p.Latitude, &p.Longitude)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Province{}, domain.ErrProvinceNotFound
	}
	return p, err
}

func (r *GeoRepository) GetWard(code string) (domain.Ward, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var w domain.Ward
	err := r.pool.QueryRow(ctx, `
		SELECT code, province_code, name, name_en, latitude, longitude
		FROM wards WHERE code = $1
	`, code).Scan(&w.Code, &w.ProvinceCode, &w.Name, &w.NameEn, &w.Latitude, &w.Longitude)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ward{}, domain.ErrWardNotFound
	}
	return w, err
}
