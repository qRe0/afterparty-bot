package ticket_repository

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

func New(db *sqlx.DB, cfg configs.DBConfig) *TicketsRepo {
	return &TicketsRepo{
		cfg: cfg,
		db:  db,
	}
}

const (
	connectingStringTemplate = "postgres://%s:%s@%s:%s/%s?sslmode=disable"

	findClientByFullSurname = "SELECT id, full_name, ticket_type, passed_control_zone FROM tickets WHERE surname=$1"
	findClientBySurname     = "SELECT id, full_name, ticket_type, passed_control_zone  FROM tickets WHERE surname LIKE $1"
	updateQuery             = "UPDATE tickets SET passed_control_zone = true WHERE id = $1 RETURNING id, full_name, ticket_type, passed_control_zone"
)

func NewDatabaseConnection(cfg configs.DBConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(connectingStringTemplate, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("repository.NewDatabaseConnection().Open(): failed to conncect to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("repository.NewDatabaseConnection().Ping(): failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	log.Println("Connected to DB successfully!")

	return db, nil
}

func (tr *TicketsRepo) SearchBySurname(ctx context.Context, surname string) ([]models.TicketResponse, error) {
	surnamePattern := surname + "%"
	rows, err := tr.db.QueryContext(ctx, findClientBySurname, surnamePattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.TicketResponse
	for rows.Next() {
		var user models.TicketResponse
		err := rows.Scan(&user.Id, &user.Name, &user.TicketType, &user.PassedControlZone)
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

func (tr *TicketsRepo) MarkAsEntered(ctx context.Context, id string) (*models.TicketResponse, error) {
	tx, err := tr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var resp models.TicketResponse
	err = tx.QueryRowContext(ctx, updateQuery, id).Scan(&resp.Id, &resp.Name, &resp.TicketType, &resp.PassedControlZone)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (tr *TicketsRepo) CheckCountOfSurnames(ctx context.Context, surname string) (int64, error) {
	res, err := tr.db.ExecContext(ctx, findClientByFullSurname, surname)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
