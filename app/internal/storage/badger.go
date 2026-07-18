package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"scan/internal/models"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB
}

func Open(path string) (*DB, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	opts := badger.DefaultOptions(path)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) CreateScan(scan *models.Scan) error {
	key := scanKey(scan.UUID)

	data, err := json.Marshal(scan)
	if err != nil {
		return fmt.Errorf("failed to marshal scan: %w", err)
	}

	err = d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
	if err != nil {
		return fmt.Errorf("failed to store scan: %w", err)
	}

	return nil
}

func (d *DB) GetScan(uuid string) (*models.Scan, error) {
	key := scanKey(uuid)

	var scan models.Scan
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &scan)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("scan not found: %s", uuid)
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	return &scan, nil
}

func (d *DB) UpdateScan(scan *models.Scan) error {
	key := scanKey(scan.UUID)

	data, err := json.Marshal(scan)
	if err != nil {
		return fmt.Errorf("failed to marshal scan: %w", err)
	}

	err = d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
	if err != nil {
		return fmt.Errorf("failed to update scan: %w", err)
	}

	return nil
}

func scanKey(uuid string) []byte {
	return []byte("scan:" + uuid)
}

func (d *DB) ListPendingScans() ([]*models.Scan, error) {
	var scans []*models.Scan

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("scan:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var scan models.Scan
				if err := json.Unmarshal(val, &scan); err != nil {
					return err
				}
				if scan.Status == models.StatusPending {
					scans = append(scans, &scan)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list pending scans: %w", err)
	}

	return scans, nil
}

func (d *DB) CleanupOldScans(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	return d.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("scan:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var scan models.Scan
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &scan)
			})
			if err != nil {
				continue
			}
			if scan.Start.Before(cutoff) {
				if err := txn.Delete(item.Key()); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
