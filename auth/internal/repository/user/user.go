package user

import (
	"database/sql"

	"github.com/wafi04/go-testing/common/pkg/logger"
)

type UserRepository struct {
	DB     *sql.DB
	logger logger.Logger
}

// func NewUserRepository(db *sql.DB) *UserRepository {
// 	return &UserRepository{
// 		DB: db,
// 	}
// }

// func (r *UserRepository) CreateUser(ctx context.Context, user *pb.CreateUserRequest) (pb.CreateUserResponse, error) {
// 	role := "user"
// 	if user.Email == "wafiq610@gmail.com" {
// 		role = "admin"
// 	}
// 	userID := uuid.New().String()
// 	now := time.Now()

// 	tx, err := r.DB.BeginTx(ctx, nil)
// 	if err != nil {
// 		r.logger.Log(logger.ErrorLevel, "Failed Trasacation : %v", err)
// 		return pb.CreateUserResponse{}, nil
// 	}
// 	defer tx.Rollback()

// 	query := `
//         INSERT INTO users (
//             id, name, email, password, role,
//             is_active, email_verified, created_at, updated_at
//         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
//     `

// 	_, err = tx.ExecContext(
// 		ctx, query,
// 		userID, user.Name, user.Email, user.Password, role,
// 		true, false, now, now,
// 	)
// 	appPW := configs.MustGetEnv("APP_PASSWORD")
// 	cleanPassword := strings.ReplaceAll(appPW, " ", "")
// 	emailSender := mailer.NewEmailSender(
// 		"smtp.gmail.com",
// 		587,
// 		configs.MustGetEnv("APP_EMAIL"),
// 		cleanPassword,
// 	)

// 	toEmail := user.Email
// 	verificationCode := mailer.GenerateVerificationCode()

// 	if err := emailSender.SendVerificationEmail(toEmail, user.Name, verificationCode); err != nil {
// 		return pb.CreateUserResponse{}, fmt.Errorf("failed to send email : %w", err)
// 	}

// 	if err != nil {
// 		return pb.CreateUserResponse{}, fmt.Errorf("failed to insert user: %w", err)
// 	}
// 	if err != nil {
// 		r.logger.Log(logger.ErrorLevel, "Failed to insert user : %v", err)
// 		return pb.CreateUserResponse{}, nil
// 	}
// 	expiresAt := now.Add(24 * time.Hour)

// 	query = `
//         INSERT INTO verification_tokens (
//             id, user_id, token, type, expires_at
//         ) VALUES ($1, $2, $3, $4, $5)
//     `

// 	_, err = tx.ExecContext(
// 		ctx, query,
// 		uuid.New().String(), userID, verificationCode, "email_verification", expiresAt,
// 	)

// 	if err != nil {
// 		return pb.CreateUserResponse{}, fmt.Errorf("failed to create verification token: %w", err)
// 	}

// 	if err = tx.Commit(); err != nil {
// 		return pb.CreateUserResponse{}, err
// 	}

// }
