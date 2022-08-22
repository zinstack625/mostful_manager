package database

import "github.com/uptrace/bun"

type Mentor struct {
	bun.BaseModel `bun:"table:mentors"`
	ID            int64  `bun:",pk,autoincrement"`
	Tag           string `bun:",pk"`
	Load          int64
	Labs          []*Lab     `bun:"rel:has-many,join:id=mentor_id"`
	DoneLabs      []*DoneLab `bun:"rel:has-many,join:id=mentor_id"`
}

type Student struct {
	bun.BaseModel `bun:"table:students"`
	ID            int64      `bun:",pk,autoincrement"`
	Tag           string     `bun:",pk"`
	Labs          []*Lab     `bun:"rel:has-many,join:id=student_id"`
	DoneLabs      []*DoneLab `bun:"rel:has-many,join:id=student_id"`
}

type Lab struct {
	bun.BaseModel `bun:"table:labs"`
	ID            int64 `bun:",pk,autoincrement"`
	Url           string
	StudentID     int64
	MentorID      int64
}

type DoneLab struct {
	bun.BaseModel `bun:"table:done_labs"`
	ID            int64 `bun:",pk,autoincrement"`
	Url           string
	StudentID     int64
	MentorID      int64
}

type Admin struct {
	bun.BaseModel `bun:"table:admins"`
	ID            int64  `bun:",pk,autoincrement"`
	Tag           string `bun:",pk"`
}
