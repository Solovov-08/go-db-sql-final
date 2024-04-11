package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	// Добавляем строку в таблицу parcel и возвращаем ID
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))

	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {

	//  Чтение строки по заданному number
	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number))

	// Заполняем объект Parcel данными из таблицы
	p := Parcel{}

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		fmt.Println(err)
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {

	// Чтение строк из таблицы parcel по заданному client
	var res []Parcel

	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		fmt.Println(err)
		return res, err
	}
	// Заполняем срез Parcel данными из таблицы
	defer rows.Close()
	for rows.Next() {
		var pbox Parcel

		err := rows.Scan(&pbox.Number, &pbox.Client, &pbox.Status, &pbox.Address, &pbox.CreatedAt)
		if err != nil {
			fmt.Println(err)
			return []Parcel{}, err
		}
		res = append(res, pbox)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {

	// Обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Начало транзакции
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Получение статуса посылки
	row := tx.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number))
	var status string
	err = row.Scan(&status)
	if err != nil {
		tx.Rollback()
		return err
	}

	if status == ParcelStatusRegistered {
		// Обновление адреса посылки
		_, err = tx.Exec("UPDATE parcel SET address = :address WHERE number = :number", sql.Named("address", address), sql.Named("number", number))
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		tx.Rollback()
		return fmt.Errorf("error: order status - registered")
	}

	// Фиксация транзакции
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", number))
	var status string
	err = row.Scan(&status)
	if err != nil {
		tx.Rollback()
		return err
	}

	if status == ParcelStatusRegistered {
		_, err = tx.Exec("DELETE FROM parcel WHERE number = :number", sql.Named("number", number))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
