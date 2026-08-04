package commands

import (
	"context"
	"io"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type CreateMerchantCommand struct {
	Email        string
	DisplayName  string
	Password     string
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
}

type MerchantResult = queries.AccountDTO

type CreateMerchantHandler struct {
	merchants  domain.AccountRepository
	hasher     domain.PasswordHasher
	geo        domain.GeoRepository
	publicBase string
}

func NewCreateMerchantHandler(
	merchants domain.AccountRepository,
	hasher domain.PasswordHasher,
	geo domain.GeoRepository,
	publicBase string,
) *CreateMerchantHandler {
	return &CreateMerchantHandler{merchants: merchants, hasher: hasher, geo: geo, publicBase: publicBase}
}

func (h *CreateMerchantHandler) Handle(_ context.Context, cmd CreateMerchantCommand) (MerchantResult, error) {
	if _, err := h.merchants.FindByEmail(cmd.Email); err == nil {
		return MerchantResult{}, domain.ErrEmailTaken
	} else if err != domain.ErrAccountNotFound {
		return MerchantResult{}, err
	}

	account, err := domain.NewAccount(cmd.Email, cmd.DisplayName)
	if err != nil {
		return MerchantResult{}, err
	}
	if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
		return MerchantResult{}, err
	}
	if err := applyShopAddress(h.geo, &account, cmd.AddressLine, cmd.CountryCode, cmd.ProvinceCode, cmd.WardCode); err != nil {
		return MerchantResult{}, err
	}
	if err := h.merchants.Save(account); err != nil {
		return MerchantResult{}, err
	}
	return queries.ToAccountDTO(account, domain.RoleMerchant, h.publicBase, h.geo), nil
}

type UpdateMerchantCommand struct {
	ID           domain.AccountID
	Email        string
	DisplayName  string
	Password     string
	AddressLine  string
	CountryCode  string
	ProvinceCode string
	WardCode     string
	// KeepEmail skips email change (merchant self-update).
	KeepEmail bool
}

type UpdateMerchantHandler struct {
	merchants  domain.AccountRepository
	hasher     domain.PasswordHasher
	geo        domain.GeoRepository
	publicBase string
}

func NewUpdateMerchantHandler(
	merchants domain.AccountRepository,
	hasher domain.PasswordHasher,
	geo domain.GeoRepository,
	publicBase string,
) *UpdateMerchantHandler {
	return &UpdateMerchantHandler{merchants: merchants, hasher: hasher, geo: geo, publicBase: publicBase}
}

func (h *UpdateMerchantHandler) Handle(_ context.Context, cmd UpdateMerchantCommand) (MerchantResult, error) {
	account, err := h.merchants.FindByID(cmd.ID)
	if err != nil {
		return MerchantResult{}, err
	}

	if !cmd.KeepEmail {
		if err := account.ChangeEmail(cmd.Email); err != nil {
			return MerchantResult{}, err
		}
		existing, err := h.merchants.FindByEmail(account.Email)
		if err == nil && existing.ID != account.ID {
			return MerchantResult{}, domain.ErrEmailTaken
		}
		if err != nil && err != domain.ErrAccountNotFound {
			return MerchantResult{}, err
		}
	}

	account.Rename(cmd.DisplayName)

	if cmd.Password != "" {
		if err := account.SetPassword(h.hasher, cmd.Password); err != nil {
			return MerchantResult{}, err
		}
	}

	if err := applyShopAddress(h.geo, &account, cmd.AddressLine, cmd.CountryCode, cmd.ProvinceCode, cmd.WardCode); err != nil {
		return MerchantResult{}, err
	}

	if err := h.merchants.Save(account); err != nil {
		return MerchantResult{}, err
	}
	return queries.ToAccountDTO(account, domain.RoleMerchant, h.publicBase, h.geo), nil
}

func applyShopAddress(
	geo domain.GeoRepository,
	account *domain.Account,
	addressLine, countryCode, provinceCode, wardCode string,
) error {
	addressLine = strings.TrimSpace(addressLine)
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	provinceCode = strings.TrimSpace(provinceCode)
	wardCode = strings.TrimSpace(wardCode)

	if addressLine == "" && countryCode == "" && provinceCode == "" && wardCode == "" {
		account.ClearShopAddress()
		return nil
	}

	fields, err := validateAddressGeo(geo, domain.AddressFields{
		AddressLine:  addressLine,
		CountryCode:  countryCode,
		ProvinceCode: provinceCode,
		WardCode:     wardCode,
	})
	if err != nil {
		return err
	}
	account.SetShopAddress(fields.AddressLine, fields.CountryCode, fields.ProvinceCode, fields.WardCode)
	return nil
}

type UploadMerchantAvatarCommand struct {
	ID          domain.AccountID
	Filename    string
	ContentType string
	Size        int64
	Body        io.Reader
}

type UploadMerchantAvatarHandler struct {
	merchants  domain.AccountRepository
	store      storage.ObjectStore
	publicBase string
	geo        domain.GeoRepository
}

func NewUploadMerchantAvatarHandler(
	merchants domain.AccountRepository,
	store storage.ObjectStore,
	publicBase string,
	geo domain.GeoRepository,
) *UploadMerchantAvatarHandler {
	return &UploadMerchantAvatarHandler{merchants: merchants, store: store, publicBase: publicBase, geo: geo}
}

func (h *UploadMerchantAvatarHandler) Handle(ctx context.Context, cmd UploadMerchantAvatarCommand) (MerchantResult, error) {
	if h.store == nil || !h.store.Enabled() {
		return MerchantResult{}, storage.ErrObjectStoreDisabled
	}
	account, err := h.merchants.FindByID(cmd.ID)
	if err != nil {
		return MerchantResult{}, err
	}
	ct := strings.ToLower(strings.TrimSpace(cmd.ContentType))
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
	default:
		return MerchantResult{}, domain.ErrInvalidAvatar
	}
	if cmd.Size <= 0 || cmd.Size > 5*1024*1024 {
		return MerchantResult{}, domain.ErrInvalidAvatar
	}

	key := h.store.NewMerchantAvatarKey(string(account.ID), cmd.Filename)
	if err := h.store.Upload(ctx, key, cmd.Body, ct, cmd.Size); err != nil {
		return MerchantResult{}, err
	}
	old := account.AvatarKey
	account.SetAvatarKey(key)
	if err := h.merchants.Save(account); err != nil {
		_ = h.store.Delete(ctx, key)
		return MerchantResult{}, err
	}
	if old != "" && old != key {
		_ = h.store.Delete(ctx, old)
	}
	return queries.ToAccountDTO(account, domain.RoleMerchant, h.publicBase, h.geo), nil
}

type DeleteMerchantAvatarCommand struct {
	ID domain.AccountID
}

type DeleteMerchantAvatarHandler struct {
	merchants  domain.AccountRepository
	store      storage.ObjectStore
	publicBase string
	geo        domain.GeoRepository
}

func NewDeleteMerchantAvatarHandler(
	merchants domain.AccountRepository,
	store storage.ObjectStore,
	publicBase string,
	geo domain.GeoRepository,
) *DeleteMerchantAvatarHandler {
	return &DeleteMerchantAvatarHandler{merchants: merchants, store: store, publicBase: publicBase, geo: geo}
}

func (h *DeleteMerchantAvatarHandler) Handle(ctx context.Context, cmd DeleteMerchantAvatarCommand) (MerchantResult, error) {
	account, err := h.merchants.FindByID(cmd.ID)
	if err != nil {
		return MerchantResult{}, err
	}
	old := account.AvatarKey
	account.SetAvatarKey("")
	if err := h.merchants.Save(account); err != nil {
		return MerchantResult{}, err
	}
	if old != "" && h.store != nil {
		_ = h.store.Delete(ctx, old)
	}
	return queries.ToAccountDTO(account, domain.RoleMerchant, h.publicBase, h.geo), nil
}
