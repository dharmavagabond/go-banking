package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/alexedwards/argon2id"
	dberrors "github.com/dharmavagabond/simple-bank/internal/db"
	"github.com/dharmavagabond/simple-bank/internal/db/mock"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser()
	testCases := []struct {
		name          string
		body          echo.Map
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
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					CreateUser(mock.AnythingOfType("*context.emptyCtx"), mock.Anything).
					Once().
					Return(db.User{}, &pgconn.PgError{Code: dberrors.ERRCODE_UNIQUE_VIOLATION})
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
			server := NewServer(store)
			rec := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			body := bytes.NewReader(data)
			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, body)
			request.Header.Add("Content-Type", "application/json")
			require.NoError(t, err)
			server.router.ServeHTTP(rec, request)
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
