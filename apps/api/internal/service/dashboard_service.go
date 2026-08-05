package service

import (
	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/repository"
)

type DashboardService struct {
	articleRepo  *repository.ArticleRepository
	eventRepo    *repository.EventRepository
	builderRepo  *repository.BuilderRepository
	leadRepo     *repository.LeadRepository
	auditLogRepo *repository.AuditLogRepository
}

func NewDashboardService(
	articleRepo *repository.ArticleRepository,
	eventRepo *repository.EventRepository,
	builderRepo *repository.BuilderRepository,
	leadRepo *repository.LeadRepository,
	auditLogRepo *repository.AuditLogRepository,
) *DashboardService {
	return &DashboardService{
		articleRepo:  articleRepo,
		eventRepo:    eventRepo,
		builderRepo:  builderRepo,
		leadRepo:     leadRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *DashboardService) Stats() (domain.DashboardStats, error) {
	articleCount, err := s.articleRepo.CountAll()
	if err != nil {
		return domain.DashboardStats{}, err
	}

	publishedArticleCount, err := s.articleRepo.CountByStatus("published")
	if err != nil {
		return domain.DashboardStats{}, err
	}

	eventCount, err := s.eventRepo.CountAll()
	if err != nil {
		return domain.DashboardStats{}, err
	}

	publishedEventCount, err := s.eventRepo.CountByStatus("published")
	if err != nil {
		return domain.DashboardStats{}, err
	}

	builderCount, err := s.builderRepo.CountAll()
	if err != nil {
		return domain.DashboardStats{}, err
	}

	publishedBuilderCount, err := s.builderRepo.CountByStatus("published")
	if err != nil {
		return domain.DashboardStats{}, err
	}

	leadCount, err := s.leadRepo.CountAll()
	if err != nil {
		return domain.DashboardStats{}, err
	}

	leadStatusDistribution, err := s.leadRepo.StatusCounts()
	if err != nil {
		return domain.DashboardStats{}, err
	}

	recentActivities, err := s.auditLogRepo.List(6)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	if recentActivities == nil {
		recentActivities = make([]domain.AuditLog, 0)
	}

	recentActionableLeads, err := s.leadRepo.ListActionable(5)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	if recentActionableLeads == nil {
		recentActionableLeads = make([]domain.Lead, 0)
	}

	return domain.DashboardStats{
		ArticleCount:           articleCount,
		PublishedArticleCount:  publishedArticleCount,
		EventCount:             eventCount,
		PublishedEventCount:    publishedEventCount,
		BuilderCount:           builderCount,
		PublishedBuilderCount:  publishedBuilderCount,
		LeadCount:              leadCount,
		ActionableLeadCount:    leadStatusDistribution.New + leadStatusDistribution.Contacted + leadStatusDistribution.Following,
		LeadStatusDistribution: leadStatusDistribution,
		RecentActivities:       recentActivities,
		RecentActionableLeads:  recentActionableLeads,
	}, nil
}
