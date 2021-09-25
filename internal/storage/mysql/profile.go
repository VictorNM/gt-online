package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/victornm/gtonline/internal/profile"
	"github.com/victornm/gtonline/internal/storage"
)

type (
	regularUser struct {
		Email       string         `db:"email"`
		Birthdate   sql.NullTime   `db:"birthdate"`
		Sex         sql.NullString `db:"sex"`
		CurrentCity sql.NullString `db:"current_city"`
		Hometown    sql.NullString `db:"hometown"`
	}

	interest struct {
		Email    string `db:"email"`
		Interest string `db:"interest"`
	}

	attend struct {
		Email         string        `db:"email"`
		SchoolName    string        `db:"school_name"`
		YearGraduated sql.NullInt32 `db:"year_graduated"`
	}

	employment struct {
		Email        string `db:"email"`
		EmployerName string `db:"employer_name"`
		JobTitle     string `db:"job_title"`
	}
)

func (s *Storage) GetProfile(ctx context.Context, email string) (*profile.Profile, error) {
	u, err := s.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	ru := new(regularUser)
	if err := s.db.Get(ru, `SELECT birthdate, sex, current_city, hometown FROM regular_users WHERE email=?`, email); err != nil {
		if err == sql.ErrNoRows {
			return &profile.Profile{
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
			}, nil
		}

		return nil, fmt.Errorf("query regular_users: %v", err)
	}

	var interests []interest
	if err := s.db.Select(&interests, `SELECT * FROM interests WHERE email=?`, email); err != nil {
		return nil, fmt.Errorf("query interests: %v", err)
	}

	var attends []attend
	if err := s.db.Select(&attends, `SELECT * FROM attends WHERE email=?`, email); err != nil {
		return nil, fmt.Errorf("query attends: %v", err)
	}

	var employments []employment
	if err := s.db.Select(&employments, `SELECT * FROM employments WHERE email=?`, email); err != nil {
		return nil, fmt.Errorf("query employments: %v", err)
	}

	p := &profile.Profile{
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Sex:         ru.Sex.String,
		Birthdate:   ru.Birthdate.Time,
		CurrentCity: ru.CurrentCity.String,
		Hometown:    ru.Hometown.String,
	}

	for _, i := range interests {
		p.Interests = append(p.Interests, i.Interest)
	}

	for _, a := range attends {
		p.Education = append(p.Education, profile.Attend{
			School:        a.SchoolName,
			YearGraduated: int(a.YearGraduated.Int32),
		})
	}

	for _, e := range employments {
		p.Professional = append(p.Professional, profile.Employment{
			Employer: e.EmployerName,
			JobTitle: e.JobTitle,
		})
	}

	return p, nil
}

func (s *Storage) UpdateProfile(ctx context.Context, req profile.UpdateProfileRequest) (err error) {
	return updateProfile(ctx, s.db, req)
}

func updateProfile(ctx context.Context, tx *sqlx.DB, req profile.UpdateProfileRequest) error {
	if err := updateRegularUser(ctx, tx, req); err != nil {
		return fmt.Errorf("update regular_users: %v", err)
	}

	if err := replaceInterests(ctx, tx, req); err != nil {
		return fmt.Errorf("replace interests: %v", err)
	}

	if err := replaceAttends(ctx, tx, req); err != nil {
		return fmt.Errorf("replace attends: %w", err)
	}

	if err := replaceEmployments(ctx, tx, req); err != nil {
		return fmt.Errorf("replace employments: %w", err)
	}

	return nil
}

func newRegularUser(req profile.UpdateProfileRequest) regularUser {
	var row regularUser

	row.Email = req.Email

	if !req.Birthdate.IsZero() {
		row.Birthdate = sql.NullTime{Time: req.Birthdate, Valid: true}
	}

	if req.Sex != "" {
		row.Sex = sql.NullString{String: req.Sex, Valid: true}
	}

	if req.CurrentCity != "" {
		row.CurrentCity = sql.NullString{String: req.CurrentCity, Valid: true}
	}

	if req.Hometown != "" {
		row.Hometown = sql.NullString{String: req.Hometown, Valid: true}
	}

	return row
}

func updateRegularUser(ctx context.Context, tx *sqlx.DB, req profile.UpdateProfileRequest) error {
	row := newRegularUser(req)
	stmt := `UPDATE regular_users SET birthdate=:birthdate, sex=:sex, current_city=:current_city, hometown=:hometown WHERE email=:email;`
	_, err := tx.NamedExecContext(ctx, stmt, row)
	return err
}

func replaceInterests(ctx context.Context, tx *sqlx.DB, req profile.UpdateProfileRequest) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM interests WHERE email=?`, req.Email); err != nil {
		return fmt.Errorf("delete interests: %v", err)
	}

	var rows []*interest
	for _, i := range req.Interests {
		rows = append(rows, &interest{
			Email:    req.Email,
			Interest: i,
		})
	}

	if len(rows) == 0 {
		return nil
	}

	if _, err := tx.NamedExecContext(ctx,
		`INSERT INTO interests (email, interest) VALUES (:email, :interest)`,
		rows); err != nil {
		return fmt.Errorf("insert interests: %v", err)
	}

	return nil
}

func replaceAttends(ctx context.Context, tx *sqlx.DB, req profile.UpdateProfileRequest) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM attends WHERE email=?`, req.Email); err != nil {
		return fmt.Errorf("delete attends: %v", err)
	}

	var rows []*attend
	for _, a := range req.Education {
		row := &attend{
			Email:      req.Email,
			SchoolName: a.School,
		}
		if a.YearGraduated != 0 {
			row.YearGraduated = sql.NullInt32{Int32: int32(a.YearGraduated), Valid: true}
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	_, err := tx.NamedExecContext(ctx, `INSERT INTO attends (email, school_name, year_graduated) VALUES (:email, :school_name, :year_graduated)`, rows)
	if err == nil {
		return nil
	}

	if isErrForeignKeyConstraint(err) {
		return storage.ErrInvalidArgument
	}

	return fmt.Errorf("insert attends: %v", err)
}

func replaceEmployments(ctx context.Context, tx *sqlx.DB, req profile.UpdateProfileRequest) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM employments WHERE email=?`, req.Email); err != nil {
		return fmt.Errorf("delete employments: %v", err)
	}

	var rows []*employment
	for _, e := range req.Professional {
		rows = append(rows, &employment{
			Email:        req.Email,
			EmployerName: e.Employer,
			JobTitle:     e.JobTitle,
		})
	}

	if len(rows) == 0 {
		return nil
	}

	_, err := tx.NamedExecContext(ctx, `INSERT INTO employments (email, employer_name, job_title) VALUES (:email, :employer_name, :job_title)`, rows)

	if err == nil {
		return nil
	}

	if isErrForeignKeyConstraint(err) {
		return storage.ErrInvalidArgument
	}

	return fmt.Errorf("insert attends: %v", err)
}

func (s *Storage) DeleteUser(ctx context.Context, email string) error {
	stmt := `DELETE FROM users WHERE email=?`
	_, err := s.db.ExecContext(ctx, stmt, email)
	return err
}

func (s *Storage) isEmailExist(ctx context.Context, table string, email string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT EXISTS(SELECT email FROM %s WHERE email=?)`, table)

	r, err := s.db.QueryContext(ctx, stmt, email)
	if err != nil {
		return false, fmt.Errorf("query: %v", err)
	}
	var exist bool
	for r.Next() {
		if err := r.Scan(&exist); err != nil {
			return false, fmt.Errorf("scan: %v", err)
		}
	}

	return exist, nil
}

func (s *Storage) ListSchools(ctx context.Context) ([]profile.School, error) {
	var schools []profile.School

	stmt := `SELECT school_name, type FROM schools;`
	if err := s.db.SelectContext(ctx, &schools, stmt); err != nil {
		return nil, fmt.Errorf("query schools: %v", err)
	}

	return schools, nil
}

func (s *Storage) ListEmployers(ctx context.Context) ([]profile.Employer, error) {
	var employers []profile.Employer

	stmt := `SELECT employer_name FROM employers;`
	if err := s.db.SelectContext(ctx, &employers, stmt); err != nil {
		return nil, fmt.Errorf("query employers: %v", err)
	}

	return employers, nil
}

func isErrForeignKeyConstraint(err error) bool {
	if err == nil {
		return false
	}

	if e := new(mysql.MySQLError); errors.As(err, &e) {
		return e.Number == 1452
	}

	return false
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}

	if e := new(mysql.MySQLError); errors.As(err, &e) {
		return e.Number == 1062
	}

	return false
}
