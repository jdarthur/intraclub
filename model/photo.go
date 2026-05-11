package model

import (
	"context"
	"errors"
	"fmt"

	"intraclub/database"
)

type PhotoType int

const (
	PhotoTypePng PhotoType = iota
	PhotoTypeJpg
	PhotoTypeJpeg
	PhotoTypeGif
	PhotoTypeWebP
	PhotoTypeInvalid
)

func (t PhotoType) String() string {
	switch t {
	case PhotoTypePng:
		return "png"
	case PhotoTypeJpg:
		return "jpg"
	case PhotoTypeJpeg:
		return "jpeg"
	case PhotoTypeWebP:
		return "webp"
	case PhotoTypeGif:
		return "gif"
	default:
		return "unknown"
	}
}

func (t PhotoType) Valid() bool {
	return t < PhotoTypeInvalid
}

type PhotoId database.RecordId

func (id PhotoId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id PhotoId) String() string {
	return id.RecordId().String()
}

func (id PhotoId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id PhotoId) UnmarshalJSON(data []byte) error {
	rid := id.RecordId()
	return (*database.RecordId)(&rid).UnmarshalJSON(data)
}

type Photo struct {
	ID       PhotoId
	Owner    UserId
	AltText  string
	Contents []byte
	FileType PhotoType
}

func (p *Photo) GetOwner() database.RecordId {
	return p.Owner.RecordId()
}

func NewPhoto() *Photo {
	return &Photo{}
}

func (p *Photo) SetOwner(recordId database.RecordId) {
	p.Owner = UserId(recordId)
}

func (p *Photo) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{p.Owner.RecordId(), database.SysAdminRecordId}
}

func (p *Photo) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.EveryoneRecordId}
}

func (p *Photo) StaticallyValid() error {
	if len(p.Contents) == 0 {
		return errors.New("photo has no content")
	}

	if !p.FileType.Valid() {
		return fmt.Errorf("photo has invalid file type %d", p.FileType)
	}
	return nil
}

func (p *Photo) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &User{}, p.Owner.RecordId())
	if err != nil {
		return err
	}
	return nil
}

func (p *Photo) Type() string {
	return "photo"
}

func (p *Photo) GetId() database.RecordId {
	return p.ID.RecordId()
}

func (p *Photo) SetId(id database.RecordId) {
	p.ID = PhotoId(id)
}

func (p *Photo) BlankRecord() database.CrudRecord {
	return new(Photo)
}
