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
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER,
			status TEXT,
			address TEXT,
			created_at TEXT
		)`)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestAddGetDelete(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add parcel
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Get parcel
	stored, err := store.Get(id)
	require.NoError(t, err)

	// Check fields
	require.Equal(t, id, stored.Number)
	require.Equal(t, parcel.Client, stored.Client)
	require.Equal(t, parcel.Status, stored.Status)
	require.Equal(t, parcel.Address, stored.Address)
	require.Equal(t, parcel.CreatedAt, stored.CreatedAt)

	// Delete parcel
	err = store.Delete(id)
	require.NoError(t, err)

	// Check deletion
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add parcel
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Set new address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Get updated parcel
	updated, err := store.Get(id)
	require.NoError(t, err)

	// Check updated address
	require.Equal(t, newAddress, updated.Address)

	// Check other fields remain unchanged
	require.Equal(t, parcel.Client, updated.Client)
	require.Equal(t, parcel.Status, updated.Status)
	require.Equal(t, parcel.CreatedAt, updated.CreatedAt)
}

func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add parcel
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Update status
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// Get updated parcel
	updated, err := store.Get(id)
	require.NoError(t, err)

	// Check updated status
	require.Equal(t, ParcelStatusSent, updated.Status)

	// Check other fields remain unchanged
	require.Equal(t, parcel.Client, updated.Client)
	require.Equal(t, parcel.Address, updated.Address)
	require.Equal(t, parcel.CreatedAt, updated.CreatedAt)
}

func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)

	// Create test parcels
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// Add parcels
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Get parcels by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// Check all parcels
	for _, p := range storedParcels {
		expected, exists := parcelMap[p.Number]
		require.True(t, exists)
		require.Equal(t, expected, p)
	}
}
