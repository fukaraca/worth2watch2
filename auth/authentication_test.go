package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
	client *redis.Client
}

var mockClient = &mockAuthService{}

func (mockAuthService) CheckCookie(c *gin.Context, toBeChecked, userId string) bool {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) CreateSession(username string, c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) CheckSession(c *gin.Context) bool {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) DeleteSession(c *gin.Context) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) CheckAdminForLoggedIn(c *gin.Context, username string) bool {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) CloseCacheConnection() {
	//TODO implement me
	panic("implement me")
}

func (mockAuthService) InitializeCache() {

}
