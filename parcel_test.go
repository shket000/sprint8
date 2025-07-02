package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC(),
	}
}

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE parcel (number INTEGER PRIMARY KEY AUTOINCREMENT, client INTEGER, status TEXT, address TEXT, created_at DATETIME)")
	require.NoError(t, err)

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, id, p.Number)
	require.Equal(t, parcel.Client, p.Client)
	require.Equal(t, parcel.Status, p.Status)
	require.Equal(t, parcel.Address, p.Address)
	require.False(t, p.CreatedAt.IsZero())

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

func TestSetAddressOnSentParcel(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	err = store.SetAddress(id, "new_address")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestSetStatus(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, p.Status)
}

func TestDeleteSentParcel(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	err = store.Delete(id)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestGetByClient(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expectedParcel.Client, parcel.Client)
		require.Equal(t, expectedParcel.Status, parcel.Status)
		require.Equal(t, expectedParcel.Address, parcel.Address)
		require.False(t, parcel.CreatedAt.IsZero())
	}
}
