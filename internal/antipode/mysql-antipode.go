package antipode

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// It assumes by default that the table where the queries will be executed has
// exactly two columns called k and value
type MySQL struct {
	db        *sql.DB
	datastore string
}

func CreateMySQL(host string, port string, user string, password string, datastore string) MySQL {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, datastore)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	return MySQL{db, datastore}
}

func (m MySQL) write(ctx context.Context, table string, key string, obj AntiObj) error {

	jsonAntiObj, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", table)
	stmt, err := m.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(key, jsonAntiObj)

	return err
}

// If there is more than one value for the same key this function only returns one
func (m MySQL) read(ctx context.Context, table string, key string) (AntiObj, error) {

	var value []byte
	query := fmt.Sprintf("SELECT value FROM %s WHERE k = ?", table)
	err := m.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return AntiObj{}, ErrNotFound
	} else if err != nil {
		return AntiObj{}, err
	}

	var antiObj AntiObj
	err = json.Unmarshal(value, &antiObj)
	if err != nil {
		return AntiObj{}, err
	}

	return antiObj, err
}

func (m MySQL) consume(context.Context, string, string, chan struct{}) (<-chan AntiObj, error) {
	return nil, nil
}

func (m MySQL) barrier(ctx context.Context, lineage []WriteIdentifier, datastoreID string) error {

	for _, writeIdentifier := range lineage {
		fmt.Println("key after for: ", writeIdentifier.Key)
		if writeIdentifier.Dtstid == datastoreID {
			for {
				// Query the database for the value associated with the writeIdentifier.Key
				query := fmt.Sprintf("SELECT value FROM %s WHERE k = ?", writeIdentifier.TableId)
				rows, err := m.db.Query(query, writeIdentifier.Key)

				if !errors.Is(err, sql.ErrNoRows) && err != nil {
					return err
				} else if errors.Is(err, sql.ErrNoRows) { //the version replication process is not yet completed
					fmt.Println("replication in progress")
					continue
				} else {
					defer rows.Close()
					replicationDone := false
					for rows.Next() {
						var value []byte
						err := rows.Scan(&value)
						if err != nil {
							return err
						}
						var antiObj AntiObj
						err = json.Unmarshal(value, &antiObj)
						if err != nil {
							return err
						}

						if antiObj.Version == writeIdentifier.Version { //the version replication process is already completed
							fmt.Println("replication done: ", antiObj.Version)
							replicationDone = true
							break
						}
					}
					//checking there were no errors during iteration
					if err := rows.Err(); err != nil && !replicationDone {
						return err
					}
					if replicationDone { //the version replication process is already completed
						break
					} else { //the version replication process is not yet completed
						fmt.Println("replication of the new version in progress!")
						continue
					}
				}
			}
		}
	}
	return nil
}
