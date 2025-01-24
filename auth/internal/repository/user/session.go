package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wafi04/common/pkg/logger"
	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
)


func (sr *UserRepository)   RevokeSession(ctx context.Context,req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse,error){
	sr.logger.Log(logger.InfoLevel, "Recieved  Session Request ")

	query :=  `
	DELETE FROM sessions
    WHERE session_id = $1 AND user_id = $2
	`
	_, err := sr.DB.ExecContext(ctx, query,req.SessionId,req.UserId)

	if err != nil {
		sr.logger.Log(logger.ErrorLevel, "Failed to Delete Session : %v",err)
		return nil, nil
	}

	return &pb.RevokeSessionResponse{
		Success: true,},nil
}
func (sr *UserRepository) CreateSession(ctx context.Context, session *pb.Session) error {
   // Pisahkan query insert dan update
   insertQuery := `
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
       ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
   `

   updateQuery := `
       UPDATE sessions 
       SET 
           access_token = $1, 
           refresh_token = $2, 
           ip_address = $3, 
           last_activity_at = $4
       WHERE user_id = $5 AND device_info = $6
   `

   if session.SessionId == "" {
       session.SessionId = uuid.New().String()
   }

   now := time.Now()
   expiresAt := now.Add(24 * time.Hour)

   // Pertama, coba insert
   _, err := sr.DB.ExecContext(
       ctx, 
       insertQuery, 
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

   // Jika insert gagal (misal duplicate), lakukan update
   if err != nil {
       _, err = sr.DB.ExecContext(
           ctx, 
           updateQuery, 
           session.AccessToken, 
           session.RefreshToken, 
           session.IpAddress, 
           now,
           session.UserId, 
           session.DeviceInfo,
       )
   }

   if err != nil {
       sr.logger.WithError(err).WithFields(map[string]interface{}{
           "user_id":    session.UserId, 
           "session_id": session.SessionId, 
       }).Error("Failed to create/update session")
       return fmt.Errorf("failed to create session: %w", err)
   }

   return nil
}
func (sr *UserRepository) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
    sr.logger.Log(logger.InfoLevel, "Refresh Token Incoming")

    query := `
        SELECT u.user_id, u.email, u.role, u.is_email_verified
        FROM sessions s
        JOIN users u ON s.user_id = u.user_id
        WHERE s.session_id = $1
    `
    var user pb.UserInfo
    err := sr.DB.QueryRowContext(ctx, query, req.SessionId).Scan(
        &user.UserId, 
        &user.Email, 
        &user.Role, 
        &user.IsEmailVerified,
    )
    if err != nil {
        sr.logger.Log(logger.ErrorLevel, "Failed to retrieve user: %v", err)
        return nil, err
    }

    updateQuery := `
        UPDATE sessions SET access_token = $1 WHERE session_id = $2
    `
    _, err = sr.DB.ExecContext(ctx, updateQuery, req.RefreshToken, req.SessionId)
    if err != nil {
        sr.logger.Log(logger.ErrorLevel, "Failed to Refresh Token: %v", err)
        return nil, err
    }

    token, err := middleware.GenerateToken(&user)
    if err != nil {
        return nil, err
    }

    return &pb.RefreshTokenResponse{
        AccessToken:  token,
        RefreshToken: token,
        ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
    }, nil
}


func (sr *UserRepository) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
    query := `
        SELECT 
            session_id,
            device_info,
            ip_address,
            EXTRACT(EPOCH FROM created_at)::bigint AS created_at,
            EXTRACT(EPOCH FROM last_activity_at)::bigint AS last_activity_at
        FROM sessions
        WHERE user_id = $1
    `
    
    rows, err := sr.DB.QueryContext(ctx, query, req.UserId)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sessions []*pb.SessionInfo
    for rows.Next() {
        session := &pb.SessionInfo{}
        err := rows.Scan(
            &session.SessionId, 
            &session.DeviceInfo, 
            &session.IpAddress, 
            &session.CreatedAt, 
            &session.LastActivityAt,
        )
        if err != nil {
            return nil, err
        }
        sessions = append(sessions, session)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return &pb.ListSessionsResponse{
        Sessions: sessions,
    }, nil
}


