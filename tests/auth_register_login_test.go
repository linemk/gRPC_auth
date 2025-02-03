package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/linemk/gRPC_auth/tests/suite"
	ssov1 "github.com/linemk/proto_buf/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

const (
	emptyAppID     = 0
	appId          = 1
	appSecret      = "test-secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appId,
	})

	require.NoError(t, err)

	loginTime := time.Now()
	token := respLogin.GetToken()

	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	respLogin.GetToken()
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	const (
		deltaSeconds = 1
	)

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}
func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestFailRegisterWithEmptyEmail_FailKey(t *testing.T) {
	ctx, st := suite.New(t)
	email := ""
	pass := randomFakePassword()

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = InvalidArgument desc = Email or password is required")
}

func TestFailRegisterWithEmptyPassword_FailKey(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := ""

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Equal(t, err.Error(), "rpc error: code = InvalidArgument desc = Email or password is required")
}

func TestRegister_Login_DuplicatedEmail_FailKey(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	var wg sync.WaitGroup

	tests := []struct {
		name          string
		email         string
		password      string
		expectedError string
	}{
		{
			name:          "empty email",
			email:         "",
			password:      randomFakePassword(),
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		}, {
			name:          "empty password",
			email:         gofakeit.Email(),
			password:      "",
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		}, {
			name:          "empty email and password",
			email:         "",
			password:      "",
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		},
	}

	for _, tt := range tests {
		wg.Add(1)
		go t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			assert.Equal(t, err.Error(), tt.expectedError)
		})
		wg.Wait()
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	var wg sync.WaitGroup

	tests := []struct {
		name          string
		email         string
		password      string
		appID         int32
		expectedError string
	}{
		{
			name:          "empty password",
			email:         gofakeit.Email(),
			password:      "",
			appID:         appId,
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		}, {
			name:          "empty email",
			email:         "",
			password:      randomFakePassword(),
			appID:         appId,
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		}, {
			name:          "empty email and password",
			email:         "",
			password:      "",
			appID:         appId,
			expectedError: "rpc error: code = InvalidArgument desc = Email or password is required",
		}, {
			name:          "empty ID",
			email:         gofakeit.Email(),
			password:      randomFakePassword(),
			appID:         0,
			expectedError: "rpc error: code = InvalidArgument desc = AppId is required",
		}, {
			name:          "fake password",
			email:         gofakeit.Email(),
			password:      randomFakePassword(),
			appID:         appId,
			expectedError: "rpc error: code = Unauthenticated desc = auth.Login: invalid credentials",
		},
	}
	for _, tt := range tests {
		wg.Add(1)
		go t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			assert.Equal(t, err.Error(), tt.expectedError)
		})
		wg.Wait()
	}

}
