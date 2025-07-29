package engine

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"
	"tprgdb/engine/api/probabilistic"
	"tprgdb/engine/api/scans"
	"tprgdb/engine/api/tokenbucket"
	"tprgdb/engine/api/user"
	"tprgdb/engine/constants"
	"tprgdb/engine/core"
	externalconfiguration "tprgdb/engine/externalConfiguration"
	itemPackage "tprgdb/engine/item"
)

const ()

type KVEngine struct {
	dbName        string
	core          *core.Core
	Probabilistic *probabilistic.ProbabilisticAPI
	tokenBucket   *tokenbucket.TokenBucket
	user          *user.User
}

func NewKVDatabase(dbName string, username, password string) (*KVEngine, error) {
	if dbName == "" {
		dbName = "default"
	}
	// Check if the database folder exists
	folderPath := path.Join("dbs", dbName)
	if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
		return nil, errors.New("database already exist: " + dbName)
	}
	// Create the database folder
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, err
	}

	user, err := user.NewUser(username, password)
	if err != nil {
		return nil, err
	}

	engine, err := newKVEngine(dbName, user)
	if err != nil {
		return nil, err
	}
	// Insert user into the database

	if err = engine.core.Set(constants.USER_PREFIX+user.Username, user.SerializeToBytes()); err != nil {
		return nil, err
	}
	return engine, nil
}

func Login(dbName string, username, password string) (*KVEngine, error) {
	if dbName == "" {
		dbName = "default"
	}
	// Check if the database folder exists
	folderPath := path.Join("dbs", dbName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, errors.New("database does not exist: " + dbName)
	}

	checkUser, err := user.NewUser(username, password)
	if err != nil {
		return nil, err
	}

	engine, err := newKVEngine(dbName, checkUser)
	if err != nil {
		return nil, err
	}

	item, err := engine.core.Get(constants.USER_PREFIX + checkUser.Username)
	if err != nil {
		return nil, err
	}
	if item == nil || item.IsDeleted() {
		return nil, errors.New("user does not exist: " + username)
	}
	existingUser := user.NewEmptyUser()
	existingUser.DeserializeFromBytes(item.GetValue())
	if !existingUser.Equals(checkUser) {
		return nil, errors.New("invalid password")
	}
	return engine, nil
}

func newKVEngine(dbName string, user *user.User) (*KVEngine, error) {
	if dbName == "" {
		dbName = "default"
	}
	// Check if the database folder exists
	folderPath := path.Join("dbs", dbName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, errors.New("database does not exist: " + folderPath)
	}

	// Read or create the external configuration file
	config, err := getConfig(folderPath, dbName)
	if err != nil {
		return nil, err
	}

	core, err := core.NewCore(config, dbName)
	if err != nil {
		return nil, err
	}

	probabilisticAPI := probabilistic.NewProbabilisticAPI(core)
	tokenBucket := tokenbucket.NewTokenBucket(core, config.TokenBucket.RefreshDuration, config.TokenBucket.TokenCount)

	return &KVEngine{
		dbName:        dbName,
		core:          core,
		Probabilistic: probabilisticAPI,
		tokenBucket:   tokenBucket,
		user:          user,
	}, nil
}

func getConfig(folderPath string, dbName string) (*externalconfiguration.ExternalConfiguration, error) {
	configFilePath := path.Join(folderPath, constants.CONFIG_FILENAME)
	var config *externalconfiguration.ExternalConfiguration
	if configFile, err := os.Open(configFilePath); err == nil {
		defer configFile.Close()
		config, err = externalconfiguration.ReadExternalConfiguration(configFile, dbName)
		if err != nil {
			return nil, err
		}
	} else {
		config = &externalconfiguration.ExternalConfiguration{}
		config.Validate(dbName)
		configFile, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		defer configFile.Close()
		if err := externalconfiguration.WriteExternalConfiguration(configFile, config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (engine *KVEngine) GetDBName() string {
	return engine.dbName
}

func (engine *KVEngine) Get(key string) (*itemPackage.DBItem, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	item, err := engine.core.Get(constants.DATA_PREFIX + key)
	if err != nil {
		return nil, err
	}
	if item == nil || item.IsDeleted() {
		return nil, nil
	}
	timestamp := time.Unix(int64(item.GetTimestamp()), 0)
	return &itemPackage.DBItem{
		Key:       item.GetKey()[constants.PREFIX_SIZE:],
		Value:     item.GetValue(),
		Timestamp: &timestamp,
	}, nil
}

func (engine *KVEngine) Set(key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if value == nil {
		return errors.New("value cannot be nil")
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return err
	}
	err := engine.core.Set(constants.DATA_PREFIX+key, value)
	if err != nil {
		return err
	}
	return nil
}

func (engine *KVEngine) Delete(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return err
	}
	err := engine.core.Delete(constants.DATA_PREFIX + key)
	if err != nil {
		return err
	}
	return nil
}

func (engine *KVEngine) GetMany(keys []string) ([]*itemPackage.DBItem, error) {
	if len(keys) == 0 {
		return nil, errors.New("keys cannot be empty")
	}
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("key at index %d cannot be empty", i)
		}
		prefixedKeys[i] = constants.DATA_PREFIX + key // Prefix each key with constants.DATA_PREFIX
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	items, err := engine.core.GetMany(prefixedKeys)
	if err != nil {
		return nil, err
	}

	result := make([]*itemPackage.DBItem, len(items))
	for i, item := range items {
		if item == nil || item.IsDeleted() {
			items[i] = nil
			continue
		}
		timestamp := time.Unix(int64(item.GetTimestamp()), 0)
		result[i] = &itemPackage.DBItem{
			Key:       item.GetKey()[constants.PREFIX_SIZE:],
			Value:     item.GetValue(),
			Timestamp: &timestamp,
		}
	}
	return result, nil
}

func (engine *KVEngine) GetUsername() string {
	if engine.user == nil {
		return ""
	}
	return engine.user.Username
}

func (engine *KVEngine) RangeScan(start, end string, pageSize int) (scans.ICursor, error) {
	if start == "" || end == "" {
		return nil, errors.New("start and end keys cannot be empty")
	}
	if start > end {
		return nil, errors.New("start key cannot be greater than end key")
	}
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	return scans.NewRangeScanCursor(engine.core, start, end, pageSize), nil
}

func (engine *KVEngine) PrefixScan(prefix string, pageSize int) (scans.ICursor, error) {
	if prefix == "" {
		return nil, errors.New("prefix cannot be empty")
	}
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	return scans.NewPrefixScanCursor(engine.core, prefix, pageSize), nil
}

func (engine *KVEngine) RangeIterate(start, end string) (scans.IIterator, error) {
	if start == "" || end == "" {
		return nil, errors.New("start and end keys cannot be empty")
	}
	if start > end {
		return nil, errors.New("start key cannot be greater than end key")
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	return scans.NewRangeIterator(engine.core, start, end), nil
}

func (engine *KVEngine) PrefixIterate(prefix string) (scans.IIterator, error) {
	if prefix == "" {
		return nil, errors.New("prefix cannot be empty")
	}
	if err := engine.tokenBucket.Validate(engine.user.Username); err != nil {
		return nil, err
	}

	return scans.NewPrefixIterator(engine.core, prefix), nil
}

func (engine *KVEngine) Validate(tableName string) (bool, error) {
	return engine.core.ValidateSSTable(tableName)
}

func (engine *KVEngine) Shutdown() error {
	if engine.core != nil {
		if err := engine.core.Shutdown(); err != nil {
			return err
		}
	}
	return nil
}

func (engine *KVEngine) SetUser(newUser *user.User) error {
	return engine.core.Set(constants.USER_PREFIX+newUser.Username, newUser.SerializeToBytes())
}
