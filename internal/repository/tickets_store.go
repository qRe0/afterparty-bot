package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/models"
)

type TicketsRepo struct {
	db  *sqlx.DB
	cfg configs.DBConfig
}

func NewTicketsRepository(db *sqlx.DB, cfg configs.DBConfig) *TicketsRepo {
	return &TicketsRepo{
		cfg: cfg,
		db:  db,
	}
}

const (
	connectingStringTemplate = "postgres://%s:%s@%s:%s/%s?sslmode=disable"

	findClientByFullSurname = "SELECT id, full_name, ticket_type FROM tickets WHERE surname=$1"
	findClientBySurnamePart = "SELECT id, full_name, ticket_type  FROM tickets WHERE surname LIKE $1"
)

func Init(cfg configs.DBConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(connectingStringTemplate, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("repository.Init().Open(): failed to conncect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("repository.Init().Ping(): failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	log.Println("Connected to DB successfully!")

	return db, nil
}

func (tr *TicketsRepo) SearchByFullSurname(ctx context.Context, surname string) (*models.TicketResponse, error) {
	var name, ticketType string
	var id string

	err := tr.db.QueryRow(findClientByFullSurname, surname).Scan(&id, &name, &ticketType)
	if err != nil {
		return nil, err
	}

	resp := &models.TicketResponse{
		Id:         id,
		Name:       name,
		TicketType: ticketType,
	}
	return resp, nil
}

func (tr *TicketsRepo) SearchBySurnamePart(ctx context.Context, surnamePart string) ([]models.TicketResponse, error) {
	surnamePattern := surnamePart + "%"
	rows, err := tr.db.QueryContext(ctx, findClientBySurnamePart, surnamePattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.TicketResponse
	for rows.Next() {
		var user models.TicketResponse
		err := rows.Scan(&user.Id, &user.Name, &user.TicketType)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
