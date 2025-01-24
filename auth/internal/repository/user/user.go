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
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository struct {
	DB     *sql.DB
	logger logger.Logger
    // configManager *configs.ConfigManager
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
        // configManager: config,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (pb.CreateUserResponse, error) {
	
	role := "user"
	if req.Email == "wafiq610@gmail.com" {
		role = "admin"
	}
	userID := uuid.New().String()
	now := time.Now()
	query := `
        INSERT INTO users (
            user_id, name, email, password_hash, role,
            is_active, is_email_verified, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    _, err := r.DB.ExecContext(
        ctx, query,
        userID, req.Name, req.Email, req.Password, role,
        true, false, now, now,
    )

	if err != nil {
		return pb.CreateUserResponse{}, fmt.Errorf("failed to create verification token: %w", err)
	}

	token, err := middleware.GenerateToken(&pb.UserInfo{
		UserId: userID,
		Name: req.Name,
		Email: req.Email,
		Role: role,
		IsEmailVerified: false,

	})
	if err != nil {
		return pb.CreateUserResponse{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := pb.Session{
		SessionId:      uuid.New().String(),
		UserId:         userID,
		AccessToken:    token,
		RefreshToken:   token,
		IpAddress:      req.IpAddress,
		DeviceInfo:     req.DeviceInfo,
		CreatedAt:      time.Now().Unix(),
		LastActivityAt: time.Now().Unix(),
		IsActive:       true,
		ExpiresAt:      time.Now().Unix(),
	}

	err = r.CreateSession(ctx, &session)
	if err != nil {
		return pb.CreateUserResponse{}, fmt.Errorf("failed to create session: %w", err)
	}

	return pb.CreateUserResponse{
    UserId: userID, 
    Name:   req.Name,
    Email:  req.Email,
    Role:   role,
    Picture: req.Picture,
	AccessToken: token,
	SessionInfo: &pb.Session{
		SessionId: session.SessionId,
		DeviceInfo: session.DeviceInfo,
		IpAddress: session.IpAddress,
	},
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


	 query = `
        SELECT 
            session_id, 
            ip_address,
            device_info, 
            EXTRACT(EPOCH FROM created_at)::bigint, 
          	EXTRACT(EPOCH FROM last_activity_at)::bigint
        FROM sessions 
        WHERE user_id = $1 AND is_active = true AND device_info = $2
    `

	var existingSession pb.Session
    err = r.DB.QueryRowContext(ctx, query, userInfo.UserId,login.DeviceInfo).Scan(
        &existingSession.SessionId,
        &existingSession.IpAddress,
        &existingSession.DeviceInfo,
        &existingSession.CreatedAt,
        &existingSession.LastActivityAt,
    )

    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("error checking existing session: %w", err)
    }

    if err == sql.ErrNoRows {
        existingSession = pb.Session{
            SessionId:      uuid.New().String(),
            UserId:         userInfo.UserId,
            AccessToken:    token,
            RefreshToken:   token,
            IpAddress:      login.IpAddress,
            DeviceInfo:     login.DeviceInfo,
            CreatedAt:      time.Now().Unix(),
            LastActivityAt: time.Now().Unix(),
            IsActive:       true,
            ExpiresAt:      time.Now().Unix(),
        }

        err = r.CreateSession(ctx, &existingSession)
        if err != nil {
            return nil, fmt.Errorf("failed to create session: %w", err)
        }
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
            SessionId:      existingSession.SessionId,
            DeviceInfo:     existingSession.DeviceInfo,
            IpAddress:      existingSession.IpAddress,
            CreatedAt:      existingSession.CreatedAt,
            LastActivityAt: existingSession.LastActivityAt,
        },
    }, nil
}

func (sr *UserRepository) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    query := `
        SELECT 
            user_id, 
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
        WHERE user_id = $1
    `
	
    user := &pb.GetUserResponse{
        User: &pb.UserInfo{},
    }

    var (
        isActive bool
        createdAt, updatedAt, lastLoginAt time.Time
        picture sql.NullString 
    )
    err := sr.DB.QueryRowContext(ctx, query, req.UserId).Scan(
        &user.User.UserId,
        &user.User.Name,
        &user.User.Email,
        &picture, 
        &user.User.Role,
        &isActive,
        &user.User.IsEmailVerified,
        &createdAt,
        &updatedAt,
        &lastLoginAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "user not found")
        }
        sr.logger.Log(logger.ErrorLevel, "Error fetching user: %v", err)
        return nil, status.Errorf(codes.Internal, "database error")
    }

	if picture.Valid {
        user.User.Picture = picture.String
    }
	
    user.User.CreatedAt = createdAt.Unix()
    user.User.UpdatedAt = updatedAt.Unix()
    user.User.LastLoginAt = lastLoginAt.Unix()
    return user, nil
}

func (sr   *UserRepository) Logout(ctx context.Context,req *pb.LogoutRequest) (*pb.LogoutResponse,error){
	query  := `
	DELETE FROM sessions
    WHERE access_token = $1 AND user_id = $2
	`
	_,err :=  sr.DB.ExecContext(ctx, query, req.AccessToken,req.UserId)

	if err != nil {
		if err == sql.ErrNoRows {
            return nil, status.Errorf(codes.NotFound, "user not found")
        }
        sr.logger.Log(logger.ErrorLevel, "Error fetching user: %v", err)
        return nil, status.Errorf(codes.Internal, "database error")
	}

	return &pb.LogoutResponse{
		Success: true,
	},nil
}
