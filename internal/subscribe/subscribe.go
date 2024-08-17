package subscribe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/internal/sender"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Notify(db *pgxpool.Pool, houseId int, flatId int64) {
	rows, err := db.Query(context.Background(), `Select email  from subscribe where  houseId = $1`, houseId)
	if err != nil {
		log.Println("Notify error: cant connect to db")
		return
	}
	defer rows.Close()
	send := sender.New()
	for rows.Next() {
		var email string
		_ = rows.Scan(&email)
		go func() {
			i := 0
			var err error = errors.New("---")
			for err != nil && i < 15 {
				err = send.SendEmail(context.Background(), email, "new Flat with id "+fmt.Sprint(flatId))
				i++
			}
			if err != nil {
				log.Print("Error on sending notification to " + email)
			}
		}()

	}
}
