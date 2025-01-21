package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wafi04/common/pkg/logger"
	pb "github.com/wafi04/go-testing/auth/grpc"
	middleware "github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/shared/pkg/mailer"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	DB     *sql.DB
	logger logger.Logger
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *pb.CreateUserRequest) (pb.CreateUserResponse, error) {
	role := "user"
	if user.Email == "wafiq610@gmail.com" {
		role = "admin"
	}
	userID := uuid.New().String()
	now := time.Now()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Log(logger.ErrorLevel, "Failed Trasacation : %v", err)
		return pb.CreateUserResponse{}, nil
	}
	defer tx.Rollback()

	query := `
        INSERT INTO users (
            user_id, name, email, password_hash, role,
            is_active, is_email_verified, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err = tx.ExecContext(
		ctx, query,
		userID, user.Name, user.Email, user.Password, role,
		true, false, now, now,
	)
	// appPW := configs.GetEnv("APP_PASSWORD",pb.LoginResponse{})
	// cleanPassword := strings.ReplaceAll(appPW, " ", "")
	// emailSender := mailer.NewEmailSender(
	// 	"smtp.gmail.com",
	// 	587,
	// 	configs.GetEnv("APP_EMAIL"),
	// 	cleanPassword,
	// )

	// toEmail := user.Email
	verificationCode := mailer.GenerateVerificationCode()

	// if err := emailSender.SendVerificationEmail(toEmail, user.Name, verificationCode); err != nil {
	// 	return pb.CreateUserResponse{}, fmt.Errorf("failed to send email : %w", err)
	// }

	if err != nil {
		r.logger.Log(logger.ErrorLevel, "Failed to insert user : %v", err)
		return pb.CreateUserResponse{}, nil
	}
	expiresAt := now.Add(24 * time.Hour)

	query = `
        INSERT INTO verification_tokens (
             user_id, token, type, expires_at,is_used
        ) VALUES ($1, $2, $3, $4, $5)
    `

	_, err = tx.ExecContext(
		ctx, query,
		userID, verificationCode, "EMAIL_VERIFICATION", expiresAt,true,
	)

	if err != nil {
		return pb.CreateUserResponse{}, fmt.Errorf("failed to create verification token: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return pb.CreateUserResponse{}, err
	}

	return pb.CreateUserResponse{
		UserId: userID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   role,
	}, nil

}

type dbUser struct {
	UserID          string
	Name            string
	Email           string
	Role            string
	Password        string
	Picture         string
	IsEmailVerified bool
	CreatedAt       int64
	UpdatedAt       int64
	LastLoginAt     int64
	IsActive        bool
}

func (r *UserRepository) Login(ctx context.Context, login *pb.LoginRequest) (*pb.LoginResponse, error) {

	query := `
    SELECT
        user_id,
        name,
        email,
        role,
        password_hash,
        COALESCE(picture, ''),
        COALESCE(is_email_verified, false)::boolean,  
        EXTRACT(EPOCH FROM created_at)::bigint,
        EXTRACT(EPOCH FROM updated_at)::bigint,
        EXTRACT(EPOCH FROM COALESCE(last_login_at, created_at))::bigint,
        is_active::boolean
    FROM users
    WHERE email = $1
`

	var dbuser dbUser

	err := r.DB.QueryRowContext(ctx, query, login.Email).Scan(
		&dbuser.UserID,
		&dbuser.Name,
		&dbuser.Email,
		&dbuser.Role,
		&dbuser.Password,
		&dbuser.Picture,
		&dbuser.IsEmailVerified,
		&dbuser.CreatedAt,
		&dbuser.UpdatedAt,
		&dbuser.LastLoginAt,
		&dbuser.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	userInfo := &pb.UserInfo{
		UserId:          dbuser.UserID,
		Name:            dbuser.Name,
		Email:           dbuser.Email,
		Role:            dbuser.Role,
		IsEmailVerified: dbuser.IsEmailVerified,
		CreatedAt:       dbuser.CreatedAt,
		UpdatedAt:       dbuser.UpdatedAt,
		LastLoginAt:     dbuser.LastLoginAt,
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbuser.Password), []byte(login.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := middleware.GenerateToken(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := pb.Session{
		SessionId:      uuid.New().String(),
		UserId:         userInfo.UserId,
		AccessToken:    token,
		RefreshToken:   token,
		IpAddress:      login.IpAddress,
		DeviceInfo:     login.DeviceInfo,
		CreatedAt:      1,
		LastActivityAt: 2,
		IsActive:       true,
		ExpiresAt:      time.Now().Unix(),
	}

	// Store session in database
	err = r.CreateSession(ctx, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	_, err = r.DB.ExecContext(
		ctx,
		"UPDATE users SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1",
		userInfo.UserId,
	)
	if err != nil {
		r.logger.Log(logger.ErrorLevel, "Failed to update last login: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken: token,
		UserId:      userInfo.UserId,
		SessionInfo: &pb.SessionInfo{
			SessionId:      session.SessionId,
			DeviceInfo:     session.DeviceInfo,
			IpAddress:      session.IpAddress,
			CreatedAt:      time.Now().Unix(),
			LastActivityAt: time.Now().Unix(),
		},
	}, nil
}
func (sr *UserRepository) CreateSession(ctx context.Context, session *pb.Session) error {
	query := `
		INSERT INTO sessions (
			session_id,
			user_id,
			access_token,
			refresh_token,
			ip_address,
			device_info,
			is_active,
			expires_at,
			last_activity_at,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	// Generate UUID if not provided
	if session.SessionId == "" {
		session.SessionId = uuid.New().String()
	}

	// Set current timestamp
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	sr.logger.WithFields(map[string]interface{}{
		"user_id":    session.UserId,
		"session_id": session.SessionId,
		"expires_at": expiresAt,
	}).Debug("Creating new session")

	_, err := sr.DB.ExecContext(
		ctx,
		query,
		session.SessionId,
		session.UserId,
		session.AccessToken,
		session.RefreshToken,
		session.IpAddress,
		session.DeviceInfo,
		true,          
		expiresAt,     
		now,           
		now,           
	)

	if err != nil {
		sr.logger.WithError(err).WithFields(map[string]interface{}{
			"user_id":    session.UserId,
			"session_id": session.SessionId,
		}).Error("Failed to create session")
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (sr *UserRepository) GetUser(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
	query := `
        SELECT 
            id, 
            name, 
            email,
			picture, 
            role, 
            is_active, 
            is_email_verified,
            created_at, 
            updated_at, 
            last_login_at
        FROM users
        WHERE id = $1
    `

	user := &pb.GetUserResponse{}
	var isActive bool
	err := sr.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.User.UserId,
		&user.User.Name,
		&user.User.Email,
		&user.User.Picture,
		&user.User.Role,
		&isActive,
		&user.User.IsEmailVerified,
		&user.User.CreatedAt,
		&user.User.UpdatedAt,
		&user.User.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}
