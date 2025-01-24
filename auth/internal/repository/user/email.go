package user

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/wafi04/common/pkg/logger"
	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
	"golang.org/x/crypto/bcrypt"
)

func generateVerificationCode() string {
    return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *UserRepository)  ResendVerification(ctx context.Context,req  *pb.ResendVerificationRequest)  (*pb.ResendVerificationResponse,error)   {
    user,err :=   s.GetUser(ctx, &pb.GetUserRequest{
        UserId: req.UserId,
    })
    if err !=  nil {
        s.logger.Log(logger.ErrorLevel, "Unauthorized : %v",err)
        return nil,fmt.Errorf("Unauthorized : %v", err)
    }


    if err != nil {
        return nil, fmt.Errorf("failed to generate verification token: %w", err)
    }
    

    verifyCode := generateVerificationCode()
    expiresAt := time.Now().Add(1 * time.Hour)

    
    

    query := `
        INSERT INTO verification_tokens (
            token, 
            user_id, 
            verify_code, 
            token_type, 
            expires_at
        ) VALUES ($1, $2, $3, $4, $5)
    `

    _, err = s.DB.ExecContext(ctx, query, 
        req.Token, 
        req.UserId, 
        verifyCode, 
        "EMAIL_VERIFICATION", 
        expiresAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to  verification token: %w", err)
    }

    return &pb.ResendVerificationResponse{
        VerificationToken: req.Token,
        VerifyCode: verifyCode,
        Success: true,
        ExpiresAt: expiresAt.Unix(),
    }, nil
}

func (s *UserRepository) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
  query := `
    SELECT user_id, expires_at  
    FROM verification_tokens 
    WHERE token = $1 
    AND verify_code = $2 
    AND is_used = false 
    AND expires_at > NOW()
`
    var (
        userId     string
        expiresAt  time.Time
    )

    err := s.DB.QueryRowContext(ctx, query, req.VerificationToken, req.VerifyCode).Scan(&userId, &expiresAt)
    if err != nil {
        return nil, fmt.Errorf("verification failed: %w", err)
    }

    tx, err := s.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("transaction error: %w", err)
    }

    updateQuery := `
        UPDATE users 
        SET is_email_verified = true, updated_at = NOW() 
        WHERE user_id = $1
    `

    markUsedQuery := `
        UPDATE verification_tokens 
        SET is_used = true 
        WHERE token = $1
    `

  
    _, err = tx.ExecContext(ctx, updateQuery, userId)
    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("update user error: %w", err)
    }

    _, err = tx.ExecContext(ctx, markUsedQuery, req.VerificationToken)
    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("mark token error: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit error: %w", err)
    }

    return &pb.VerifyEmailResponse{
        Success: true,
        UserId:  userId,
        Message: "Email verified successfully",
    }, nil
}
func (s *UserRepository) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.RequestPasswordResetResponse, error) {
    var user pb.User
    checkUserQuery := `SELECT user_id,name,email,Role,is_email_verified,is_active  FROM users WHERE email = $1`
    err := s.DB.QueryRowContext(ctx, checkUserQuery, req.Email).Scan(&user.UserId,user.Name,user.Email,user.Role,user.IsEmailVerified,user.IsActive)
    if err != nil {
        if err == sql.ErrNoRows {
            return &pb.RequestPasswordResetResponse{
                Success: false,
            }, nil
        }
        return nil, fmt.Errorf("failed to check user: %v", err)
    }

    token ,err :=  middleware.GenerateToken(&pb.UserInfo{
        UserId: user.UserId,
        Name: user.Name,
        Email: user.Email,
        Role: user.Role,
        IsEmailVerified: user.IsEmailVerified,
    })

    if err != nil {
        return nil, err
    }

    insertTokenQuery := `
        INSERT INTO verification_tokens 
        (token, user_id, token_type, expires_at) 
        VALUES ($1, $2, 'PASSWORD', CURRENT_TIMESTAMP + INTERVAL '1 hour')`

    _, err = s.DB.ExecContext(ctx, insertTokenQuery, token, user.UserId)
    if err != nil {
        return nil, fmt.Errorf("failed to create reset token: %v", err)
    }

    return &pb.RequestPasswordResetResponse{
        Success:          true,
        ResetToken:       token,
        ExpiresAt:        time.Now().Add(1 * time.Hour).Unix(),
    }, nil
}

func (s *UserRepository) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
    tx, err := s.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    verifyTokenQuery := `
        SELECT user_id 
        FROM verification_tokens 
        WHERE token = $1 
        AND token_type = 'password_reset'
        AND expires_at > CURRENT_TIMESTAMP 
        AND is_used = false`

    var userID string
    err = tx.QueryRowContext(ctx, verifyTokenQuery, req.ResetToken).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return &pb.ResetPasswordResponse{
                Success: false,
                Message: "Invalid or expired reset token",
            }, nil
        }
        return nil, fmt.Errorf("failed to verify reset token: %v", err)
    }

    updatePasswordQuery := `
        UPDATE users 
        SET password_hash = $1, 
            updated_at = CURRENT_TIMESTAMP 
        WHERE user_id = $2 
        RETURNING extract(epoch from updated_at)::bigint`

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %v", err)
    }

    var updatedAt int64
    err = tx.QueryRowContext(ctx, updatePasswordQuery, hashedPassword, userID).Scan(&updatedAt)
    if err != nil {
        return nil, fmt.Errorf("failed to update password: %v", err)
    }

    markTokenUsedQuery := `
        UPDATE verification_tokens 
        SET is_used = true 
        WHERE token = $1`

    _, err = tx.ExecContext(ctx, markTokenUsedQuery, req.ResetToken)
    if err != nil {
        return nil, fmt.Errorf("failed to mark token as used: %v", err)
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %v", err)
    }

    return &pb.ResetPasswordResponse{
        Success: true,
        Message: "Password successfully reset",
        UpdatedAt: updatedAt,
    }, nil
}
