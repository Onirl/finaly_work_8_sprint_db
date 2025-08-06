package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {

	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "fail connection DataBase")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "fail insert parcel in table")
	require.NotEmpty(t, id)

	parcelNew, err := store.Get(id)
	require.NoError(t, err, "fail select parcel")
	assert.Equal(t, parcel.Client, parcelNew.Client,
		fmt.Sprintf("Expected %d received  %d", parcel.Client, parcelNew.Client))
	assert.Equal(t, parcel.Status, parcelNew.Status,
		fmt.Sprintf("Expected %s received  %s", parcel.Status, parcelNew.Status))
	assert.Equal(t, parcel.Address, parcelNew.Address,
		fmt.Sprintf("Expected %s received  %s", parcel.Address, parcelNew.Address))
	assert.Equal(t, parcel.CreatedAt, parcelNew.CreatedAt,
		fmt.Sprintf("Expected %s received  %s", parcel.CreatedAt, parcelNew.CreatedAt))

	err = store.Delete(id)
	require.NoError(t, err, "fail delete from table")

	_, err = store.Get(id)
	require.Equal(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "fail connection DataBase")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "fail select parcel")
	require.NotEmpty(t, id, "received id empty")

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "fail update address")

	parcelNew, err := store.Get(id)
	require.NoError(t, err, "fail select parcel")
	assert.Equal(t, newAddress, parcelNew.Address,
		fmt.Sprintf("Expected %s received  %s", newAddress, parcelNew.Address))

}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "fail connection DataBase")
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "fail insert parcel in table")
	require.NotEmpty(t, id)

	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err, "fail update status")

	parcelNew, err := store.Get(id)
	require.NoError(t, err, "fail select parcel")
	assert.Equal(t, ParcelStatusDelivered, parcelNew.Status,
		fmt.Sprintf("Expected %s received  %s", ParcelStatusDelivered, parcelNew.Status))
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "fail connection DataBase")
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "fail insert parcel in table")
		require.NotEmpty(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "fail get client")
	require.Len(t, storedParcels, len(parcelMap),
		fmt.Sprintf("Expected %d received  %d", len(parcelMap), len(storedParcels)))

	// check
	for _, parcel := range storedParcels {
		retrievedParcel, ok := parcelMap[parcel.Number]
		assert.True(t, ok,
			fmt.Sprintf("parcel with number %d not found in map", parcel.Number))

		assert.Equal(t, retrievedParcel, parcel,
			fmt.Sprintf("Expected %v received  %v", retrievedParcel, parcel))
	}
}
