package mysql

type DecoderState struct {
	ThingID string `gorm:"primaryKey"`
	Key     string `gorm:"primaryKey"`
	Value   []byte
}

func (d *DB) GetState(thingID, key string) ([]byte, error) {
	var state DecoderState
	err := d.db.Model(&DecoderState{}).Where("thing_id = ? AND key = ?", thingID, key).Take(&state).Error
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
