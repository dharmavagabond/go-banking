package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type AuthErrorBody struct {
	Error   string
	Message string
}

func addAuthorization(
	t *testing.T,
	req *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationToken := fmt.Sprintf("%s %s", authorizationType, token)
	req.Header.Set(AUTH_HEADER, authorizationToken)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMake token.Maker)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				username := "erosennin"
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, username, time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
				checkErrorMessage(t, rec.Body, "missing key in request header")
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				username := "erosennin"
				addAuthorization(t, req, tokenMaker, "Token", username, time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
				checkErrorMessage(t, rec.Body, "invalid key in the request header")
			},
		},
		{
			name: "InvalidAuthorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				username := "erosennin"
				addAuthorization(t, req, tokenMaker, "", username, time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
				checkErrorMessage(t, rec.Body, "invalid key in the request header")
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				username := "erosennin"
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, username, -time.Minute)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			server := NewServer(nil)
			authPath := "/auth"
			server.router.POST(
				authPath,
				func(ectx echo.Context) error {
					return ectx.JSON(http.StatusOK, echo.Map{})
				},
				authMiddleware,
			)
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, authPath, nil)
			req.Header.Add("Content-Type", "application/json")
			require.NoError(t, err)
			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func checkErrorMessage(t *testing.T, body *bytes.Buffer, message string) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var errorBody AuthErrorBody
	err = json.Unmarshal(data, &errorBody)
	require.NoError(t, err)
	require.Equal(t, message, errorBody.Message)
}
