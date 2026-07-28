package lstep

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type failingLineLinkAuditTxLogger struct {
	err error
}

func (l failingLineLinkAuditTxLogger) LogOwnerLineLinkTx(
	context.Context,
	uint64,
	uint64,
) error {
	return l.err
}

func TestLineLinkService_AuditFailureRollsBackOwnerAndTokenInDatabase(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	owner := testdb.MakeTestOwner(t, db, 1, "LINE audit rollback owner")
	rawToken := "audit-rollback-token"
	token := makeLineLinkToken(
		t,
		db,
		owner.ClinicID,
		owner.ID,
		digestLineLinkToken(rawToken),
		time.Now().Add(time.Hour),
		nil,
	)
	auditErr := errors.New("forced audit failure")
	service := &lineLinkService{
		ownerRepo:         ownerdomain.NewRepository(db, nil),
		lineLinkTokenRepo: NewLineLinkTokenRepository(db),
		lineSettingRepo:   &mockLineLinkSettingRepo{},
		transactor:        persistence.NewTransactor(db),
		auditTx:           failingLineLinkAuditTxLogger{err: auditErr},
		httpClient:        jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	}

	got, err := service.LinkAccount(
		context.Background(),
		owner.ClinicID,
		LinkAccountInput{LinkToken: rawToken, LineIDToken: "valid"},
	)

	assert.Nil(t, got)
	require.ErrorIs(t, err, auditErr)

	var persistedOwner model.Owner
	require.NoError(t, db.First(&persistedOwner, owner.ID).Error)
	assert.Nil(t, persistedOwner.LineUserID)

	var persistedToken model.LineLinkToken
	require.NoError(t, db.First(&persistedToken, token.ID).Error)
	assert.Nil(t, persistedToken.UsedAt)
}
