package mysql

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/connctd/connector-go"
	"github.com/connctd/lora-connector/lorawan/decoder"
	"github.com/connctd/restapi-go"
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
	ConfigThingID  string        `gorm:"uniqueIndex"`
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

var configThing = restapi.Thing{
	Name:         "configuration Thin",
	Manufacturer: "IoT connctd GmbH",
	DisplayType:  "loranetwork",
	Status:       restapi.StatusTypeAvailable,
	Components: []restapi.Component{
		{
			ID:            "lora",
			Name:          "LoRaWAN config",
			ComponentType: "config",
			Capabilities:  []string{"loraconfig"},
			Properties: []restapi.Property{
				{
					ID:           "url",
					Name:         "HTTP Callback URL",
					Value:        "",
					Unit:         "",
					Type:         restapi.ValueTypeString,
					PropertyType: "URL",
				},
				{
					ID:    "decoders",
					Name:  "Decoders",
					Value: "ldds75",
					Unit:  "",
					Type:  restapi.ValueTypeString,
				},
			},
			Actions: []restapi.Action{
				{
					ID:   "addmapping",
					Name: "AddMapping",
					Parameters: []restapi.ActionParameter{
						{
							Name: "ApplicationId",
							Type: restapi.ValueTypeNumber,
						},
						{
							Name: "PayloadDecoder",
							Type: restapi.ValueTypeString,
						},
					},
				},
			},
		},
	},
}

type DB struct {
	db              *gorm.DB
	connectorClient connector.Client
	host            string
}

func NewDB(dsn string, connectorClient connector.Client, host string) (*DB, error) {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	d := &DB{
		db:              gdb,
		connectorClient: connectorClient,
		host:            host,
	}
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

func (d *DB) AddInstance(ctx context.Context, req connector.InstantiationRequest) error {
	instance := Instance{
		ID:             req.ID,
		Token:          string(req.Token),
		InstallationID: req.InstallationID,
	}
	// TODO add config thing
	db := d.db.WithContext(ctx).Begin()
	defer db.Rollback()
	configThing, err := d.connectorClient.CreateThing(ctx, req.Token, configThing)
	if err != nil {
		return err
	}
	instance.ConfigThingID = configThing.ID
	err = db.WithContext(ctx).Create(instance).Error
	if err != nil {
		return err
	}
	callbackUrl := fmt.Sprintf("https://%s/lorawan/%s/%s", d.host, req.InstallationID, req.ID)
	err = d.connectorClient.UpdateThingPropertyValue(ctx, req.Token, instance.ConfigThingID, "lora", "url", callbackUrl, time.Now())
	if err != nil {
		return err
	}

	db.Commit()
	return nil
}

func (d *DB) PerformAction(ctx context.Context, req connector.ActionRequest) (*connector.ActionResponse, error) {
	var instance Instance
	err := d.db.WithContext(ctx).Model(&Instance{}).Where("config_thing_id = ?", req.ThingID).Error
	if err != nil {
		return &connector.ActionResponse{
			Status: restapi.ActionRequestStatusCanceled,
			Error:  err.Error(),
		}, err
	}
	if req.ActionID != "addmapping" || req.ComponentID != "lora" {
		return &connector.ActionResponse{
			Status: restapi.ActionRequestStatusCanceled,
			Error:  "Invalid action or component ID",
		}, nil
	}
	appIdString := req.Parameters["ApplicationId"]
	appId, err := strconv.ParseUint(appIdString, 10, 64)
	if err != nil {
		return &connector.ActionResponse{
			Status: restapi.ActionRequestStatusCanceled,
			Error:  "Invalid LoRaWAN application id",
		}, nil
	}
	decoderName := req.Parameters["PayloadDecoder"]
	dec := decoder.GetDecoder(decoderName)
	if dec == nil {
		return &connector.ActionResponse{
			Status: restapi.ActionRequestStatusCanceled,
			Error:  "Invalid decoder name",
		}, nil
	}
	config := DecoderConfig{
		ApplicationID: appId,
		DecoderName:   decoderName,
		InstanceID:    instance.ID,
	}
	err = d.db.WithContext(ctx).Create(config).Error
	if err != nil {
		return &connector.ActionResponse{
			Status: restapi.ActionRequestStatusCanceled,
			Error:  "Internal Error",
		}, err
	}
	return &connector.ActionResponse{
		Status: restapi.ActionRequestStatusCompleted,
	}, nil
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
