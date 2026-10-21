package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"voidlabai/apps/api/internal/domain"
)

var ErrNetworkNotFound = errors.New("builder network record not found")
var ErrNetworkConflict = errors.New("builder network record conflicts with existing data")
var ErrNetworkInviteInvalid = errors.New("invitation code is invalid or exhausted")

type BuilderNetworkRepository struct {
	db *sql.DB
}

func NewBuilderNetworkRepository(db *sql.DB) *BuilderNetworkRepository {
	return &BuilderNetworkRepository{db: db}
}

func (r *BuilderNetworkRepository) CreateInvite(codeHash string, maxUses int, expiresAt string, createdBy int64) error {
	_, err := r.db.Exec(
		`INSERT INTO network_invite_codes (code_hash, max_uses, expires_at, status, created_by) VALUES (?, ?, NULLIF(?, ''), 'active', ?)`,
		codeHash, maxUses, expiresAt, createdBy,
	)
	return err
}

func (r *BuilderNetworkRepository) ListInvites() ([]domain.NetworkInviteCode, error) {
	rows, err := r.db.Query(`
		SELECT id, max_uses, used_count, COALESCE(expires_at, ''), status, created_by, created_at
		FROM network_invite_codes
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	invites := make([]domain.NetworkInviteCode, 0)
	for rows.Next() {
		var invite domain.NetworkInviteCode
		if err := rows.Scan(&invite.ID, &invite.MaxUses, &invite.UsedCount, &invite.ExpiresAt,
			&invite.Status, &invite.CreatedBy, &invite.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (r *BuilderNetworkRepository) UpdateInviteStatus(id int64, status string) error {
	result, err := r.db.Exec(`UPDATE network_invite_codes SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNetworkNotFound
	}
	return nil
}

func (r *BuilderNetworkRepository) VerifyInvite(codeHash string) (bool, error) {
	var valid int
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM network_invite_codes
			WHERE code_hash = ? AND status = 'active' AND used_count < max_uses
			  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		)
	`, codeHash).Scan(&valid)
	return valid == 1, err
}

func (r *BuilderNetworkRepository) Register(email, passwordHash, inviteHash string, profile domain.NetworkProfile) (domain.NetworkAccount, domain.NetworkProfile, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	defer tx.Rollback()

	var inviteID int64
	var maxUses, usedCount int
	err = tx.QueryRow(`
		SELECT id, max_uses, used_count
		FROM network_invite_codes
		WHERE code_hash = ? AND status = 'active'
		  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`, inviteHash).Scan(&inviteID, &maxUses, &usedCount)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, ErrNetworkInviteInvalid
	}
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	if usedCount >= maxUses {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, ErrNetworkInviteInvalid
	}

	result, err := tx.Exec(
		`INSERT INTO network_accounts (email, password_hash, status) VALUES (?, ?, 'active')`,
		strings.ToLower(strings.TrimSpace(email)), passwordHash,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NetworkAccount{}, domain.NetworkProfile{}, ErrNetworkConflict
		}
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	accountID, err := result.LastInsertId()
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}

	profile.AccountID = accountID
	profile.Slug = profileSlug(profile.DisplayName, accountID)
	result, err = tx.Exec(`
		INSERT INTO network_profiles
			(account_id, slug, display_name, primary_role, headline, city, availability, available_from)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, accountID, profile.Slug, profile.DisplayName, profile.PrimaryRole, profile.Headline, profile.City, profile.Availability, profile.AvailableFrom)
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	profile.ID, err = result.LastInsertId()
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	if _, err = tx.Exec(`INSERT INTO network_private_contacts (profile_id, email) VALUES (?, ?)`, profile.ID, strings.ToLower(strings.TrimSpace(email))); err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	update, err := tx.Exec(`UPDATE network_invite_codes SET used_count = used_count + 1 WHERE id = ? AND used_count < max_uses`, inviteID)
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	updated, err := update.RowsAffected()
	if err != nil || updated != 1 {
		if err != nil {
			return domain.NetworkAccount{}, domain.NetworkProfile{}, err
		}
		return domain.NetworkAccount{}, domain.NetworkProfile{}, ErrNetworkInviteInvalid
	}
	if err := tx.Commit(); err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}

	account, savedProfile, err := r.GetAccountProfile(accountID)
	return account, savedProfile, err
}

func (r *BuilderNetworkRepository) FindAccountByEmail(email string) (domain.NetworkAccount, string, error) {
	var account domain.NetworkAccount
	var passwordHash string
	err := r.db.QueryRow(`
		SELECT id, email, status, created_at, password_hash
		FROM network_accounts WHERE email = ?
	`, strings.ToLower(strings.TrimSpace(email))).Scan(
		&account.ID, &account.Email, &account.Status, &account.CreatedAt, &passwordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAccount{}, "", ErrNetworkNotFound
	}
	return account, passwordHash, err
}

func (r *BuilderNetworkRepository) CreateSession(accountID int64, tokenHash, expiresAt string) error {
	_, err := r.db.Exec(`INSERT INTO network_sessions (account_id, token_hash, expires_at) VALUES (?, ?, ?)`, accountID, tokenHash, expiresAt)
	return err
}

func (r *BuilderNetworkRepository) DeleteSession(tokenHash string) error {
	_, err := r.db.Exec(`DELETE FROM network_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

func (r *BuilderNetworkRepository) ResolveSession(tokenHash string) (domain.NetworkAccount, error) {
	var account domain.NetworkAccount
	err := r.db.QueryRow(`
		SELECT a.id, a.email, a.status, a.created_at
		FROM network_sessions s
		JOIN network_accounts a ON a.id = s.account_id
		WHERE s.token_hash = ? AND s.expires_at > CURRENT_TIMESTAMP AND a.status = 'active'
	`, tokenHash).Scan(&account.ID, &account.Email, &account.Status, &account.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAccount{}, ErrNetworkNotFound
	}
	return account, err
}

func (r *BuilderNetworkRepository) GetAccountProfile(accountID int64) (domain.NetworkAccount, domain.NetworkProfile, error) {
	var account domain.NetworkAccount
	var profile domain.NetworkProfile
	var skillsJSON, modesJSON string
	err := r.db.QueryRow(`
		SELECT a.id, a.email, a.status, a.created_at,
		       p.id, p.slug, p.display_name, p.primary_role, p.headline, p.bio, p.city,
		       p.skills_json, p.collaboration_modes_json, p.availability, p.available_from, p.created_at, p.updated_at
		FROM network_accounts a
		JOIN network_profiles p ON p.account_id = a.id
		WHERE a.id = ?
	`, accountID).Scan(
		&account.ID, &account.Email, &account.Status, &account.CreatedAt,
		&profile.ID, &profile.Slug, &profile.DisplayName, &profile.PrimaryRole, &profile.Headline, &profile.Bio, &profile.City,
		&skillsJSON, &modesJSON, &profile.Availability, &profile.AvailableFrom, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, ErrNetworkNotFound
	}
	if err != nil {
		return domain.NetworkAccount{}, domain.NetworkProfile{}, err
	}
	profile.AccountID = accountID
	_ = json.Unmarshal([]byte(skillsJSON), &profile.Skills)
	_ = json.Unmarshal([]byte(modesJSON), &profile.Collaboration)
	return account, profile, nil
}

func (r *BuilderNetworkRepository) UpdateProfile(profile domain.NetworkProfile, accountID int64) (domain.NetworkProfile, error) {
	skillsJSON, _ := json.Marshal(profile.Skills)
	modesJSON, _ := json.Marshal(profile.Collaboration)
	result, err := r.db.Exec(`
		UPDATE network_profiles
		SET display_name = ?, primary_role = ?, headline = ?, bio = ?, city = ?,
		    skills_json = ?, collaboration_modes_json = ?, availability = ?, available_from = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE account_id = ?
	`, profile.DisplayName, profile.PrimaryRole, profile.Headline, profile.Bio, profile.City,
		string(skillsJSON), string(modesJSON), profile.Availability, profile.AvailableFrom, accountID)
	if err != nil {
		return domain.NetworkProfile{}, err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.NetworkProfile{}, ErrNetworkNotFound
	}
	_, updated, err := r.GetAccountProfile(accountID)
	return updated, err
}

func (r *BuilderNetworkRepository) SaveContacts(profileID int64, email, wechat, phone string, shareEmail, shareWeChat, sharePhone bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
		INSERT INTO network_private_contacts (profile_id, email, wechat, phone, share_email, share_wechat, share_phone)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id) DO UPDATE SET
			email = excluded.email, wechat = excluded.wechat, phone = excluded.phone,
			share_email = excluded.share_email, share_wechat = excluded.share_wechat,
			share_phone = excluded.share_phone, updated_at = CURRENT_TIMESTAMP
	`, profileID, email, wechat, phone, boolInt(shareEmail), boolInt(shareWeChat), boolInt(sharePhone)); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE network_contact_exchanges
		SET applicant_share_email = CASE WHEN ? = 0 THEN 0 ELSE applicant_share_email END,
		    applicant_share_wechat = CASE WHEN ? = 0 THEN 0 ELSE applicant_share_wechat END,
		    applicant_share_phone = CASE WHEN ? = 0 THEN 0 ELSE applicant_share_phone END
		WHERE revoked_at IS NULL AND request_id IN (
			SELECT id FROM network_contact_requests
			WHERE from_profile_id = ?
		)
	`, boolInt(shareEmail), boolInt(shareWeChat), boolInt(sharePhone), profileID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE network_contact_exchanges
		SET recipient_share_email = CASE WHEN ? = 0 THEN 0 ELSE recipient_share_email END,
		    recipient_share_wechat = CASE WHEN ? = 0 THEN 0 ELSE recipient_share_wechat END,
		    recipient_share_phone = CASE WHEN ? = 0 THEN 0 ELSE recipient_share_phone END
		WHERE revoked_at IS NULL AND request_id IN (
			SELECT id FROM network_contact_requests
			WHERE target_profile_id = ?
		)
	`, boolInt(shareEmail), boolInt(shareWeChat), boolInt(sharePhone), profileID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *BuilderNetworkRepository) GetContacts(profileID int64) (domain.NetworkContactDetails, bool, bool, bool, error) {
	var details domain.NetworkContactDetails
	var shareEmail, shareWeChat, sharePhone int
	err := r.db.QueryRow(`
		SELECT email, wechat, phone, share_email, share_wechat, share_phone
		FROM network_private_contacts WHERE profile_id = ?
	`, profileID).Scan(&details.Email, &details.WeChat, &details.Phone, &shareEmail, &shareWeChat, &sharePhone)
	if errors.Is(err, sql.ErrNoRows) {
		return details, false, false, false, nil
	}
	return details, shareEmail != 0, shareWeChat != 0, sharePhone != 0, err
}

func (r *BuilderNetworkRepository) ListPublicProfiles() ([]domain.NetworkProfile, error) {
	rows, err := r.db.Query(`
		SELECT id, slug, display_name, primary_role, headline, bio, city,
		       skills_json, collaboration_modes_json, availability, available_from, created_at, updated_at
		FROM network_profiles WHERE status = 'published'
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]domain.NetworkProfile, 0)
	for rows.Next() {
		var profile domain.NetworkProfile
		var skillsJSON, modesJSON string
		if err := rows.Scan(&profile.ID, &profile.Slug, &profile.DisplayName, &profile.PrimaryRole, &profile.Headline,
			&profile.Bio, &profile.City, &skillsJSON, &modesJSON, &profile.Availability, &profile.AvailableFrom,
			&profile.CreatedAt, &profile.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(skillsJSON), &profile.Skills)
		_ = json.Unmarshal([]byte(modesJSON), &profile.Collaboration)
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

func (r *BuilderNetworkRepository) BrowsePublicProfiles(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkProfile], error) {
	page, pageSize, offset := normalizePage(query.Page, query.PageSize)
	conditions := []string{"status = 'published'"}
	args := make([]any, 0)
	addNetworkLike(&conditions, &args, []string{"display_name", "primary_role", "headline", "bio"}, query.Query)
	addNetworkLike(&conditions, &args, []string{"primary_role"}, query.Role)
	addNetworkLike(&conditions, &args, []string{"skills_json"}, query.Skill)
	addNetworkLike(&conditions, &args, []string{"city"}, query.City)
	addNetworkLike(&conditions, &args, []string{"availability"}, query.Availability)
	where := strings.Join(conditions, " AND ")

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM network_profiles WHERE `+where, args...).Scan(&total); err != nil {
		return domain.NetworkPage[domain.NetworkProfile]{}, err
	}
	rows, err := r.db.Query(`
		SELECT id, slug, display_name, primary_role, headline, bio, city,
		       skills_json, collaboration_modes_json, availability, available_from, created_at, updated_at
		FROM network_profiles WHERE `+where+`
		ORDER BY updated_at DESC, id DESC LIMIT ? OFFSET ?
	`, append(args, pageSize, offset)...)
	if err != nil {
		return domain.NetworkPage[domain.NetworkProfile]{}, err
	}
	defer rows.Close()
	profiles := make([]domain.NetworkProfile, 0)
	for rows.Next() {
		var profile domain.NetworkProfile
		var skillsJSON, modesJSON string
		if err := rows.Scan(&profile.ID, &profile.Slug, &profile.DisplayName, &profile.PrimaryRole, &profile.Headline,
			&profile.Bio, &profile.City, &skillsJSON, &modesJSON, &profile.Availability, &profile.AvailableFrom,
			&profile.CreatedAt, &profile.UpdatedAt); err != nil {
			return domain.NetworkPage[domain.NetworkProfile]{}, err
		}
		_ = json.Unmarshal([]byte(skillsJSON), &profile.Skills)
		_ = json.Unmarshal([]byte(modesJSON), &profile.Collaboration)
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return domain.NetworkPage[domain.NetworkProfile]{}, err
	}
	return domain.NetworkPage[domain.NetworkProfile]{Items: profiles, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *BuilderNetworkRepository) CreateWork(profileID int64, work domain.NetworkWork) (domain.NetworkWork, error) {
	tagsJSON, _ := json.Marshal(work.Tags)
	tx, err := r.db.Begin()
	if err != nil {
		return domain.NetworkWork{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO network_works
			(profile_id, slug, title, summary, body, cover_url, demo_url, repo_url, tags_json, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, profileID, work.Slug, work.Title, work.Summary, work.Body, work.CoverURL, work.DemoURL, work.RepoURL, string(tagsJSON), work.Status)
	if err != nil {
		return domain.NetworkWork{}, err
	}
	work.ID, err = result.LastInsertId()
	if err != nil {
		return domain.NetworkWork{}, err
	}
	if err := replaceWorkMembers(tx, work.ID, profileID, work.Members); err != nil {
		return domain.NetworkWork{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.NetworkWork{}, err
	}
	return r.getWorkByID(work.ID)
}

func (r *BuilderNetworkRepository) ListPublicWorks() ([]domain.NetworkWork, error) {
	rows, err := r.db.Query(`
		SELECT w.id, w.profile_id, w.slug, w.title, w.summary, w.body, w.cover_url, w.demo_url, w.repo_url,
		       w.tags_json, w.status, w.created_at, w.updated_at, p.display_name
		FROM network_works w JOIN network_profiles p ON p.id = w.profile_id
		WHERE w.status = 'published' AND p.status = 'published'
		ORDER BY w.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	works := make([]domain.NetworkWork, 0)
	for rows.Next() {
		var work domain.NetworkWork
		var tagsJSON string
		if err := rows.Scan(&work.ID, &work.ProfileID, &work.Slug, &work.Title, &work.Summary, &work.Body,
			&work.CoverURL, &work.DemoURL, &work.RepoURL, &tagsJSON, &work.Status, &work.CreatedAt,
			&work.UpdatedAt, &work.AuthorName); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &work.Tags)
		works = append(works, work)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range works {
		members, err := r.listWorkMembers(works[i].ID)
		if err != nil {
			return nil, err
		}
		works[i].Members = members
	}
	return works, nil
}

func (r *BuilderNetworkRepository) BrowsePublicWorks(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkWork], error) {
	page, pageSize, offset := normalizePage(query.Page, query.PageSize)
	conditions := []string{"w.status = 'published'", "p.status = 'published'"}
	args := make([]any, 0)
	addNetworkLike(&conditions, &args, []string{"w.title", "w.summary", "w.body"}, query.Query)
	addNetworkLike(&conditions, &args, []string{"p.primary_role"}, query.Role)
	addNetworkLike(&conditions, &args, []string{"w.tags_json", "p.skills_json"}, query.Skill)
	addNetworkLike(&conditions, &args, []string{"p.city"}, query.City)
	addNetworkLike(&conditions, &args, []string{"p.availability"}, query.Availability)
	where := strings.Join(conditions, " AND ")
	join := " FROM network_works w JOIN network_profiles p ON p.id = w.profile_id WHERE " + where

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*)`+join, args...).Scan(&total); err != nil {
		return domain.NetworkPage[domain.NetworkWork]{}, err
	}
	rows, err := r.db.Query(`
		SELECT w.id, w.profile_id, w.slug, w.title, w.summary, w.body, w.cover_url, w.demo_url, w.repo_url,
		       w.tags_json, w.status, w.created_at, w.updated_at, p.display_name`+join+`
		ORDER BY w.created_at DESC, w.id DESC LIMIT ? OFFSET ?
	`, append(args, pageSize, offset)...)
	if err != nil {
		return domain.NetworkPage[domain.NetworkWork]{}, err
	}
	defer rows.Close()
	works := make([]domain.NetworkWork, 0)
	for rows.Next() {
		work, err := r.scanWork(rows)
		if err != nil {
			return domain.NetworkPage[domain.NetworkWork]{}, err
		}
		works = append(works, work)
	}
	if err := rows.Err(); err != nil {
		return domain.NetworkPage[domain.NetworkWork]{}, err
	}
	if err := rows.Close(); err != nil {
		return domain.NetworkPage[domain.NetworkWork]{}, err
	}
	for i := range works {
		works[i].Members, err = r.listWorkMembers(works[i].ID)
		if err != nil {
			return domain.NetworkPage[domain.NetworkWork]{}, err
		}
	}
	return domain.NetworkPage[domain.NetworkWork]{Items: works, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *BuilderNetworkRepository) GetPublicWorkBySlug(slug string) (domain.NetworkWork, error) {
	work, err := r.scanWork(r.db.QueryRow(`
		SELECT w.id, w.profile_id, w.slug, w.title, w.summary, w.body, w.cover_url, w.demo_url, w.repo_url,
		       w.tags_json, w.status, w.created_at, w.updated_at, p.display_name
		FROM network_works w JOIN network_profiles p ON p.id = w.profile_id
		WHERE w.slug = ? AND w.status = 'published' AND p.status = 'published'
	`, slug))
	if err != nil {
		return domain.NetworkWork{}, err
	}
	work.Members, err = r.listWorkMembers(work.ID)
	return work, err
}

func (r *BuilderNetworkRepository) ListWorksByProfile(profileID int64) ([]domain.NetworkWork, error) {
	rows, err := r.db.Query(`
		SELECT w.id, w.profile_id, w.slug, w.title, w.summary, w.body, w.cover_url, w.demo_url, w.repo_url,
		       w.tags_json, w.status, w.created_at, w.updated_at, p.display_name
		FROM network_works w JOIN network_profiles p ON p.id = w.profile_id
		WHERE w.profile_id = ? ORDER BY w.created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	works := make([]domain.NetworkWork, 0)
	for rows.Next() {
		var work domain.NetworkWork
		var tagsJSON string
		if err := rows.Scan(&work.ID, &work.ProfileID, &work.Slug, &work.Title, &work.Summary, &work.Body,
			&work.CoverURL, &work.DemoURL, &work.RepoURL, &tagsJSON, &work.Status, &work.CreatedAt,
			&work.UpdatedAt, &work.AuthorName); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &work.Tags)
		works = append(works, work)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range works {
		members, err := r.listWorkMembers(works[i].ID)
		if err != nil {
			return nil, err
		}
		works[i].Members = members
	}
	return works, nil
}

func (r *BuilderNetworkRepository) UpdateWork(profileID, workID int64, work domain.NetworkWork) (domain.NetworkWork, error) {
	tagsJSON, _ := json.Marshal(work.Tags)
	tx, err := r.db.Begin()
	if err != nil {
		return domain.NetworkWork{}, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE network_works
		SET title = ?, summary = ?, body = ?, cover_url = ?, demo_url = ?, repo_url = ?,
		    tags_json = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ?
	`, work.Title, work.Summary, work.Body, work.CoverURL, work.DemoURL, work.RepoURL,
		string(tagsJSON), work.Status, workID, profileID)
	if err != nil {
		return domain.NetworkWork{}, err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return domain.NetworkWork{}, err
	}
	if updated == 0 {
		return domain.NetworkWork{}, ErrNetworkNotFound
	}
	if _, err := tx.Exec(`DELETE FROM network_work_members WHERE work_id = ?`, workID); err != nil {
		return domain.NetworkWork{}, err
	}
	if err := replaceWorkMembers(tx, workID, profileID, work.Members); err != nil {
		return domain.NetworkWork{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.NetworkWork{}, err
	}
	return r.getWorkByID(workID)
}

func (r *BuilderNetworkRepository) GetPublicProfileBySlug(slug string) (domain.NetworkProfile, error) {
	var profile domain.NetworkProfile
	var skillsJSON, modesJSON string
	err := r.db.QueryRow(`
		SELECT id, slug, display_name, primary_role, headline, bio, city,
		       skills_json, collaboration_modes_json, availability, available_from, created_at, updated_at
		FROM network_profiles WHERE slug = ? AND status = 'published'
	`, slug).Scan(&profile.ID, &profile.Slug, &profile.DisplayName, &profile.PrimaryRole, &profile.Headline,
		&profile.Bio, &profile.City, &skillsJSON, &modesJSON, &profile.Availability, &profile.AvailableFrom,
		&profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkProfile{}, ErrNetworkNotFound
	}
	if err != nil {
		return domain.NetworkProfile{}, err
	}
	_ = json.Unmarshal([]byte(skillsJSON), &profile.Skills)
	_ = json.Unmarshal([]byte(modesJSON), &profile.Collaboration)
	return profile, nil
}

func (r *BuilderNetworkRepository) CreateOpportunity(profileID int64, opportunity domain.NetworkOpportunity) (domain.NetworkOpportunity, error) {
	rolesJSON, _ := json.Marshal(opportunity.LookingFor)
	skillsJSON, _ := json.Marshal(opportunity.RequiredSkills)
	result, err := r.db.Exec(`
		INSERT INTO network_opportunities
			(profile_id, slug, title, summary, opportunity_type, looking_for_roles_json,
			 required_skills_json, engagement_mode, budget_hint, duration_hint, city_scope,
			 remote_allowed, status, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''))
	`, profileID, opportunity.Slug, opportunity.Title, opportunity.Summary, opportunity.OpportunityType,
		string(rolesJSON), string(skillsJSON), opportunity.EngagementMode, opportunity.BudgetHint,
		opportunity.DurationHint, opportunity.CityScope, boolInt(opportunity.RemoteAllowed), opportunity.Status, opportunity.ExpiresAt)
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	opportunity.ID, err = result.LastInsertId()
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	opportunity.ProfileID = profileID
	return opportunity, nil
}

func (r *BuilderNetworkRepository) ListOpportunitiesByProfile(profileID int64) ([]domain.NetworkOpportunity, error) {
	rows, err := r.db.Query(`
		SELECT o.id, o.profile_id, o.slug, o.title, o.summary, o.opportunity_type,
		       o.looking_for_roles_json, o.required_skills_json, o.engagement_mode, o.budget_hint,
		       o.duration_hint, o.city_scope, o.remote_allowed, o.status, COALESCE(o.expires_at, ''),
		       o.created_at, o.updated_at, p.display_name
		FROM network_opportunities o JOIN network_profiles p ON p.id = o.profile_id
		WHERE o.profile_id = ?
		ORDER BY o.created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	opportunities := make([]domain.NetworkOpportunity, 0)
	for rows.Next() {
		opportunity, err := scanOpportunity(rows)
		if err != nil {
			return nil, err
		}
		opportunities = append(opportunities, opportunity)
	}
	return opportunities, rows.Err()
}

func (r *BuilderNetworkRepository) UpdateOpportunity(profileID, opportunityID int64, opportunity domain.NetworkOpportunity) (domain.NetworkOpportunity, error) {
	rolesJSON, _ := json.Marshal(opportunity.LookingFor)
	skillsJSON, _ := json.Marshal(opportunity.RequiredSkills)
	result, err := r.db.Exec(`
		UPDATE network_opportunities
		SET title = ?, summary = ?, opportunity_type = ?, looking_for_roles_json = ?,
		    required_skills_json = ?, engagement_mode = ?, budget_hint = ?, duration_hint = ?,
		    city_scope = ?, remote_allowed = ?, status = ?, expires_at = NULLIF(?, ''),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ?
	`, opportunity.Title, opportunity.Summary, opportunity.OpportunityType, string(rolesJSON),
		string(skillsJSON), opportunity.EngagementMode, opportunity.BudgetHint, opportunity.DurationHint,
		opportunity.CityScope, boolInt(opportunity.RemoteAllowed), opportunity.Status, opportunity.ExpiresAt,
		opportunityID, profileID)
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	if updated == 0 {
		return domain.NetworkOpportunity{}, ErrNetworkNotFound
	}
	return r.getOpportunityByID(opportunityID)
}

func (r *BuilderNetworkRepository) GetPublicOpportunityBySlug(slug string) (domain.NetworkOpportunity, error) {
	return scanOpportunity(r.db.QueryRow(`
		SELECT o.id, o.profile_id, o.slug, o.title, o.summary, o.opportunity_type,
		       o.looking_for_roles_json, o.required_skills_json, o.engagement_mode, o.budget_hint,
		       o.duration_hint, o.city_scope, o.remote_allowed, o.status, COALESCE(o.expires_at, ''),
		       o.created_at, o.updated_at, p.display_name
		FROM network_opportunities o JOIN network_profiles p ON p.id = o.profile_id
		WHERE o.slug = ? AND o.status = 'open' AND p.status = 'published'
		  AND (o.expires_at IS NULL OR o.expires_at > CURRENT_TIMESTAMP)
	`, slug))
}

func (r *BuilderNetworkRepository) ListPublicOpportunities() ([]domain.NetworkOpportunity, error) {
	rows, err := r.db.Query(`
		SELECT o.id, o.profile_id, o.slug, o.title, o.summary, o.opportunity_type,
		       o.looking_for_roles_json, o.required_skills_json, o.engagement_mode, o.budget_hint,
		       o.duration_hint, o.city_scope, o.remote_allowed, o.status, COALESCE(o.expires_at, ''),
		       o.created_at, o.updated_at, p.display_name
		FROM network_opportunities o JOIN network_profiles p ON p.id = o.profile_id
		WHERE o.status = 'open' AND p.status = 'published'
		  AND (o.expires_at IS NULL OR o.expires_at > CURRENT_TIMESTAMP)
		ORDER BY o.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	opportunities := make([]domain.NetworkOpportunity, 0)
	for rows.Next() {
		opportunity, err := scanOpportunity(rows)
		if err != nil {
			return nil, err
		}
		opportunities = append(opportunities, opportunity)
	}
	return opportunities, rows.Err()
}

func (r *BuilderNetworkRepository) BrowsePublicOpportunities(query domain.NetworkDiscoveryQuery) (domain.NetworkPage[domain.NetworkOpportunity], error) {
	page, pageSize, offset := normalizePage(query.Page, query.PageSize)
	conditions := []string{
		"o.status = 'open'",
		"p.status = 'published'",
		"(o.expires_at IS NULL OR o.expires_at > CURRENT_TIMESTAMP)",
	}
	args := make([]any, 0)
	addNetworkLike(&conditions, &args, []string{"o.title", "o.summary"}, query.Query)
	addNetworkLike(&conditions, &args, []string{"o.opportunity_type"}, query.Type)
	addNetworkLike(&conditions, &args, []string{"o.looking_for_roles_json", "p.primary_role"}, query.Role)
	addNetworkLike(&conditions, &args, []string{"o.required_skills_json"}, query.Skill)
	addNetworkLike(&conditions, &args, []string{"o.city_scope", "p.city"}, query.City)
	if query.Remote != nil {
		conditions = append(conditions, "o.remote_allowed = ?")
		if *query.Remote {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	where := strings.Join(conditions, " AND ")
	join := " FROM network_opportunities o JOIN network_profiles p ON p.id = o.profile_id WHERE " + where

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*)`+join, args...).Scan(&total); err != nil {
		return domain.NetworkPage[domain.NetworkOpportunity]{}, err
	}
	rows, err := r.db.Query(`
		SELECT o.id, o.profile_id, o.slug, o.title, o.summary, o.opportunity_type,
		       o.looking_for_roles_json, o.required_skills_json, o.engagement_mode, o.budget_hint,
		       o.duration_hint, o.city_scope, o.remote_allowed, o.status, COALESCE(o.expires_at, ''),
		       o.created_at, o.updated_at, p.display_name`+join+`
		ORDER BY o.created_at DESC, o.id DESC LIMIT ? OFFSET ?
	`, append(args, pageSize, offset)...)
	if err != nil {
		return domain.NetworkPage[domain.NetworkOpportunity]{}, err
	}
	defer rows.Close()
	opportunities := make([]domain.NetworkOpportunity, 0)
	for rows.Next() {
		opportunity, err := scanOpportunity(rows)
		if err != nil {
			return domain.NetworkPage[domain.NetworkOpportunity]{}, err
		}
		opportunities = append(opportunities, opportunity)
	}
	if err := rows.Err(); err != nil {
		return domain.NetworkPage[domain.NetworkOpportunity]{}, err
	}
	return domain.NetworkPage[domain.NetworkOpportunity]{Items: opportunities, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *BuilderNetworkRepository) getWorkByID(workID int64) (domain.NetworkWork, error) {
	work, err := r.scanWork(r.db.QueryRow(`
		SELECT w.id, w.profile_id, w.slug, w.title, w.summary, w.body, w.cover_url, w.demo_url, w.repo_url,
		       w.tags_json, w.status, w.created_at, w.updated_at, p.display_name
		FROM network_works w JOIN network_profiles p ON p.id = w.profile_id
		WHERE w.id = ?
	`, workID))
	if err != nil {
		return domain.NetworkWork{}, err
	}
	work.Members, err = r.listWorkMembers(work.ID)
	return work, err
}

func (r *BuilderNetworkRepository) scanWork(row interface{ Scan(...any) error }) (domain.NetworkWork, error) {
	var work domain.NetworkWork
	var tagsJSON string
	err := row.Scan(&work.ID, &work.ProfileID, &work.Slug, &work.Title, &work.Summary, &work.Body,
		&work.CoverURL, &work.DemoURL, &work.RepoURL, &tagsJSON, &work.Status, &work.CreatedAt,
		&work.UpdatedAt, &work.AuthorName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkWork{}, ErrNetworkNotFound
	}
	if err != nil {
		return domain.NetworkWork{}, err
	}
	_ = json.Unmarshal([]byte(tagsJSON), &work.Tags)
	return work, nil
}

func (r *BuilderNetworkRepository) listWorkMembers(workID int64) ([]domain.NetworkWorkMember, error) {
	rows, err := r.db.Query(`
		SELECT m.profile_id, p.display_name, m.role, m.contribution, m.sort_order, m.is_owner
		FROM network_work_members m
		JOIN network_profiles p ON p.id = m.profile_id
		WHERE m.work_id = ?
		ORDER BY m.sort_order, m.is_owner DESC, m.profile_id
	`, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := make([]domain.NetworkWorkMember, 0)
	for rows.Next() {
		var member domain.NetworkWorkMember
		var isOwner int
		if err := rows.Scan(&member.ProfileID, &member.DisplayName, &member.Role, &member.Contribution,
			&member.SortOrder, &isOwner); err != nil {
			return nil, err
		}
		member.IsOwner = isOwner != 0
		members = append(members, member)
	}
	return members, rows.Err()
}

func replaceWorkMembers(tx *sql.Tx, workID, ownerProfileID int64, members []domain.NetworkWorkMember) error {
	if _, err := tx.Exec(`
		INSERT INTO network_work_members (work_id, profile_id, role, contribution, sort_order, is_owner)
		VALUES (?, ?, 'Owner', '', 0, 1)
	`, workID, ownerProfileID); err != nil {
		return err
	}
	for _, member := range members {
		if member.ProfileID == ownerProfileID {
			continue
		}
		var exists int
		if err := tx.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM network_profiles WHERE id = ? AND status = 'published')
		`, member.ProfileID).Scan(&exists); err != nil {
			return err
		}
		if exists != 1 {
			return ErrNetworkNotFound
		}
		if _, err := tx.Exec(`
			INSERT INTO network_work_members (work_id, profile_id, role, contribution, sort_order, is_owner)
			VALUES (?, ?, ?, ?, ?, 0)
		`, workID, member.ProfileID, member.Role, member.Contribution, member.SortOrder); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				return ErrNetworkConflict
			}
			return err
		}
	}
	return nil
}

func scanOpportunity(row interface{ Scan(...any) error }) (domain.NetworkOpportunity, error) {
	var opportunity domain.NetworkOpportunity
	var rolesJSON, skillsJSON string
	var remote int
	err := row.Scan(&opportunity.ID, &opportunity.ProfileID, &opportunity.Slug, &opportunity.Title,
		&opportunity.Summary, &opportunity.OpportunityType, &rolesJSON, &skillsJSON,
		&opportunity.EngagementMode, &opportunity.BudgetHint, &opportunity.DurationHint,
		&opportunity.CityScope, &remote, &opportunity.Status, &opportunity.ExpiresAt,
		&opportunity.CreatedAt, &opportunity.UpdatedAt, &opportunity.OwnerName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkOpportunity{}, ErrNetworkNotFound
	}
	if err != nil {
		return domain.NetworkOpportunity{}, err
	}
	opportunity.RemoteAllowed = remote != 0
	_ = json.Unmarshal([]byte(rolesJSON), &opportunity.LookingFor)
	_ = json.Unmarshal([]byte(skillsJSON), &opportunity.RequiredSkills)
	return opportunity, nil
}

func (r *BuilderNetworkRepository) getOpportunityByID(opportunityID int64) (domain.NetworkOpportunity, error) {
	return scanOpportunity(r.db.QueryRow(`
		SELECT o.id, o.profile_id, o.slug, o.title, o.summary, o.opportunity_type,
		       o.looking_for_roles_json, o.required_skills_json, o.engagement_mode, o.budget_hint,
		       o.duration_hint, o.city_scope, o.remote_allowed, o.status, COALESCE(o.expires_at, ''),
		       o.created_at, o.updated_at, p.display_name
		FROM network_opportunities o JOIN network_profiles p ON p.id = o.profile_id
		WHERE o.id = ?
	`, opportunityID))
}

func (r *BuilderNetworkRepository) CreateContactRequest(fromProfileID int64, request domain.NetworkContactRequest) (domain.NetworkContactRequest, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return domain.NetworkContactRequest{}, err
	}
	defer tx.Rollback()

	var targetProfileID int64
	switch request.TargetType {
	case "profile":
		err := tx.QueryRow(`SELECT id FROM network_profiles WHERE id = ? AND status = 'published'`, request.TargetID).Scan(&targetProfileID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NetworkContactRequest{}, ErrNetworkNotFound
		}
		if err != nil {
			return domain.NetworkContactRequest{}, err
		}
	case "opportunity":
		err := tx.QueryRow(`
			SELECT o.profile_id FROM network_opportunities o
			JOIN network_profiles p ON p.id = o.profile_id
			WHERE o.id = ? AND o.status = 'open' AND p.status = 'published'
			  AND (o.expires_at IS NULL OR o.expires_at > CURRENT_TIMESTAMP)
		`, request.TargetID).Scan(&targetProfileID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NetworkContactRequest{}, ErrNetworkNotFound
		}
		if err != nil {
			return domain.NetworkContactRequest{}, err
		}
	default:
		return domain.NetworkContactRequest{}, errors.New("target_type must be profile or opportunity")
	}
	if targetProfileID == fromProfileID {
		return domain.NetworkContactRequest{}, ErrNetworkConflict
	}

	seenWorks := make(map[int64]struct{}, len(request.RelatedWorkIDs))
	for _, workID := range request.RelatedWorkIDs {
		if workID <= 0 {
			return domain.NetworkContactRequest{}, ErrNetworkNotFound
		}
		if _, exists := seenWorks[workID]; exists {
			return domain.NetworkContactRequest{}, ErrNetworkConflict
		}
		seenWorks[workID] = struct{}{}
		var count int
		if err := tx.QueryRow(`
			SELECT COUNT(1) FROM network_works
			WHERE id = ? AND profile_id = ? AND status = 'published'
		`, workID, fromProfileID).Scan(&count); err != nil {
			return domain.NetworkContactRequest{}, err
		}
		if count != 1 {
			return domain.NetworkContactRequest{}, ErrNetworkNotFound
		}
	}

	workIDsJSON, _ := json.Marshal(request.RelatedWorkIDs)
	result, err := tx.Exec(`
		INSERT INTO network_contact_requests
			(from_profile_id, target_type, target_id, target_profile_id, message, related_work_ids_json,
			 share_email, share_wechat, share_phone)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, fromProfileID, request.TargetType, request.TargetID, targetProfileID, request.Message,
		string(workIDsJSON), boolInt(request.ShareEmail), boolInt(request.ShareWeChat), boolInt(request.SharePhone))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NetworkContactRequest{}, ErrNetworkConflict
		}
		return domain.NetworkContactRequest{}, err
	}
	request.ID, err = result.LastInsertId()
	if err != nil {
		return domain.NetworkContactRequest{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.NetworkContactRequest{}, err
	}
	request.FromProfileID = fromProfileID
	request.Status = "pending"
	return request, nil
}

func (r *BuilderNetworkRepository) ListContactRequests(accountID int64, incoming bool) ([]domain.NetworkContactRequest, error) {
	profileID, err := r.profileIDByAccount(accountID)
	if err != nil {
		return nil, err
	}
	query := `
		SELECT r.id, r.from_profile_id, r.target_type, r.target_id, r.message, r.related_work_ids_json,
		       r.share_email, r.share_wechat, r.share_phone, r.status, r.progress_status,
		       CASE WHEN x.revoked_at IS NOT NULL THEN 1 ELSE 0 END, r.created_at, r.updated_at,
		       p.display_name, p.primary_role
		FROM network_contact_requests r
		JOIN network_profiles p ON p.id = r.from_profile_id
		LEFT JOIN network_contact_exchanges x ON x.request_id = r.id
		WHERE r.target_profile_id = ?
		ORDER BY r.created_at DESC`
	if !incoming {
		query = `
			SELECT r.id, r.from_profile_id, r.target_type, r.target_id, r.message, r.related_work_ids_json,
			       r.share_email, r.share_wechat, r.share_phone, r.status, r.progress_status,
			       CASE WHEN x.revoked_at IS NOT NULL THEN 1 ELSE 0 END, r.created_at, r.updated_at,
			       p.display_name, p.primary_role
			FROM network_contact_requests r
			JOIN network_profiles p ON p.id = r.from_profile_id
			LEFT JOIN network_contact_exchanges x ON x.request_id = r.id
			WHERE r.from_profile_id = ?
			ORDER BY r.created_at DESC`
	}
	rows, err := r.db.Query(query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]domain.NetworkContactRequest, 0)
	for rows.Next() {
		var request domain.NetworkContactRequest
		var workIDsJSON string
		var shareEmail, shareWeChat, sharePhone, exchangeRevoked int
		if err := rows.Scan(&request.ID, &request.FromProfileID, &request.TargetType, &request.TargetID,
			&request.Message, &workIDsJSON, &shareEmail, &shareWeChat, &sharePhone, &request.Status,
			&request.ProgressStatus, &exchangeRevoked, &request.CreatedAt, &request.UpdatedAt,
			&request.ApplicantName, &request.ApplicantRole); err != nil {
			return nil, err
		}
		request.ShareEmail = shareEmail != 0
		request.ShareWeChat = shareWeChat != 0
		request.SharePhone = sharePhone != 0
		request.ExchangeRevoked = exchangeRevoked != 0
		_ = json.Unmarshal([]byte(workIDsJSON), &request.RelatedWorkIDs)
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *BuilderNetworkRepository) UpdateContactRequestStatus(requestID, accountID int64, status string) error {
	profileID, err := r.profileIDByAccount(accountID)
	if err != nil {
		return err
	}
	if status == "accepted" {
		tx, err := r.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var applicantProfileID int64
		var requestEmail, requestWeChat, requestPhone int
		var applicantEmail, applicantWeChat, applicantPhone int
		var recipientEmail, recipientWeChat, recipientPhone int
		err = tx.QueryRow(`
			SELECT r.from_profile_id, r.share_email, r.share_wechat, r.share_phone,
			       COALESCE(a.share_email, 0), COALESCE(a.share_wechat, 0), COALESCE(a.share_phone, 0),
			       COALESCE(t.share_email, 0), COALESCE(t.share_wechat, 0), COALESCE(t.share_phone, 0)
			FROM network_contact_requests r
			LEFT JOIN network_private_contacts a ON a.profile_id = r.from_profile_id
			LEFT JOIN network_private_contacts t ON t.profile_id = r.target_profile_id
			WHERE r.id = ? AND r.target_profile_id = ? AND r.status = 'pending'
		`, requestID, profileID).Scan(
			&applicantProfileID, &requestEmail, &requestWeChat, &requestPhone,
			&applicantEmail, &applicantWeChat, &applicantPhone,
			&recipientEmail, &recipientWeChat, &recipientPhone,
		)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNetworkNotFound
		}
		if err != nil {
			return err
		}
		result, err := tx.Exec(`
			UPDATE network_contact_requests SET status = 'accepted', updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND target_profile_id = ? AND status = 'pending'
		`, requestID, profileID)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return ErrNetworkNotFound
		}
		if _, err := tx.Exec(`
			INSERT INTO network_contact_exchanges (
				request_id, accepted_by_profile_id,
				applicant_share_email, applicant_share_wechat, applicant_share_phone,
				recipient_share_email, recipient_share_wechat, recipient_share_phone
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, requestID, profileID,
			boolInt(requestEmail != 0 && applicantEmail != 0),
			boolInt(requestWeChat != 0 && applicantWeChat != 0),
			boolInt(requestPhone != 0 && applicantPhone != 0),
			recipientEmail, recipientWeChat, recipientPhone); err != nil {
			return err
		}
		return tx.Commit()
	}
	var result sql.Result
	if status == "declined" {
		result, err = r.db.Exec(`
			UPDATE network_contact_requests SET status = 'declined', updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND target_profile_id = ? AND status = 'pending'
		`, requestID, profileID)
	} else {
		result, err = r.db.Exec(`
			UPDATE network_contact_requests SET status = 'withdrawn', updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND from_profile_id = ? AND status = 'pending'
		`, requestID, profileID)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrNetworkNotFound
	}
	return nil
}

func (r *BuilderNetworkRepository) UpdateContactRequestProgress(requestID, accountID int64, progress string) error {
	profileID, err := r.profileIDByAccount(accountID)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(`
		UPDATE network_contact_requests
		SET progress_status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'accepted'
		  AND (from_profile_id = ? OR target_profile_id = ?)
	`, progress, requestID, profileID, profileID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrNetworkNotFound
	}
	return nil
}

func (r *BuilderNetworkRepository) GetContactExchange(requestID, accountID int64) (domain.NetworkContactExchange, error) {
	viewerProfileID, err := r.profileIDByAccount(accountID)
	if err != nil {
		return domain.NetworkContactExchange{}, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return domain.NetworkContactExchange{}, err
	}
	defer tx.Rollback()
	var applicantProfileID, recipientProfileID int64
	var applicantEmail, applicantWeChat, applicantPhone int
	var recipientEmail, recipientWeChat, recipientPhone int
	var exchange domain.NetworkContactExchange
	err = tx.QueryRow(`
		SELECT r.from_profile_id, r.target_profile_id,
		       x.applicant_share_email, x.applicant_share_wechat, x.applicant_share_phone,
		       x.recipient_share_email, x.recipient_share_wechat, x.recipient_share_phone, x.revealed_at
		FROM network_contact_requests r
		JOIN network_contact_exchanges x ON x.request_id = r.id AND x.revoked_at IS NULL
		WHERE r.id = ? AND r.status = 'accepted'
		  AND (r.from_profile_id = ? OR r.target_profile_id = ?)
	`, requestID, viewerProfileID, viewerProfileID).Scan(
		&applicantProfileID, &recipientProfileID,
		&applicantEmail, &applicantWeChat, &applicantPhone,
		&recipientEmail, &recipientWeChat, &recipientPhone, &exchange.RevealedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.NetworkContactExchange{}, ErrNetworkNotFound
	}
	if err != nil {
		return domain.NetworkContactExchange{}, err
	}
	exchange.RequestID = requestID
	exchange.Applicant, err = readExchangeContact(tx, applicantProfileID, applicantEmail != 0, applicantWeChat != 0, applicantPhone != 0)
	if err != nil {
		return domain.NetworkContactExchange{}, err
	}
	exchange.Recipient, err = readExchangeContact(tx, recipientProfileID, recipientEmail != 0, recipientWeChat != 0, recipientPhone != 0)
	if err != nil {
		return domain.NetworkContactExchange{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.NetworkContactExchange{}, err
	}
	return exchange, nil
}

func (r *BuilderNetworkRepository) RevokeContactExchange(requestID, accountID int64) error {
	profileID, err := r.profileIDByAccount(accountID)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(`
		UPDATE network_contact_exchanges
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE request_id = ? AND revoked_at IS NULL
		  AND EXISTS (
			SELECT 1 FROM network_contact_requests r
			WHERE r.id = network_contact_exchanges.request_id
			  AND r.status = 'accepted'
			  AND (r.from_profile_id = ? OR r.target_profile_id = ?)
		  )
	`, requestID, profileID, profileID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return ErrNetworkNotFound
	}
	return nil
}

type networkQueryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

func readExchangeContact(queryer networkQueryRower, profileID int64, requestEmail, requestWeChat, requestPhone bool) (domain.NetworkContactDetails, error) {
	var details domain.NetworkContactDetails
	var email, wechat, phone string
	var shareEmail, shareWeChat, sharePhone int
	err := queryer.QueryRow(`
		SELECT email, wechat, phone, share_email, share_wechat, share_phone
		FROM network_private_contacts WHERE profile_id = ?
	`, profileID).Scan(&email, &wechat, &phone, &shareEmail, &shareWeChat, &sharePhone)
	if errors.Is(err, sql.ErrNoRows) {
		return details, nil
	}
	if err != nil {
		return details, err
	}
	if requestEmail && shareEmail != 0 {
		details.Email = email
	}
	if requestWeChat && shareWeChat != 0 {
		details.WeChat = wechat
	}
	if requestPhone && sharePhone != 0 {
		details.Phone = phone
	}
	return details, nil
}

func (r *BuilderNetworkRepository) profileIDByAccount(accountID int64) (int64, error) {
	var profileID int64
	err := r.db.QueryRow(`SELECT id FROM network_profiles WHERE account_id = ?`, accountID).Scan(&profileID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNetworkNotFound
	}
	return profileID, err
}

func profileSlug(name string, accountID int64) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "builder"
	}
	return slug + "-" + strconvFormatInt(accountID)
}

func strconvFormatInt(value int64) string {
	return fmt.Sprintf("%d", value)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizePage(page, pageSize int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if page > 1_000_000 {
		page = 1_000_000
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, (page - 1) * pageSize
}

func addNetworkLike(conditions *[]string, args *[]any, columns []string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	escaped := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value)
	terms := make([]string, 0, len(columns))
	for _, column := range columns {
		terms = append(terms, column+" LIKE ? ESCAPE '!'")
		*args = append(*args, "%"+escaped+"%")
	}
	*conditions = append(*conditions, "("+strings.Join(terms, " OR ")+")")
}
