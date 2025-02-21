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

	findClientByFullSurname = "SELECT ticketno, full_name, ticket_type, passed_control_zone FROM tickets WHERE surname=$1"
	findClientBySurname     = "SELECT ticketno, full_name, ticket_type, passed_control_zone  FROM tickets WHERE surname LIKE $1"
	updateQuery             = "UPDATE tickets SET passed_control_zone = true WHERE ticketno = $1 RETURNING ticketno, full_name, ticket_type, passed_control_zone"
	searchById              = "SELECT ticketno, full_name, ticket_type, passed_control_zone FROM tickets WHERE ticketno=$1"
	sellTicket              = "INSERT INTO tickets (surname, full_name, ticket_type, seller_name, ticket_price, actual_ticket_price, ticketno) VALUES ($1, $2, $3, $4, $5, $6, (SELECT COALESCE(MAX(ticketNo), 0) + 1 FROM tickets)) RETURNING ticketNo"
	updateSellersTable      = "INSERT INTO ticket_sellers (ticket_id, seller_tag, seller_tg_id) VALUES ($1, $2, $3)"
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
	rows, err := tr.db.QueryContext(ctx, findClientBySurname, surname)
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
		_ = tx.Rollback()
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

func (tr *TicketsRepo) SearchById(ctx context.Context, id string) (*models.TicketResponse, error) {
	var resp models.TicketResponse
	err := tr.db.QueryRowContext(ctx, searchById, id).Scan(&resp.Id, &resp.Name, &resp.TicketType, &resp.PassedControlZone)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (tr *TicketsRepo) SellTicket(ctx context.Context, client models.ClientData, seller string, clientSurname string, actualPrice int) (int64, error) {
	tx, err := tr.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, sellTicket, clientSurname, client.FIO, client.TicketType, seller, client.Price, actualPrice).Scan(&id)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (tr *TicketsRepo) UpdateSellersTable(ctx context.Context, ticketId, sellerId int64, seller string) error {
	tx, err := tr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, updateSellersTable, ticketId, seller, sellerId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
