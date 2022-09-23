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
	err := d.db.NewSelect().Model(ment).Where("ID = ?", key).Relation("Labs").Relation("DoneLabs").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ment, nil
}

func (d *_db) GetMentorByTag(ctx context.Context, key string) (*Mentor, error) {
	ment := new(Mentor)
	err := d.db.NewSelect().Model(ment).Where("TAG = ?", key).Relation("Labs").Relation("DoneLabs").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ment, nil
}

func (d *_db) AddMentor(ctx context.Context, ment *Mentor) error {
	d.db.NewRaw("SELECT AVG(LOAD) FROM MENTORS").Scan(ctx, &ment.Load)
	_, err := d.db.NewInsert().Model(ment).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	err = d.db.NewSelect().Model(ment).Where("TAG = ?", ment.Tag).Scan(ctx)
	return err
}

func (d *_db) RemoveMentor(ctx context.Context, ment *Mentor) error {
	_, err := d.db.NewDelete().Model(ment).WhereOr("MMST_ID = ?", ment.MmstID).WhereOr("TAG = ?", ment.Tag).Exec(ctx)
	return err
}

func (d *_db) UpdateMentor(ctx context.Context, ment *Mentor) error {
	_, err := d.db.NewUpdate().Model(ment).WherePK().Exec(ctx)
	return err
}

func (d *_db) CheckMentor(ctx context.Context, ment *Mentor) (bool, error) {
	return d.db.NewSelect().Model(ment).WhereOr("MMST_ID = ?", ment.MmstID).WhereOr("TAG = ?", ment.Tag).Exists(ctx)
}

func (d *_db) GetStudents(ctx context.Context) ([]Student, error) {
	var res []Student
	err := d.db.NewSelect().Model(&res).Relation("Labs").Relation("DoneLabs").Scan(ctx)
	return res, err
}

func (d *_db) GetStudentById(ctx context.Context, key int64) (*Student, error) {
	stud := new(Student)
	err := d.db.NewSelect().Model(stud).Where("ID = ?", key).Relation("Labs").Relation("DoneLabs").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stud, nil
}

func (d *_db) GetStudentByTag(ctx context.Context, key string) (*Student, error) {
	stud := new(Student)
	err := d.db.NewSelect().Model(stud).Where("TAG = ?", key).Relation("Labs").Relation("DoneLabs").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stud, nil
}

func (d *_db) AddStudent(ctx context.Context, stud *Student) error {
	_, err := d.db.NewInsert().Model(stud).On("CONFLICT DO NOTHING").Exec(ctx)
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
	selectedMentor.Load += 2
	d.db.NewUpdate().Model(&selectedMentor).Where("ID = ?", selectedMentor.ID).Column("load").Exec(ctx)
	lab.MentorID = selectedMentor.ID
	_, err = d.db.NewInsert().Model(lab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return selectedMentor, err
	}
	return selectedMentor, err
}

func (d *_db) UpdateLab(ctx context.Context, lab *Lab) error {
	_, err := d.db.NewUpdate().Model(lab).WherePK().Exec(ctx)
	return err
}

func (d *_db) GetLabPK(ctx context.Context, lab *Lab) error {
	return d.db.NewSelect().Model(lab).Where("ID = ?", lab.ID).Scan(ctx)
}

func (d *_db) GetDoneLabPK(ctx context.Context, lab *DoneLab) error {
	return d.db.NewSelect().Model(lab).Where("ID = ?", lab.ID).Scan(ctx)
}

func (d *_db) GetLabs(ctx context.Context, lab *Lab) ([]Lab, error) {
	var labs []Lab
	err := d.db.NewSelect().Model(&labs).Where("URL = ?", lab.Url).Where("STUDENT_ID = ?", lab.StudentID).Scan(ctx)
	return labs, err
}

func (d *_db) GetStudentLabs(ctx context.Context, stud *Student) ([]Lab, error) {
	var labs []Lab
	err := d.db.NewSelect().Model(&labs).Where("STUDENT_ID = ?", stud.ID).Scan(ctx)
	return labs, err
}

func (d *_db) FinishLab(ctx context.Context, lab *Lab) error {
	var selectedMentor Mentor
	err := d.db.NewSelect().Model(&selectedMentor).Where("ID = ?", lab.MentorID).Column("load").Scan(ctx)
	if err != nil {
		return err
	}
	selectedMentor.Load -= 1
	d.db.NewUpdate().Model(&selectedMentor).Where("ID = ?", lab.MentorID).Column("load").Exec(ctx)
	doneLab := DoneLab{
		ID:        lab.ID,
		Url:       lab.Url,
		StudentID: lab.StudentID,
		MentorID:  lab.MentorID,
		Number:    lab.Number,
	}
	_, err = d.db.NewInsert().Model(&doneLab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	_, err = d.db.NewDelete().Model(lab).WherePK().Exec(ctx)
	return err
}

func (d *_db) UnfinishLab(ctx context.Context, lab *DoneLab) error {
	var selectedMentor Mentor
	err := d.db.NewSelect().Model(&selectedMentor).Where("ID = ?", lab.MentorID).Column("load").Scan(ctx)
	if err != nil {
		return err
	}
	selectedMentor.Load += 1
	d.db.NewUpdate().Model(&selectedMentor).Where("ID = ?", lab.MentorID).Column("load").Exec(ctx)
	undoneLab := Lab{
		ID:        lab.ID,
		Url:       lab.Url,
		StudentID: lab.StudentID,
		MentorID:  lab.MentorID,
		Number:    lab.Number,
	}
	_, err = d.db.NewInsert().Model(&undoneLab).On("CONFLICT DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	_, err = d.db.NewDelete().Model(lab).WherePK().Exec(ctx)
	return err
}

func (d *_db) CheckAdmin(ctx context.Context, adm *Admin) (bool, error) {
	cnt, err := d.db.NewSelect().Model((*Admin)(nil)).WhereOr("MMST_ID = ?", adm.MmstID).WhereOr("TAG = ?", adm.Tag).Count(ctx)
	return cnt > 0, err
}
