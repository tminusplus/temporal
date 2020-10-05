package tests

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/server/common/persistence/sql/sqlplugin"
	"go.temporal.io/server/common/shuffle"
)

type (
	historyShardSuite struct {
		suite.Suite
		*require.Assertions

		store sqlplugin.HistoryShard
	}
)

const (
	testHistoryShardEncoding = "random encoding"
)

var (
	testHistoryShardData = []byte("random history shard data")
)

func newHistoryShardSuite(
	t *testing.T,
	store sqlplugin.HistoryShard,
) *historyShardSuite {
	return &historyShardSuite{
		Assertions: require.New(t),
		store:      store,
	}
}

func (s *historyShardSuite) SetupSuite() {

}

func (s *historyShardSuite) TearDownSuite() {

}

func (s *historyShardSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *historyShardSuite) TearDownTest() {

}

func (s *historyShardSuite) TestInsert_Success() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))
}

func (s *historyShardSuite) TestInsert_Fail_Duplicate() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	shard = s.newRandomShardRow(shardID, rangeID)
	_, err = s.store.InsertIntoShards(&shard)
	s.Error(err) // TODO persistence layer should do proper error translation
}

func (s *historyShardSuite) TestInsertSelect() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	filter := &sqlplugin.ShardsFilter{
		ShardID: shardID,
	}
	row, err := s.store.SelectFromShards(filter)
	s.NoError(err)
	s.Equal(&shard, row)
}

func (s *historyShardSuite) TestInsertUpdate_Success() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	rangeID += 100
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	shard = s.newRandomShardRow(shardID, rangeID)
	result, err = s.store.UpdateShards(&shard)
	s.NoError(err)
	rowsAffected, err = result.RowsAffected()
	s.Equal(1, int(rowsAffected))
}

func (s *historyShardSuite) TestUpdate_Fail() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.UpdateShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.Equal(0, int(rowsAffected))
}

func (s *historyShardSuite) TestInsertUpdateSelect() {
	shardID := int64(rand.Int31())
	rangeID := int64(1)

	shard := s.newRandomShardRow(shardID, rangeID)
	rangeID += 100
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	shard = s.newRandomShardRow(shardID, rangeID)
	result, err = s.store.UpdateShards(&shard)
	s.NoError(err)
	rowsAffected, err = result.RowsAffected()
	s.Equal(1, int(rowsAffected))

	filter := &sqlplugin.ShardsFilter{
		ShardID: shardID,
	}
	row, err := s.store.SelectFromShards(filter)
	s.NoError(err)
	s.Equal(&shard, row)
}

func (s *historyShardSuite) TestSelectReadLock() {
	shardID := int64(rand.Int31())
	rangeID := int64(rand.Int31())

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	// NOTE: lock without transaction is equivalent to select
	//  this test only test the select functionality
	filter := &sqlplugin.ShardsFilter{
		ShardID: shardID,
	}
	shardRange, err := s.store.ReadLockShards(filter)
	s.NoError(err)
	s.Equal(rangeID, shardRange)
}

func (s *historyShardSuite) TestSelectWriteLock() {
	shardID := int64(rand.Int31())
	rangeID := int64(rand.Int31())

	shard := s.newRandomShardRow(shardID, rangeID)
	result, err := s.store.InsertIntoShards(&shard)
	s.NoError(err)
	rowsAffected, err := result.RowsAffected()
	s.NoError(err)
	s.Equal(1, int(rowsAffected))

	// NOTE: lock without transaction is equivalent to select
	//  this test only test the select functionality
	filter := &sqlplugin.ShardsFilter{
		ShardID: shardID,
	}
	shardRange, err := s.store.WriteLockShards(filter)
	s.NoError(err)
	s.Equal(rangeID, shardRange)
}

func (s *historyShardSuite) newRandomShardRow(
	shardID int64,
	rangeID int64,
) sqlplugin.ShardsRow {
	return sqlplugin.ShardsRow{
		ShardID:      shardID,
		RangeID:      rangeID,
		Data:         shuffle.Bytes(testHistoryShardData),
		DataEncoding: testHistoryShardEncoding,
	}
}