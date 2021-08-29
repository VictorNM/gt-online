package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/user"
)

func (s *Storage) UpdateProfile(ctx context.Context, req user.UpdateProfileRequest) error {
	exist, err := s.isEmailExist(ctx, "users", req.Email)
	if err != nil {
		return err
	}

	if !exist {
		return gterr.ErrNotFound
	}

	exist, err = s.isEmailExist(ctx, "regular_users", req.Email)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if exist {
		return s.updateProfile(ctx, tx, req)
	}

	return s.insertProfile(ctx, tx, req)
}

type (
	regularUser struct {
		Email       string    `db:"email"`
		Sex         string    `db:"sex"`
		Birthdate   time.Time `db:"birthdate"`
		CurrentCity string    `db:"current_city"`
		Hometown    string    `db:"hometown"`
	}

	interest struct {
		Email    string `db:"email"`
		Interest string `db:"interest"`
	}

	attend struct {
		Email        string `db:"email"`
		SchoolName   string `db:"school_name"`
		YearGraduate int    `db:"year_graduate"`
	}

	employment struct {
		Email        string `db:"email"`
		EmployerName string `db:"employer_name"`
		JobTitle     string `db:"job_title"`
	}
)

func (s *Storage) insertProfile(ctx context.Context, tx *sqlx.Tx, req user.UpdateProfileRequest) error {
	stmt := `INSERT INTO regular_users (email, sex, birthdate, current_city, hometown) VALUES(:email, :sex, :birthdate, :current_city, :hometown);`

	_, err := tx.NamedExecContext(ctx, stmt, regularUser{
		Email:       req.Email,
		Sex:         req.Sex,
		Birthdate:   req.Birthdate,
		CurrentCity: req.CurrentCity,
		Hometown:    req.Hometown,
	})
	if err != nil {
		return fmt.Errorf("insert regular_users: %v", err)
	}

	stmt = `INSERT INTO interests (email, interest) VALUES(:email, :interest);`
	var interests []*interest
	for _, i := range req.Interests {
		interests = append(interests, &interest{
			Email:    req.Email,
			Interest: i,
		})
	}
	if _, err := tx.NamedExecContext(ctx, stmt, interests); err != nil {
		return fmt.Errorf("insert interests: %v", err)
	}

	stmt = `INSERT INTO attends (email, school_name, year_graduated) VALUES (:email, :school_name, :year_graduated)`
	var attends []*attend
	for _, a := range req.Education {
		attends = append(attends, &attend{
			Email:        req.Email,
			SchoolName:   a.School,
			YearGraduate: a.YearGraduated,
		})
	}
	if _, err := tx.NamedExecContext(ctx, stmt, attends); err != nil {
		return fmt.Errorf("insert attends: %v", err)
	}

	stmt = `INSERT INTO employments (email, employer_name, job_title) VALUES (:email, :employer_name, :job_title)`
	var employments []*employment
	for _, e := range req.Professional {
		employments = append(employments, &employment{
			Email:        req.Email,
			EmployerName: e.Employer,
			JobTitle:     e.JobTitle,
		})
	}
	if _, err := tx.NamedExecContext(ctx, stmt, employments); err != nil {
		return fmt.Errorf("insert employments: %v", err)
	}

	return nil
}

func (s *Storage) updateProfile(ctx context.Context, tx *sqlx.Tx, req user.UpdateProfileRequest) error {
	return nil
}

func (s *Storage) isEmailExist(ctx context.Context, table string, email string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT EXISTS(SELECT email FROM %s WHERE email=?)`, table)

	r, err := s.db.QueryContext(ctx, stmt, email)
	if err != nil {
		return false, err
	}
	var exist bool
	if err := r.Scan(&exist); err != nil {
		return false, err
	}
	return exist, nil
}
