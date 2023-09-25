package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/mocks"
	"github.com/dharmavagabond/simple-bank/internal/token"
	"github.com/dharmavagabond/simple-bank/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: randomdata.Currency(),
	}
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser()
	account := createRandomAccount(user.Username)
	log.Print(account)
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMake token.Maker)
		buildStubs    func(store *mocks.Store)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				t.Helper()
				addAuthorization(
					t,
					req,
					tokenMaker,
					AUTH_TYPE_BEARER,
					account.Owner,
					time.Minute,
				)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.IsType(account.ID)).
					Once().
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				requireBodyMatchAccount(t, rec.Body, account)
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				t.Helper()
				addAuthorization(
					t,
					req,
					tokenMaker,
					AUTH_TYPE_BEARER,
					"unauthorized_user",
					time.Minute,
				)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.IsType(account.ID)).
					Once().
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {}, //nolint: thelper
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.IsType(account.ID)).
					Maybe().
					Times(0)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				t.Helper()
				addAuthorization(
					t,
					req,
					tokenMaker,
					AUTH_TYPE_BEARER,
					user.Username,
					time.Minute,
				)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.IsType(account.ID)).
					Once().
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				t.Helper()
				addAuthorization(
					t,
					req,
					tokenMaker,
					AUTH_TYPE_BEARER,
					user.Username,
					time.Minute,
				)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.IsType(account.ID)).
					Once().
					Return(db.Account{}, &pgconn.PgError{})
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				t.Helper()
				addAuthorization(
					t,
					req,
					tokenMaker,
					AUTH_TYPE_BEARER,
					user.Username,
					time.Minute,
				)
			},
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.AnythingOfType("context.todoCtx"), mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
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
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequestWithContext(
				context.TODO(),
				http.MethodGet,
				url,
				nil,
			)
			require.NoError(t, err)
			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func requireBodyMatchAccount(
	t *testing.T,
	body *bytes.Buffer,
	expected db.Account,
) {
	t.Helper()
	var account db.Account
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	err = json.Unmarshal(data, &account)
	require.NoError(t, err)
	require.Equal(t, expected, account)
}
