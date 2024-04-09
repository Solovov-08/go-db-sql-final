package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	_ "github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
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
	// Подготовка бд
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Error adding parcel: %v", err)
	}
	parcel.Number = id

	// Получение
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Error getting parcel: %v", err)
	}

	// Проверка, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	if !reflect.DeepEqual(parcel, retrievedParcel) {
		t.Fatalf("Retrieved parcel does not match expected parcel")
	}

	// Удаление
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Error deleting parcel: %v", err)
	}

	// Проверка, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	if err == nil {
		t.Fatalf("Expected error when trying to get deleted parcel, got nil")
	}

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Error adding parcel: %v", err)
	}
	parcel.Number = id
	// Установка нового адреса
	newAddress := "test"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatalf("Error setting address: %v", err)
	}

	// Проверка
	// Получение добавленной посылки
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Error getting parcel: %v", err)
	}

	// Проверка, что адрес обновился
	if retrievedParcel.Address != newAddress {
		fmt.Println(retrievedParcel)
		t.Fatalf("Expected address to be %s, got %s", newAddress, retrievedParcel.Address)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Добавление
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Error adding parcel: %v", err)
	}
	parcel.Number = id
	// Установка нового статуса
	newStatus := "new status"
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatalf("Error setting status: %v", err)
	}

	// Проверка
	// Получение добавленной посылки
	retrievedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Error getting parcel: %v", err)
	}

	// Проверка, что статус обновился
	if retrievedParcel.Status != newStatus {
		t.Fatalf("Expected status to be %s, got %s", newStatus, retrievedParcel.Status)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// Подготовка
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	// Задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// Добавление
	for i := 0; i < len(parcels); i++ {
		fmt.Println(parcels[i])
		// Добавляем новую посылку в базу данных
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("Error adding parcel: %v", err)
		}
		parcels[i].Number = id // Записываем id в посылку
		// Сохраняем добавленную посылку в структуру map
		parcelMap[id] = parcels[i]
	}

	// Получение посылок по идентификатору клиента
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("Error getting parcels by client: %v", err)
	}
	fmt.Println(storedParcels)

	// Проверка, что количество полученных посылок совпадает с количеством добавленных
	if len(storedParcels) != len(parcels) {
		t.Fatalf("Expected %d parcels, got %d", len(parcels), len(storedParcels))
	}

	// Проверка, что все посылки из storedParcels есть в parcelMap и что значения полей верны
	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]
		if !ok {
			t.Fatalf("Parcel with number %d not found in parcelMap", parcel.Number)
		}

		// Проверка значений полей
		if !reflect.DeepEqual(parcel, expectedParcel) {
			t.Fatalf("Parcel values do not match expected values")
		}
	}
}
