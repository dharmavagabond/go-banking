package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/dharmavagabond/simple-bank/internal/db/mock"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/util"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createRandomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    randomdata.FullName(randomdata.RandomGender),
		Balance:  util.RandomMoney(),
		Currency: randomdata.Currency(),
	}
}

func TestGetAccountAPI(t *testing.T) {
	account := createRandomAccount()
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mocks.Store)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.Anything, mock.IsType(account.ID)).
					Once().
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
				requireBodyMatchAccount(t, rec.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.Anything, mock.IsType(account.ID)).
					Once().
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, rec.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.Anything, mock.IsType(account.ID)).
					Once().
					Return(db.Account{}, &pgconn.PgError{})
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStubs: func(store *mocks.Store) {
				store.
					EXPECT().
					GetAccount(mock.Anything, mock.Anything).
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
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			server.router.ServeHTTP(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, expected db.Account) {
	var account db.Account
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	err = json.Unmarshal(data, &account)
	require.NoError(t, err)
	require.Equal(t, account, expected)
}
