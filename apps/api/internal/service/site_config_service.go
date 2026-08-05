package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

const (
	SiteConfigHomeBanner     = "home_banner"
	SiteConfigHomeFeatured   = "home_featured"
	SiteConfigContactChannel = "contact_channels"
	SiteConfigFooter         = "footer_config"
	SiteConfigGlobalCTA      = "global_cta"
	SiteConfigFeaturedSlots  = "featured_content_slots"
)

var allowedSiteConfigKeys = map[string]struct{}{
	SiteConfigHomeBanner:     {},
	SiteConfigHomeFeatured:   {},
	SiteConfigContactChannel: {},
	SiteConfigFooter:         {},
	SiteConfigGlobalCTA:      {},
	SiteConfigFeaturedSlots:  {},
}

type SiteConfigService struct {
	repo *repository.SiteConfigRepository
}

type SiteConfigUpsertInput struct {
	ConfigValue any `json:"config_value"`
}

func NewSiteConfigService(repo *repository.SiteConfigRepository) *SiteConfigService {
	return &SiteConfigService{repo: repo}
}

func (s *SiteConfigService) List() ([]domain.SiteConfig, error) {
	return s.repo.List()
}

func (s *SiteConfigService) GetByKey(key string) (domain.SiteConfig, error) {
	return s.repo.GetByKey(strings.TrimSpace(key))
}

func (s *SiteConfigService) Upsert(key string, input SiteConfigUpsertInput, updatedBy int64) error {
	normalizedKey := strings.TrimSpace(key)
	if _, ok := allowedSiteConfigKeys[normalizedKey]; !ok {
		return fmt.Errorf("unsupported config key")
	}

	if input.ConfigValue == nil {
		return fmt.Errorf("config value is required")
	}

	payload, err := json.Marshal(input.ConfigValue)
	if err != nil {
		return fmt.Errorf("invalid config value")
	}

	if string(payload) == "null" {
		return fmt.Errorf("config value is required")
	}

	return s.repo.Upsert(normalizedKey, string(payload), updatedBy)
}
