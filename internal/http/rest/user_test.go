package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/alexedwards/argon2id"
	"github.com/dharmavagabond/simple-bank/internal/db/mock"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser()
	testCases := []struct {
		name          string
		body          echo.Map
		setupAuth     func(t *testing.T, req *http.Request, tokenMake token.Maker)
		buildStubs    func(store *mocks.Store)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: echo.Map{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.
					EXPECT().
					CreateUser(
						mock.AnythingOfType("*context.emptyCtx"),
						mock.MatchedBy(func(input db.CreateUserParams) bool {
							if ok, err := argon2id.ComparePasswordAndHash(password, input.HashedPassword); err != nil || !ok {
								return false
							}

							arg.HashedPassword = input.HashedPassword
							return reflect.DeepEqual(input, arg)
						}),
					).
					Once().
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchUser(t, rec.Body, user)
			},
		},
		{
			name: "InternalError",
			body: echo.Map{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Once().
					Return(db.User{}, &pgconn.PgError{})
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: echo.Map{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Once().
					Return(db.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, rec.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: echo.Map{
				"username":  "invalid-user#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: echo.Map{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: echo.Map{
				"username":  user.Username,
				"password":  "123",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, req, tokenMaker, AUTH_TYPE_BEARER, user.Username, time.Minute)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			store := mocks.NewStore(t)
			tc.buildStubs(store)
			server, err := NewServer(store)
			require.NoError(t, err)
			rec := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			body := bytes.NewReader(data)
			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, body)
			req.Header.Add("Content-Type", "application/json")
			require.NoError(t, err)
			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func randomUser() (user db.User, password string) {
	password = randomdata.Alphanumeric(16)
	hashedPassword, _ := argon2id.CreateHash(randomdata.Alphanumeric(16), argon2id.DefaultParams)
	user = db.User{
		Username:       strings.ToLower(randomdata.SillyName()),
		HashedPassword: hashedPassword,
		FullName:       randomdata.FullName(randomdata.RandomGender),
		Email:          randomdata.Email(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
