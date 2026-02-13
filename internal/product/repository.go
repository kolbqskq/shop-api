package product

import (
	"context"

	"github.com/google/uuid"
)

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Save(ctx context.Context, reserved int, id uuid.UUID, version int64) error {
	return nil

}

func (r *Repository) Load() error {
	return nil
}

/*
	 	query := `
		   	UPDATE products
		   	SET
		   	 reserved= reserved + @reserved,
		   	 version = version + 1
		   	WHERE id = @id AND version = @version AND stock - reserved >= @reserved;
		   	`
		args := pgx.NamedArgs{
			"reserved": reserved,
			"id":       id,
			"version":  version,
		}
		if res, err := r.dbPool.Exec(ctx, query, args); err != nil {
			return err
		}
		if res.RowsAffected == 0 {
			return errors.New("race")
		}
		return nil
*/
