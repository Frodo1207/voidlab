package http

import (
	"errors"
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"voidlabai/apps/api/internal/domain"
	"voidlabai/apps/api/internal/http/middleware"
	"voidlabai/apps/api/internal/repository"
	"voidlabai/apps/api/internal/service"
)

type BuilderNetworkHandler struct {
	service           *service.BuilderNetworkService
	agentTokenService *service.BuilderNetworkAgentTokenService
	auditLogService   *service.AuditLogService
}

func NewBuilderNetworkHandler(
	networkService *service.BuilderNetworkService,
	agentTokenService *service.BuilderNetworkAgentTokenService,
	auditLogService *service.AuditLogService,
) *BuilderNetworkHandler {
	return &BuilderNetworkHandler{
		service:           networkService,
		agentTokenService: agentTokenService,
		auditLogService:   auditLogService,
	}
}

func (h *BuilderNetworkHandler) VerifyInvite(ctx *gin.Context) {
	var input struct {
		Code string `json:"code"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	valid, err := h.service.VerifyInvite(input.Code)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, gin.H{"valid": valid})
}

func (h *BuilderNetworkHandler) Register(ctx *gin.Context) {
	var input struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		InviteCode string `json:"invite_code"`
		Profile    struct {
			DisplayName   string `json:"display_name"`
			PrimaryRole   string `json:"primary_role"`
			Headline      string `json:"headline"`
			City          string `json:"city"`
			Availability  string `json:"availability"`
			AvailableFrom string `json:"available_from"`
		} `json:"profile"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	session, err := h.service.Register(input.Email, input.Password, input.InviteCode, domain.NetworkProfile{
		DisplayName: input.Profile.DisplayName, PrimaryRole: input.Profile.PrimaryRole,
		Headline: input.Profile.Headline, City: input.Profile.City,
		Availability: input.Profile.Availability, AvailableFrom: input.Profile.AvailableFrom,
	})
	if err != nil {
		networkError(ctx, err)
		return
	}
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": session})
}

func (h *BuilderNetworkHandler) Login(ctx *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	session, err := h.service.Login(input.Email, input.Password)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, session)
}

func (h *BuilderNetworkHandler) Logout(ctx *gin.Context) {
	parts := strings.Fields(ctx.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	if err := h.service.Logout(parts[1]); err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, gin.H{"logged_out": true})
}

func (h *BuilderNetworkHandler) ListAgentTokens(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	tokens, err := h.agentTokenService.List(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, tokens)
}

func (h *BuilderNetworkHandler) CreateAgentToken(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input service.NetworkAgentTokenCreateInput
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	result, err := h.agentTokenService.Create(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	if err := recordAuditLog(ctx, h.auditLogService, domain.User{}, "create", "network_agent_token",
		toInt64Pointer(result.Record.ID), result.Record.Name, gin.H{"scopes": result.Record.Scopes}); err != nil {
		ctx.Error(err)
	}
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": result})
}

func (h *BuilderNetworkHandler) UpdateAgentTokenStatus(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	tokenID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || tokenID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid token id")
		return
	}
	var input struct {
		IsActive *bool `json:"is_active"`
	}
	if ctx.ShouldBindJSON(&input) != nil || input.IsActive == nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "is_active is required")
		return
	}
	record, err := h.agentTokenService.UpdateStatus(account.ID, tokenID, *input.IsActive)
	if err != nil {
		networkError(ctx, err)
		return
	}
	if err := recordAuditLog(ctx, h.auditLogService, domain.User{}, "update_status", "network_agent_token",
		toInt64Pointer(record.ID), record.Name, gin.H{"is_active": record.IsActive}); err != nil {
		ctx.Error(err)
	}
	OK(ctx, record)
}

func (h *BuilderNetworkHandler) PublicProfiles(ctx *gin.Context) {
	profiles, err := h.service.PublicProfiles()
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, profiles)
}

func (h *BuilderNetworkHandler) PublicProfileDetail(ctx *gin.Context) {
	profile, err := h.service.PublicProfileBySlug(ctx.Param("slug"))
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, profile)
}

func (h *BuilderNetworkHandler) BrowsePublicProfiles(ctx *gin.Context) {
	query, ok := parseNetworkDiscoveryQuery(ctx)
	if !ok {
		return
	}
	profiles, err := h.service.BrowsePublicProfiles(query)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, profiles)
}

func (h *BuilderNetworkHandler) PublicWorks(ctx *gin.Context) {
	works, err := h.service.PublicWorks()
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, works)
}

func (h *BuilderNetworkHandler) PublicWorkDetail(ctx *gin.Context) {
	work, err := h.service.PublicWorkBySlug(ctx.Param("slug"))
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, work)
}

func (h *BuilderNetworkHandler) BrowsePublicWorks(ctx *gin.Context) {
	query, ok := parseNetworkDiscoveryQuery(ctx)
	if !ok {
		return
	}
	works, err := h.service.BrowsePublicWorks(query)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, works)
}

func (h *BuilderNetworkHandler) PublicProjects(ctx *gin.Context) {
	projects, err := h.service.PublicProjects()
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, projects)
}

func (h *BuilderNetworkHandler) PublicProjectDetail(ctx *gin.Context) {
	project, err := h.service.PublicProjectBySlug(ctx.Param("slug"))
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, project)
}

func (h *BuilderNetworkHandler) PublicOpportunities(ctx *gin.Context) {
	opportunities, err := h.service.PublicOpportunities()
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, opportunities)
}

func (h *BuilderNetworkHandler) PublicOpportunityDetail(ctx *gin.Context) {
	opportunity, err := h.service.PublicOpportunityBySlug(ctx.Param("slug"))
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, opportunity)
}

func (h *BuilderNetworkHandler) BrowsePublicOpportunities(ctx *gin.Context) {
	query, ok := parseNetworkDiscoveryQuery(ctx)
	if !ok {
		return
	}
	opportunities, err := h.service.BrowsePublicOpportunities(query)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, opportunities)
}

func (h *BuilderNetworkHandler) CreateInvite(ctx *gin.Context) {
	var input struct {
		MaxUses   int    `json:"max_uses"`
		ExpiresAt string `json:"expires_at"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	user, ok := middleware.CurrentUser(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	code, err := h.service.CreateInvite(input.MaxUses, input.ExpiresAt, user.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "create", "network_invite_code", nil, "BuilderNetwork invitation", gin.H{
		"max_uses":   input.MaxUses,
		"expires_at": input.ExpiresAt,
	})
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": gin.H{"code": code, "max_uses": input.MaxUses, "expires_at": input.ExpiresAt}})
}

func (h *BuilderNetworkHandler) ListInvites(ctx *gin.Context) {
	invites, err := h.service.ListInvites()
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "read", "network_invite_codes", nil, "BuilderNetwork invitations", gin.H{
		"count": len(invites),
	})
	OK(ctx, invites)
}

func (h *BuilderNetworkHandler) UpdateInviteStatus(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid invitation id")
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	if err := h.service.UpdateInviteStatus(id, input.Status); err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update_status", "network_invite_code", toInt64Pointer(id), "BuilderNetwork invitation", gin.H{
		"status": input.Status,
	})
	OK(ctx, gin.H{"updated": true})
}

func (h *BuilderNetworkHandler) Me(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	_, profile, err := h.service.Me(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, profile)
}

func (h *BuilderNetworkHandler) UpdateProfile(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input domain.NetworkProfile
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	profile, err := h.service.UpdateProfile(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update", "network_profile", toInt64Pointer(profile.ID), profile.DisplayName, gin.H{})
	OK(ctx, profile)
}

func (h *BuilderNetworkHandler) Contacts(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	details, shareEmail, shareWeChat, sharePhone, err := h.service.Contacts(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "read", "network_private_contacts", toInt64Pointer(account.ID), "private contacts", gin.H{
		"account_id": account.ID,
	})
	OK(ctx, gin.H{"contacts": details, "share_email": shareEmail, "share_wechat": shareWeChat, "share_phone": sharePhone})
}

func (h *BuilderNetworkHandler) UpdateContacts(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input struct {
		Contacts    domain.NetworkContactDetails `json:"contacts"`
		ShareEmail  bool                         `json:"share_email"`
		ShareWeChat bool                         `json:"share_wechat"`
		SharePhone  bool                         `json:"share_phone"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	if err := h.service.UpdateContacts(account.ID, input.Contacts, input.ShareEmail, input.ShareWeChat, input.SharePhone); err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update", "network_private_contacts", toInt64Pointer(account.ID), "private contacts", gin.H{
		"share_email": input.ShareEmail, "share_wechat": input.ShareWeChat, "share_phone": input.SharePhone,
	})
	OK(ctx, gin.H{"updated": true})
}

func (h *BuilderNetworkHandler) CreateWork(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input domain.NetworkWork
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	work, err := h.service.CreateWork(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "create", "network_work", toInt64Pointer(work.ID), work.Title, gin.H{"status": work.Status})
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": work})
}

func (h *BuilderNetworkHandler) MyWorks(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	works, err := h.service.MyWorks(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, works)
}

func (h *BuilderNetworkHandler) UpdateWork(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	workID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || workID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid work id")
		return
	}
	var input domain.NetworkWork
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	work, err := h.service.UpdateWork(account.ID, workID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update", "network_work", toInt64Pointer(work.ID), work.Title, gin.H{"status": work.Status})
	OK(ctx, work)
}

func (h *BuilderNetworkHandler) MyProjects(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	projects, err := h.service.MyProjects(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, projects)
}

func (h *BuilderNetworkHandler) MyProjectDetail(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	projectID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || projectID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid project id")
		return
	}
	project, err := h.service.ProjectByID(account.ID, projectID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, project)
}

func (h *BuilderNetworkHandler) CreateProject(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input domain.NetworkProject
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	project, err := h.service.CreateProject(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "create", "network_project", toInt64Pointer(project.ID), project.Title, gin.H{"status": project.Status})
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": project})
}

func (h *BuilderNetworkHandler) UpdateProject(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	projectID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || projectID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid project id")
		return
	}
	var input domain.NetworkProject
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	project, err := h.service.UpdateProject(account.ID, projectID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update", "network_project", toInt64Pointer(project.ID), project.Title, gin.H{"status": project.Status})
	OK(ctx, project)
}

func (h *BuilderNetworkHandler) CreateOpportunity(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input domain.NetworkOpportunity
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	opportunity, err := h.service.CreateOpportunity(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "create", "network_opportunity", toInt64Pointer(opportunity.ID), opportunity.Title, gin.H{"status": opportunity.Status})
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": opportunity})
}

func (h *BuilderNetworkHandler) MyOpportunities(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	opportunities, err := h.service.MyOpportunities(account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, opportunities)
}

func (h *BuilderNetworkHandler) UpdateOpportunity(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	opportunityID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || opportunityID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid opportunity id")
		return
	}
	var input domain.NetworkOpportunity
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	opportunity, err := h.service.UpdateOpportunity(account.ID, opportunityID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update", "network_opportunity", toInt64Pointer(opportunity.ID), opportunity.Title, gin.H{"status": opportunity.Status})
	OK(ctx, opportunity)
}

func (h *BuilderNetworkHandler) CreateContactRequest(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	var input domain.NetworkContactRequest
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	request, err := h.service.CreateContactRequest(account.ID, input)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "create", "network_contact_request", toInt64Pointer(request.ID), "contact request", gin.H{
		"target_type": request.TargetType, "target_id": request.TargetID,
		"share_email": request.ShareEmail, "share_wechat": request.ShareWeChat, "share_phone": request.SharePhone,
	})
	ctx.JSON(stdhttp.StatusCreated, gin.H{"code": 0, "message": "ok", "data": request})
}

func (h *BuilderNetworkHandler) ContactRequests(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	direction := strings.ToLower(strings.TrimSpace(ctx.DefaultQuery("direction", "incoming")))
	if direction != "incoming" && direction != "outgoing" {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "direction must be incoming or outgoing")
		return
	}
	requests, err := h.service.ContactRequests(account.ID, direction == "incoming")
	if err != nil {
		networkError(ctx, err)
		return
	}
	OK(ctx, requests)
}

func (h *BuilderNetworkHandler) AcceptContactRequest(ctx *gin.Context) {
	h.updateContactRequest(ctx, "accept", h.service.AcceptRequest)
}

func (h *BuilderNetworkHandler) DeclineContactRequest(ctx *gin.Context) {
	h.updateContactRequest(ctx, "decline", h.service.DeclineRequest)
}

func (h *BuilderNetworkHandler) WithdrawContactRequest(ctx *gin.Context) {
	h.updateContactRequest(ctx, "withdraw", h.service.WithdrawRequest)
}

func (h *BuilderNetworkHandler) UpdateContactRequestProgress(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	requestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || requestID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request id")
		return
	}
	var input struct {
		ProgressStatus string `json:"progress_status"`
	}
	if ctx.ShouldBindJSON(&input) != nil {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request")
		return
	}
	if err := h.service.UpdateContactRequestProgress(requestID, account.ID, input.ProgressStatus); err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "update_progress", "network_contact_request", toInt64Pointer(requestID), "contact request", gin.H{
		"request_id": requestID, "progress_status": input.ProgressStatus,
	})
	OK(ctx, gin.H{"updated": true, "progress_status": input.ProgressStatus})
}

func (h *BuilderNetworkHandler) ContactExchange(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	requestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || requestID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request id")
		return
	}
	exchange, err := h.service.ContactExchange(requestID, account.ID)
	if err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "read", "network_contact_exchange", toInt64Pointer(requestID), "contact exchange", gin.H{
		"request_id": requestID,
	})
	OK(ctx, exchange)
}

func (h *BuilderNetworkHandler) RevokeContactExchange(ctx *gin.Context) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	requestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || requestID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request id")
		return
	}
	if err := h.service.RevokeContactExchange(requestID, account.ID); err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, "revoke", "network_contact_exchange", toInt64Pointer(requestID), "contact exchange", gin.H{
		"request_id": requestID,
	})
	OK(ctx, gin.H{"revoked": true})
}

func (h *BuilderNetworkHandler) updateContactRequest(ctx *gin.Context, action string, update func(int64, int64) error) {
	account, ok := middleware.CurrentNetworkAccount(ctx)
	if !ok {
		Fail(ctx, stdhttp.StatusUnauthorized, 4010, "unauthorized")
		return
	}
	requestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || requestID <= 0 {
		Fail(ctx, stdhttp.StatusBadRequest, 4000, "invalid request id")
		return
	}
	if err := update(requestID, account.ID); err != nil {
		networkError(ctx, err)
		return
	}
	recordNetworkAudit(ctx, h.auditLogService, action, "network_contact_request", toInt64Pointer(requestID), "contact request", gin.H{
		"request_id": requestID,
	})
	OK(ctx, gin.H{"updated": true})
}

func parseNetworkDiscoveryQuery(ctx *gin.Context) (domain.NetworkDiscoveryQuery, bool) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 1_000_000 {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "page must be between 1 and 1000000")
		return domain.NetworkDiscoveryQuery{}, false
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "page_size must be between 1 and 100")
		return domain.NetworkDiscoveryQuery{}, false
	}
	query := domain.NetworkDiscoveryQuery{
		Query:        strings.TrimSpace(ctx.Query("q")),
		Role:         strings.TrimSpace(ctx.Query("role")),
		Skill:        strings.TrimSpace(ctx.Query("skill")),
		City:         strings.TrimSpace(ctx.Query("city")),
		Availability: strings.TrimSpace(ctx.Query("availability")),
		Type:         strings.TrimSpace(ctx.Query("type")),
		Page:         page,
		PageSize:     pageSize,
	}
	if len([]rune(query.Query)) > 100 || len([]rune(query.Role)) > 100 ||
		len([]rune(query.Skill)) > 100 || len([]rune(query.City)) > 100 ||
		len([]rune(query.Type)) > 100 {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "discovery filters must be 100 characters or fewer")
		return domain.NetworkDiscoveryQuery{}, false
	}
	if query.Availability != "" && query.Availability != "open" &&
		query.Availability != "limited" && query.Availability != "unavailable" {
		Fail(ctx, stdhttp.StatusBadRequest, 4001, "availability must be open, limited, or unavailable")
		return domain.NetworkDiscoveryQuery{}, false
	}
	if value := strings.TrimSpace(ctx.Query("remote")); value != "" {
		remote, err := strconv.ParseBool(value)
		if err != nil {
			Fail(ctx, stdhttp.StatusBadRequest, 4001, "remote must be true or false")
			return domain.NetworkDiscoveryQuery{}, false
		}
		query.Remote = &remote
	}
	return query, true
}

func networkError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNetworkValidation):
		Fail(ctx, stdhttp.StatusBadRequest, 4001, err.Error())
	case errors.Is(err, service.ErrNetworkUnauthorized):
		Fail(ctx, stdhttp.StatusUnauthorized, 4011, err.Error())
	case errors.Is(err, repository.ErrNetworkInviteInvalid):
		Fail(ctx, stdhttp.StatusBadRequest, 4002, err.Error())
	case errors.Is(err, repository.ErrNetworkConflict):
		Fail(ctx, stdhttp.StatusConflict, 4090, err.Error())
	case errors.Is(err, repository.ErrNetworkNotFound):
		Fail(ctx, stdhttp.StatusNotFound, 4040, err.Error())
	default:
		ctx.Error(err)
		Fail(ctx, stdhttp.StatusInternalServerError, 5000, "internal server error")
	}
}
