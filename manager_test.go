package multicache_test

import (
	"testing"
	"time"

	"github.com/shoraid/multicache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupManager(t *testing.T) (*multicache.MockStore, multicache.Manager) {
	mockStore := new(multicache.MockStore)
	manager, err := multicache.NewManager("default", map[string]multicache.Store{
		"default": mockStore,
	})
	assert.NoError(t, err)
	return mockStore, manager
}

func TestNewManager_InvalidDefault(t *testing.T) {
	manager, err := multicache.NewManager("notfound", map[string]multicache.Store{})
	assert.Nil(t, manager)
	assert.ErrorIs(t, err, multicache.ErrInvalidDefaultStore)
}

func TestManager_Clear(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Clear").Return(nil)

	err := manager.Clear()
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestManager_Delete(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Delete", "foo").Return(nil)

	err := manager.Delete("foo")
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestManager_DeleteByPattern(t *testing.T) {
	store, manager := setupManager(t)
	store.On("DeleteByPattern", "user:*").Return(nil)

	err := manager.DeleteByPattern("user:*")
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestManager_DeleteMany(t *testing.T) {
	store, manager := setupManager(t)
	store.On("DeleteMany", []string{"a", "b"}).Return(nil)

	err := manager.DeleteMany("a", "b")
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestManager_Get(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "foo").Return("bar", nil)

	val, err := manager.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)
	store.AssertExpectations(t)
}

func TestManager_GetBool(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "flag").Return(true, nil)

	val, err := manager.GetBool("flag")
	assert.NoError(t, err)
	assert.True(t, val)
	store.AssertExpectations(t)
}

func TestManager_GetInt(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "num").Return(42, nil)

	val, err := manager.GetInt("num")
	assert.NoError(t, err)
	assert.Equal(t, 42, val)
	store.AssertExpectations(t)
}

func TestManager_GetInt64(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "num64").Return(int64(12345), nil)

	val, err := manager.GetInt64("num64")
	assert.NoError(t, err)
	assert.Equal(t, int64(12345), val)
	store.AssertExpectations(t)
}

func TestManager_GetString(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "name").Return("john", nil)

	val, err := manager.GetString("name")
	assert.NoError(t, err)
	assert.Equal(t, "john", val)
	store.AssertExpectations(t)
}

func TestManager_GetOrSet(t *testing.T) {
	store, manager := setupManager(t)
	store.On("GetOrSet", "foo", mock.Anything, "bar").Return("bar", nil)

	val, err := manager.GetOrSet("foo", time.Minute, "bar")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)
	store.AssertExpectations(t)
}

func TestManager_Has(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Has", "foo").Return(true, nil)

	ok, err := manager.Has("foo")
	assert.NoError(t, err)
	assert.True(t, ok)
	store.AssertExpectations(t)
}

func TestManager_Set(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Set", "foo", "bar", mock.Anything).Return(nil)

	err := manager.Set("foo", "bar", time.Minute)
	assert.NoError(t, err)
	store.AssertExpectations(t)
}

func TestManager_StoreSwitch(t *testing.T) {
	storeA := new(multicache.MockStore)
	storeB := new(multicache.MockStore)

	manager, err := multicache.NewManager("a", map[string]multicache.Store{
		"a": storeA,
		"b": storeB,
	})
	assert.NoError(t, err)

	// storeA used by default
	storeA.On("Set", "foo", "bar", mock.Anything).Return(nil).Once()
	err = manager.Set("foo", "bar", time.Minute)
	assert.NoError(t, err)

	// switch to storeB
	mB := manager.Store("b")
	storeB.On("Set", "baz", "qux", mock.Anything).Return(nil).Once()
	err = mB.Set("baz", "qux", time.Minute)
	assert.NoError(t, err)

	storeA.AssertExpectations(t)
	storeB.AssertExpectations(t)
}

func TestManager_CastError(t *testing.T) {
	store, manager := setupManager(t)
	store.On("Get", "badint").Return("not-an-int", nil)

	_, err := manager.GetInt("badint")
	assert.Error(t, err)
}
