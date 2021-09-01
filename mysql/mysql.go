package mysql

import (
	"context"
	"fmt"

	"github.com/connctd/connector-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Installation struct {
	gorm.Model
	ID    string `gorm:"primaryKey"`
	Token string
}

type Instance struct {
	gorm.Model
	ID             string `gorm:"primaryKey"`
	Token          string
	InstallationID string        `gorm:"REFERENCES installations(id)"`
	Installation   *Installation `gorm:"foreignKey:InstallationID;AssociationForeignKey:ID"`
}

type IDMapping struct {
	gorm.Model
	DevEUI     []byte    `gorm:"primaryKey"`
	ThingID    string    `gorm:"uniqueIndex"`
	InstanceID string    `gorm:"REFERENCES instances(id)"`
	Instance   *Instance `gorm:"foreignKey:InstanceID;AssociationForeignKey:ID"`
}

type DecoderConfig struct {
	gorm.Model
	ApplicationID uint64 `grom:"primaryKey"`
	DecoderName   string
	InstanceID    string    `gorm:"REFERENCES instances(id)"`
	Instance      *Instance `gorm:"foreignKey:InstanceID;AssociationForeignKey:ID"`
}

type DB struct {
	db *gorm.DB
}

func NewDB(dsn string) (*DB, error) {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	d := &DB{db: gdb}
	if err := d.CreateOrMigrate(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DB) CreateOrMigrate() error {
	for _, model := range []interface{}{
		&Installation{},
		&Instance{},
		&IDMapping{},
		&DecoderConfig{},
	} {
		if err := d.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to automigrate %T table: %w", model, err)
		}
	}

	return nil
}

func (d *DB) AddInstallation(ctx context.Context, req connector.InstallationRequest) (err error) {
	installation := Installation{
		ID:    req.ID,
		Token: string(req.Token),
	}
	err = d.db.WithContext(ctx).Create(installation).Error
	return
}

func (d *DB) AddInstance(ctx context.Context, req connector.InstantiationRequest) (err error) {
	instance := Instance{
		ID:             req.ID,
		Token:          string(req.Token),
		InstallationID: req.InstallationID,
	}
	err = d.db.WithContext(ctx).Create(instance).Error
	return
}

func (d *DB) GetInstallationToken(installationId string) (connector.InstallationToken, error) {
	var installation Installation
	err := d.db.Model(&Installation{}).Where("id = ?", installationId).Take(&installation).Error
	return connector.InstallationToken(installation.Token), err
}

func (d *DB) GetInstanceToken(instanceId string) (connector.InstantiationToken, error) {
	var instance Instance
	err := d.db.Model(&Instance{}).Where("id = ?", instanceId).Take(&instance).Error
	return connector.InstantiationToken(instance.Token), err
}

func (d *DB) GetInstance(instanceId string) (connector.InstantiationRequest, error) {
	var instance Instance
	err := d.db.Model(&Instance{}).Where("id = ?", instanceId).Take(&instance).Error
	return connector.InstantiationRequest{
		ID:             instance.ID,
		Token:          connector.InstantiationToken(instance.Token),
		InstallationID: instance.InstallationID,
	}, err
}

func (d *DB) StoreDEVUIToThingID(instanceID string, devEUI []byte, thingID string) error {
	mapping := IDMapping{
		DevEUI:     devEUI,
		ThingID:    thingID,
		InstanceID: instanceID,
	}
	return d.db.Create(mapping).Error
}

func (d *DB) MapDevEUIToThingID(instanceId string, devEUI []byte) (string, error) {
	var mapping IDMapping
	err := d.db.Model(&IDMapping{}).Where("deveui = ? AND instance_id = ?", devEUI, instanceId).Take(&mapping).Error
	return mapping.ThingID, err
}

func (d *DB) DecoderNameForApp(instanceID string, appId uint64) (string, error) {
	var config DecoderConfig
	err := d.db.Model(&DecoderConfig{}).Where("application_id = ? AND instance_id = ?", appId, instanceID).Take(&config).Error
	return config.DecoderName, err
}
