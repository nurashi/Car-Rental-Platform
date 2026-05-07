package grpc

import (
	"context"
	"log"

	"github.com/nurashi/car-rental-identity/internal/api/grpc/pb"
	"github.com/nurashi/car-rental-identity/internal/repository"
	"github.com/nurashi/car-rental-identity/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedIdentityServiceServer
	authService    *service.AuthService
	userService    *service.UserService
	licenseService *service.LicenseService
	notifRepo      repository.NotificationRepository
}

func NewServer(authService *service.AuthService, userService *service.UserService, licenseService *service.LicenseService, notifRepo repository.NotificationRepository) *Server {
	return &Server{
		authService:    authService,
		userService:    userService,
		licenseService: licenseService,
		notifRepo:      notifRepo,
	}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, err := s.authService.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		log.Printf("register error: %v", err)
		return nil, status.Error(codes.Internal, "failed to register user")
	}
	return &pb.RegisterResponse{UserId: userID, Message: "registration successful, check your email for verification"}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, token, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if err == service.ErrUserBlocked {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		log.Printf("login error: %v", err)
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &pb.LoginResponse{Token: token, User: pb.UserToProto(*user)}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	if err := s.authService.VerifyEmail(ctx, req.Token); err != nil {
		if err == service.ErrInvalidToken {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		log.Printf("verify email error: %v", err)
		return nil, status.Error(codes.Internal, "failed to verify email")
	}
	return &pb.VerifyEmailResponse{Message: "email verified successfully"}, nil
}

func (s *Server) ResendVerification(ctx context.Context, req *pb.ResendVerificationRequest) (*pb.ResendVerificationResponse, error) {
	if err := s.authService.ResendVerification(ctx, req.Email); err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		log.Printf("resend verification error: %v", err)
		return nil, status.Error(codes.Internal, "failed to resend verification")
	}
	return &pb.ResendVerificationResponse{Message: "verification email sent"}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := s.authService.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}
	return &pb.LogoutResponse{Message: "logged out successfully"}, nil
}

func (s *Server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.ProfileResponse, error) {
	user, err := s.userService.GetProfile(ctx, req.UserId)
	if err != nil {
		if err == service.ErrProfileNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Printf("get profile error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get profile")
	}
	return &pb.ProfileResponse{User: pb.UserToProto(*user)}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.ProfileResponse, error) {
	input := service.UpdateProfileInput{UserID: req.UserId}
	if req.FirstName != nil {
		input.FirstName = req.FirstName
	}
	if req.LastName != nil {
		input.LastName = req.LastName
	}
	if req.Phone != nil {
		input.Phone = req.Phone
	}
	if req.AvatarUrl != nil {
		input.AvatarURL = req.AvatarUrl
	}

	user, err := s.userService.UpdateProfile(ctx, input)
	if err != nil {
		log.Printf("update profile error: %v", err)
		return nil, status.Error(codes.Internal, "failed to update profile")
	}
	return &pb.ProfileResponse{User: pb.UserToProto(*user)}, nil
}

func (s *Server) SubmitLicense(ctx context.Context, req *pb.SubmitLicenseRequest) (*pb.SubmitLicenseResponse, error) {
	input := service.SubmitLicenseInput{
		UserID:         req.UserId,
		LicenseNumber:  req.LicenseNumber,
		IssuingCountry: req.IssuingCountry,
		IssueDate:      req.IssueDate,
		ExpiryDate:     req.ExpiryDate,
	}

	license, err := s.licenseService.SubmitLicense(ctx, input)
	if err != nil {
		log.Printf("submit license error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.SubmitLicenseResponse{LicenseId: license.ID, Status: string(license.Status)}, nil
}

func (s *Server) ValidateLicense(ctx context.Context, req *pb.ValidateLicenseRequest) (*pb.ValidateLicenseResponse, error) {
	license, isValid, err := s.licenseService.ValidateLicense(ctx, req.UserId)
	if err != nil {
		if err == service.ErrLicenseNotFound {
			return &pb.ValidateLicenseResponse{IsValid: false, Status: "not_found"}, nil
		}
		log.Printf("validate license error: %v", err)
		return nil, status.Error(codes.Internal, "failed to validate license")
	}
	return &pb.ValidateLicenseResponse{
		IsValid: isValid,
		Status:  string(license.Status),
		License: pb.DriverLicenseToProto(*license),
	}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := s.userService.ListUsers(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		log.Printf("list users error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list users")
	}
	return &pb.ListUsersResponse{Users: pb.UsersToProto(users), Total: int32(total)}, nil
}

func (s *Server) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	if err := s.userService.BlockUser(ctx, req.UserId, req.Reason); err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		log.Printf("block user error: %v", err)
		return nil, status.Error(codes.Internal, "failed to block user")
	}
	return &pb.BlockUserResponse{Message: "user blocked successfully"}, nil
}

func (s *Server) SendEmailNotification(ctx context.Context, req *pb.SendEmailNotificationRequest) (*pb.SendEmailNotificationResponse, error) {
	log.Printf("SendEmailNotification: to=%s subject=%s (SMTP not configured)", req.To, req.Subject)
	return &pb.SendEmailNotificationResponse{Success: true, Message: "notification recorded"}, nil
}

func (s *Server) GetUserNotifications(ctx context.Context, req *pb.GetUserNotificationsRequest) (*pb.GetUserNotificationsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 20
	}
	notifications, err := s.notifRepo.GetByUserID(ctx, req.UserId, limit)
	if err != nil {
		log.Printf("get notifications error: %v", err)
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}
	return &pb.GetUserNotificationsResponse{Notifications: pb.NotificationsToProto(notifications)}, nil
}

var _ pb.IdentityServiceServer = (*Server)(nil)
