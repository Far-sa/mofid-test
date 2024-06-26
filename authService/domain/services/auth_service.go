package services

import (
	"authentication-service/domain/entities"
	"authentication-service/domain/param"
	"authentication-service/interfaces"
	"authentication-service/utils"
	mapper "authentication-service/utils/protobufMapper"
	"context"
	"errors"
	"log"
	user "user-service/pb"

	"golang.org/x/crypto/bcrypt"
)

// AuthenticationService interface defines methods for user authentication

// AuthService implements the AuthenticationService interface
type AuthService struct {
	authRepository   interfaces.AuthRepository
	messagePublisher interfaces.AuthEvents
	userClient       user.UserServiceClient
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(authRepository interfaces.AuthRepository, messagePublisher interfaces.AuthEvents) *AuthService {
	return &AuthService{
		authRepository:   authRepository,
		messagePublisher: messagePublisher,
	}
}

// * Event Publication: The authentication service publishes a UserRegisteredEvent to RabbitMQ.
//*  event := models.UserRegisteredEvent / user_registered

func (s *AuthService) Login(ctx context.Context, loginRequest param.LoginRequest) (param.LoginResponse, error) {

	protoReq, err := mapper.ToProtoGetUserEmailRequest(loginRequest)
	if err != nil {
		log.Printf("Error converting login request to proto: %v", err)
		return param.LoginResponse{}, errors.New("internal server error")
	}

	userResp, err := s.userClient.GetUserByEmail(ctx, protoReq)
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
		return param.LoginResponse{}, errors.New("internal server error")
	}

	paramUser, err := mapper.ToParamGetUserResponse(userResp)
	if err != nil {
		log.Printf("Error converting user response to param: %v", err)
		return param.LoginResponse{}, errors.New("internal server error")
	}

	//* get user data from own database and compare passwords
	// user, err := s.authRepository.FindByUserEmail(ctx, loginRequest.Email)
	// if err != nil {
	// 	return param.LoginResponse{}, err
	// }

	//* Validate the password (compare hashed password with provided password)
	if !isValidPassword(loginRequest.Password, string(paramUser.Password)) {
		return param.LoginResponse{}, errors.New("invalid credentials")
	}

	//* Generate tokens using the utils package
	accessToken, err := utils.GenerateAccessToken(paramUser.Id)
	if err != nil {
		return param.LoginResponse{}, err
	}

	// Optionally, generate a refresh token as well
	refreshToken, err := utils.GenerateRefreshToken(paramUser.Id)
	if err != nil {
		return param.LoginResponse{}, err
	}

	tokens := &entities.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	s.authRepository.SaveToken(ctx, tokens)

	return param.LoginResponse{UserID: paramUser.Id, Tokens: param.Token{AccessToken: accessToken, RefreshToken: refreshToken}}, nil
}

// func (s *AuthService) convertTokens(userID string, tokenStrings ...string) []entities.TokenPair {
// 	var tokens []entities.TokenPair
// 	for _, tokenString := range tokenStrings {
// 		token := entities.TokenPair{
// 			AccessToken:  entities.AccessToken{Token: tokenString, ExpiresAt: time.Now().Add(24 * time.Hour)},
// 			RefreshToken: entities.RefreshToken{Token: tokenString, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)},
// 		}
// 		tokens = append(tokens, token)
// 	}
// 	return tokens
// }

// isValidPassword checks if the provided password matches the hashed password.
func isValidPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func generateCorrelationID() string {
	// Implement a method to generate a unique correlation ID
	return "some-unique-correlation-id"
}

//!!

// func (s *AuthService) fetchUserData(ctx context.Context, usernameOrEmail string) (param.UserResponse, error) {
// 	request := map[string]string{"usernameOrEmail": usernameOrEmail}
// 	body, err := json.Marshal(request)
// 	if err != nil {
// 		return param.UserResponse{}, err
// 	}

// 	err = s.messagePublisher.Publish("auth_exchange", "auth_to_user_key", body)
// 	if err != nil {
// 		return param.UserResponse{}, err
// 	}

// 	// Listen for response from User Service
// 	msgs, err := s.messagePublisher.Consume("auth_response_queue")
// 	if err != nil {
// 		return param.UserResponse{}, err
// 	}

// 	for d := range msgs {
// 		var user param.UserResponse
// 		if err := json.Unmarshal(d.Body, &user); err != nil {
// 			log.Printf("Failed to unmarshal user response: %v", err)
// 			continue
// 		}
// 		return &user, nil
// 	}

// 	return param.UserResponse{}, errors.New("no response from user service")
// }
