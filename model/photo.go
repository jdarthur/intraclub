package model

import (
	"errors"
	"fmt"
	"intraclub/common"
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

type PhotoId common.RecordId

func (id PhotoId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id PhotoId) String() string {
	return id.RecordId().String()
}

type Photo struct {
	ID       PhotoId
	Owner    UserId
	AltText  string
	Contents []byte
	FileType PhotoType
}

func NewPhoto() *Photo {
	return &Photo{}
}

func (p *Photo) SetOwner(recordId common.RecordId) {
	p.Owner = UserId(recordId)
}

func (p *Photo) EditableBy(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{p.Owner.RecordId(), common.SysAdminRecordId}
}

func (p *Photo) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.EveryoneRecordId}
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

func (p *Photo) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	err := common.ExistsById(db, &User{}, p.Owner.RecordId())
	if err != nil {
		return err
	}
	return nil
}

func (p *Photo) Type() string {
	return "photo"
}

func (p *Photo) GetId() common.RecordId {
	return p.ID.RecordId()
}

func (p *Photo) SetId(id common.RecordId) {
	p.ID = PhotoId(id)
}
