package repository

import (
	"context"

	"github.com/qRe0/afterparty-bot/internal/models"
)

type TicketsRepoInterface interface {
	SearchByFullSurname(ctx context.Context, surname string) (*models.TicketResponse, error)
	SearchBySurnamePart(ctx context.Context, surnamePart string) ([]models.TicketResponse, error)
	SearchByID(ctx context.Context, id string) (*models.TicketResponse, error)
	MarkAsEntered(ctx context.Context, id string) (*models.TicketResponse, error)
}
