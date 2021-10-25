package mysql

import "gorm.io/gorm"

type DecoderState struct {
	gorm.Model
	ThingID string `gorm:"primaryKey;size:36"`
	Key     string `gorm:"primaryKey;size:36"`
	Value   []byte
}

func (d *DB) GetState(thingID, key string) ([]byte, error) {
	var state DecoderState
	err := d.db.Where("thing_id = ? AND `key` = ?", thingID, key).First(&state).Error
	return state.Value, err
}

func (d *DB) SetState(thingId, key string, value []byte) error {
	state := DecoderState{
		ThingID: thingId,
		Key:     key,
		Value:   value,
	}
	return d.db.Create(state).Error
}
