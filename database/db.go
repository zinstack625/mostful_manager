package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var DB _db

type _db struct {
	sqldb *sql.DB
	db    *bun.DB
}

func (d *_db) Init(conn string) {
	d.sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(conn)))
	d.db = bun.NewDB(d.sqldb, pgdialect.New())
	d.initTables()
}

func (d *_db) initTables() {
	queryCtx := context.Background()
	go d.db.NewCreateTable().Model((*Mentor)(nil)).IfNotExists().Exec(queryCtx)
	go d.db.NewCreateTable().Model((*Lab)(nil)).IfNotExists().Exec(queryCtx)
	go d.db.NewCreateTable().Model((*DoneLab)(nil)).IfNotExists().Exec(queryCtx)
	go d.db.NewCreateTable().Model((*Student)(nil)).IfNotExists().Exec(queryCtx)
	go d.db.NewCreateTable().Model((*Admin)(nil)).IfNotExists().Exec(queryCtx)
}

func (d *_db) GetMentorById(ctx context.Context, key int64) (*Mentor, error) {
	ment := new(Mentor)
	err := d.db.NewSelect().Model(ment).Where("ID = ?", key).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ment, nil
}

func (d *_db) GetMentorByTag(ctx context.Context, key string) (*Mentor, error) {
	ment := new(Mentor)
	err := d.db.NewSelect().Model(ment).Where("TAG = ?", key).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ment, nil
}

func (d *_db) AddMentor(ctx context.Context, ment *Mentor) error {
	_, err := d.db.NewInsert().Model(ment).Exec(ctx)
	if err != nil {
		return err
	}
	err = d.db.NewSelect().Model(ment).Where("TAG = ?", ment.Tag).Scan(ctx)
	return err
}

func (d *_db) UpdateMentor(ctx context.Context, ment *Mentor) error {
	_, err := d.db.NewUpdate().Model(ment).WherePK().Exec(ctx)
	return err
}

func (d *_db) GetStudentById(ctx context.Context, key int64) (*Student, error) {
	stud := new(Student)
	err := d.db.NewSelect().Model(stud).Where("ID = ?", key).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stud, nil
}

func (d *_db) GetStudentByTag(ctx context.Context, key string) (*Student, error) {
	stud := new(Student)
	err := d.db.NewSelect().Model(stud).Where("TAG = ?", key).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stud, nil
}

func (d *_db) AddStudent(ctx context.Context, stud *Student) error {
	_, err := d.db.NewInsert().Model(stud).Exec(ctx)
	if err != nil {
		return err
	}
	err = d.db.NewSelect().Model(stud).Where("TAG = ?", stud.Tag).Scan(ctx)
	return err
}

func (d *_db) UpdateStudent(ctx context.Context, stud *Student) error {
	_, err := d.db.NewUpdate().Model(stud).WherePK().Exec(ctx)
	return err
}

func (d *_db) AddLab(ctx context.Context, lab *Lab) (Mentor, error) {
	var selectedMentor Mentor
	err := d.db.NewSelect().Model(&selectedMentor).Order("load asc").Limit(1).Scan(ctx)
	if err != nil {
		return selectedMentor, err
	}
	selectedMentor.Load++
	d.db.NewUpdate().Model(&selectedMentor).WherePK().Exec(ctx)
	lab.MentorID = selectedMentor.ID
	_, err = d.db.NewInsert().Model(lab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return selectedMentor, err
	}
	err = d.db.NewSelect().Model(lab).Where("URL = ?", lab.Url).Scan(ctx)
	return selectedMentor, err
}

func (d *_db) UpdateLab(ctx context.Context, lab *Lab) error {
	_, err := d.db.NewUpdate().Model(lab).WherePK().Exec(ctx)
	return err
}

func (d *_db) FinishLab(ctx context.Context, lab *Lab) error {
	doneLab := DoneLab{
		ID:        lab.ID,
		Url:       lab.Url,
		StudentID: lab.StudentID,
		MentorID:  lab.MentorID,
	}
	_, err := d.db.NewInsert().Model(&doneLab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	_, err = d.db.NewDelete().Model(lab).WherePK().Exec(ctx)
	return err
}

func (d *_db) UnfinishLab(ctx context.Context, lab *DoneLab) error {
	undoneLab := Lab{
		ID:        lab.ID,
		Url:       lab.Url,
		StudentID: lab.StudentID,
		MentorID:  lab.MentorID,
	}
	_, err := d.db.NewInsert().Model(&undoneLab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	_, err = d.db.NewDelete().Model(lab).WherePK().Exec(ctx)
	return err
}
