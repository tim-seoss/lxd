// +build linux,cgo,!agent

package db

// The code below was generated by lxd-generate - DO NOT EDIT!

import (
	"database/sql"
	"fmt"
	"github.com/lxc/lxd/lxd/db/cluster"
	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/shared/api"
	"github.com/pkg/errors"
)

var _ = api.ServerEnvironment{}

var profileNames = cluster.RegisterStmt(`
SELECT projects.name AS project, profiles.name
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  ORDER BY projects.id, profiles.name
`)

var profileNamesByProject = cluster.RegisterStmt(`
SELECT projects.name AS project, profiles.name
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  WHERE project = ? ORDER BY projects.id, profiles.name
`)

var profileNamesByProjectAndName = cluster.RegisterStmt(`
SELECT projects.name AS project, profiles.name
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  WHERE project = ? AND profiles.name = ? ORDER BY projects.id, profiles.name
`)

var profileObjects = cluster.RegisterStmt(`
SELECT profiles.id, projects.name AS project, profiles.name, coalesce(profiles.description, '')
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  ORDER BY projects.id, profiles.name
`)

var profileObjectsByProject = cluster.RegisterStmt(`
SELECT profiles.id, projects.name AS project, profiles.name, coalesce(profiles.description, '')
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  WHERE project = ? ORDER BY projects.id, profiles.name
`)

var profileObjectsByProjectAndName = cluster.RegisterStmt(`
SELECT profiles.id, projects.name AS project, profiles.name, coalesce(profiles.description, '')
  FROM profiles JOIN projects ON profiles.project_id = projects.id
  WHERE project = ? AND profiles.name = ? ORDER BY projects.id, profiles.name
`)

var profileConfigRef = cluster.RegisterStmt(`
SELECT project, name, key, value FROM profiles_config_ref ORDER BY project, name
`)

var profileConfigRefByProject = cluster.RegisterStmt(`
SELECT project, name, key, value FROM profiles_config_ref WHERE project = ? ORDER BY project, name
`)

var profileConfigRefByProjectAndName = cluster.RegisterStmt(`
SELECT project, name, key, value FROM profiles_config_ref WHERE project = ? AND name = ? ORDER BY project, name
`)

var profileDevicesRef = cluster.RegisterStmt(`
SELECT project, name, device, type, key, value FROM profiles_devices_ref ORDER BY project, name
`)

var profileDevicesRefByProject = cluster.RegisterStmt(`
SELECT project, name, device, type, key, value FROM profiles_devices_ref WHERE project = ? ORDER BY project, name
`)

var profileDevicesRefByProjectAndName = cluster.RegisterStmt(`
SELECT project, name, device, type, key, value FROM profiles_devices_ref WHERE project = ? AND name = ? ORDER BY project, name
`)

var profileUsedByRef = cluster.RegisterStmt(`
SELECT project, name, value FROM profiles_used_by_ref ORDER BY project, name
`)

var profileUsedByRefByProject = cluster.RegisterStmt(`
SELECT project, name, value FROM profiles_used_by_ref WHERE project = ? ORDER BY project, name
`)

var profileUsedByRefByProjectAndName = cluster.RegisterStmt(`
SELECT project, name, value FROM profiles_used_by_ref WHERE project = ? AND name = ? ORDER BY project, name
`)

var profileID = cluster.RegisterStmt(`
SELECT profiles.id FROM profiles JOIN projects ON profiles.project_id = projects.id
  WHERE projects.name = ? AND profiles.name = ?
`)

var profileCreate = cluster.RegisterStmt(`
INSERT INTO profiles (project_id, name, description)
  VALUES ((SELECT projects.id FROM projects WHERE projects.name = ?), ?, ?)
`)

var profileCreateConfigRef = cluster.RegisterStmt(`
INSERT INTO profiles_config (profile_id, key, value)
  VALUES (?, ?, ?)
`)

var profileCreateDevicesRef = cluster.RegisterStmt(`
INSERT INTO profiles_devices (profile_id, name, type)
  VALUES (?, ?, ?)
`)
var profileCreateDevicesConfigRef = cluster.RegisterStmt(`
INSERT INTO profiles_devices_config (profile_device_id, key, value)
  VALUES (?, ?, ?)
`)

var profileRename = cluster.RegisterStmt(`
UPDATE profiles SET name = ? WHERE project_id = (SELECT projects.id FROM projects WHERE projects.name = ?) AND name = ?
`)

var profileDelete = cluster.RegisterStmt(`
DELETE FROM profiles WHERE project_id = (SELECT projects.id FROM projects WHERE projects.name = ?) AND name = ?
`)

var profileDeleteConfigRef = cluster.RegisterStmt(`
DELETE FROM profiles_config WHERE profile_id = ?
`)

var profileDeleteDevicesRef = cluster.RegisterStmt(`
DELETE FROM profiles_devices WHERE profile_id = ?
`)

var profileUpdate = cluster.RegisterStmt(`
UPDATE profiles
  SET project_id = (SELECT id FROM projects WHERE name = ?), name = ?, description = ?
 WHERE id = ?
`)

// GetProfileURIs returns all available profile URIs.
func (c *ClusterTx) GetProfileURIs(filter ProfileFilter) ([]string, error) {
	// Check which filter criteria are active.
	criteria := map[string]interface{}{}
	if filter.Project != "" {
		criteria["Project"] = filter.Project
	}
	if filter.Name != "" {
		criteria["Name"] = filter.Name
	}

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if criteria["Project"] != nil && criteria["Name"] != nil {
		stmt = c.stmt(profileNamesByProjectAndName)
		args = []interface{}{
			filter.Project,
			filter.Name,
		}
	} else if criteria["Project"] != nil {
		stmt = c.stmt(profileNamesByProject)
		args = []interface{}{
			filter.Project,
		}
	} else {
		stmt = c.stmt(profileNames)
		args = []interface{}{}
	}

	code := cluster.EntityTypes["profile"]
	formatter := cluster.EntityFormatURIs[code]

	return query.SelectURIs(stmt, formatter, args...)
}

// GetProfiles returns all available profiles.
func (c *ClusterTx) GetProfiles(filter ProfileFilter) ([]Profile, error) {
	// Result slice.
	objects := make([]Profile, 0)

	// Check which filter criteria are active.
	criteria := map[string]interface{}{}
	if filter.Project != "" {
		criteria["Project"] = filter.Project
	}
	if filter.Name != "" {
		criteria["Name"] = filter.Name
	}

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if criteria["Project"] != nil && criteria["Name"] != nil {
		stmt = c.stmt(profileObjectsByProjectAndName)
		args = []interface{}{
			filter.Project,
			filter.Name,
		}
	} else if criteria["Project"] != nil {
		stmt = c.stmt(profileObjectsByProject)
		args = []interface{}{
			filter.Project,
		}
	} else {
		stmt = c.stmt(profileObjects)
		args = []interface{}{}
	}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, Profile{})
		return []interface{}{
			&objects[i].ID,
			&objects[i].Project,
			&objects[i].Name,
			&objects[i].Description,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch profiles")
	}

	// Fill field Config.
	configObjects, err := c.ProfileConfigRef(filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch field Config")
	}

	for i := range objects {
		_, ok0 := configObjects[objects[i].Project]
		if !ok0 {
			subIndex := map[string]map[string]string{}
			configObjects[objects[i].Project] = subIndex
		}

		value := configObjects[objects[i].Project][objects[i].Name]
		if value == nil {
			value = map[string]string{}
		}
		objects[i].Config = value
	}

	// Fill field Devices.
	devicesObjects, err := c.ProfileDevicesRef(filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch field Devices")
	}

	for i := range objects {
		_, ok0 := devicesObjects[objects[i].Project]
		if !ok0 {
			subIndex := map[string]map[string]map[string]string{}
			devicesObjects[objects[i].Project] = subIndex
		}

		value := devicesObjects[objects[i].Project][objects[i].Name]
		if value == nil {
			value = map[string]map[string]string{}
		}
		objects[i].Devices = value
	}

	// Fill field UsedBy.
	usedByObjects, err := c.ProfileUsedByRef(filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch field UsedBy")
	}

	for i := range objects {
		_, ok0 := usedByObjects[objects[i].Project]
		if !ok0 {
			subIndex := map[string][]string{}
			usedByObjects[objects[i].Project] = subIndex
		}

		value := usedByObjects[objects[i].Project][objects[i].Name]
		if value == nil {
			value = []string{}
		}
		objects[i].UsedBy = value
	}

	return objects, nil
}

// GetProfile returns the profile with the given key.
func (c *ClusterTx) GetProfile(project string, name string) (*Profile, error) {
	filter := ProfileFilter{}
	filter.Project = project
	filter.Name = name

	objects, err := c.GetProfiles(filter)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch Profile")
	}

	switch len(objects) {
	case 0:
		return nil, ErrNoSuchObject
	case 1:
		return &objects[0], nil
	default:
		return nil, fmt.Errorf("More than one profile matches")
	}
}

// ProfileExists checks if a profile with the given key exists.
func (c *ClusterTx) ProfileExists(project string, name string) (bool, error) {
	_, err := c.GetProfileID(project, name)
	if err != nil {
		if err == ErrNoSuchObject {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetProfileID return the ID of the profile with the given key.
func (c *ClusterTx) GetProfileID(project string, name string) (int64, error) {
	stmt := c.stmt(profileID)
	rows, err := stmt.Query(project, name)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get profile ID")
	}
	defer rows.Close()

	// For sanity, make sure we read one and only one row.
	if !rows.Next() {
		return -1, ErrNoSuchObject
	}
	var id int64
	err = rows.Scan(&id)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to scan ID")
	}
	if rows.Next() {
		return -1, fmt.Errorf("More than one row returned")
	}
	err = rows.Err()
	if err != nil {
		return -1, errors.Wrap(err, "Result set failure")
	}

	return id, nil
}

// ProfileConfigRef returns entities used by profiles.
func (c *ClusterTx) ProfileConfigRef(filter ProfileFilter) (map[string]map[string]map[string]string, error) {
	// Result slice.
	objects := make([]struct {
		Project string
		Name    string
		Key     string
		Value   string
	}, 0)

	// Check which filter criteria are active.
	criteria := map[string]interface{}{}
	if filter.Project != "" {
		criteria["Project"] = filter.Project
	}
	if filter.Name != "" {
		criteria["Name"] = filter.Name
	}

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if criteria["Project"] != nil && criteria["Name"] != nil {
		stmt = c.stmt(profileConfigRefByProjectAndName)
		args = []interface{}{
			filter.Project,
			filter.Name,
		}
	} else if criteria["Project"] != nil {
		stmt = c.stmt(profileConfigRefByProject)
		args = []interface{}{
			filter.Project,
		}
	} else {
		stmt = c.stmt(profileConfigRef)
		args = []interface{}{}
	}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, struct {
			Project string
			Name    string
			Key     string
			Value   string
		}{})
		return []interface{}{
			&objects[i].Project,
			&objects[i].Name,
			&objects[i].Key,
			&objects[i].Value,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch  ref for profiles")
	}

	// Build index by primary name.
	index := map[string]map[string]map[string]string{}

	for _, object := range objects {
		_, ok0 := index[object.Project]
		if !ok0 {
			subIndex := map[string]map[string]string{}
			index[object.Project] = subIndex
		}

		item, ok := index[object.Project][object.Name]
		if !ok {
			item = map[string]string{}
		}

		index[object.Project][object.Name] = item
		item[object.Key] = object.Value
	}

	return index, nil
}

// ProfileDevicesRef returns entities used by profiles.
func (c *ClusterTx) ProfileDevicesRef(filter ProfileFilter) (map[string]map[string]map[string]map[string]string, error) {
	// Result slice.
	objects := make([]struct {
		Project string
		Name    string
		Device  string
		Type    int
		Key     string
		Value   string
	}, 0)

	// Check which filter criteria are active.
	criteria := map[string]interface{}{}
	if filter.Project != "" {
		criteria["Project"] = filter.Project
	}
	if filter.Name != "" {
		criteria["Name"] = filter.Name
	}

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if criteria["Project"] != nil && criteria["Name"] != nil {
		stmt = c.stmt(profileDevicesRefByProjectAndName)
		args = []interface{}{
			filter.Project,
			filter.Name,
		}
	} else if criteria["Project"] != nil {
		stmt = c.stmt(profileDevicesRefByProject)
		args = []interface{}{
			filter.Project,
		}
	} else {
		stmt = c.stmt(profileDevicesRef)
		args = []interface{}{}
	}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, struct {
			Project string
			Name    string
			Device  string
			Type    int
			Key     string
			Value   string
		}{})
		return []interface{}{
			&objects[i].Project,
			&objects[i].Name,
			&objects[i].Device,
			&objects[i].Type,
			&objects[i].Key,
			&objects[i].Value,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch  ref for profiles")
	}

	// Build index by primary name.
	index := map[string]map[string]map[string]map[string]string{}

	for _, object := range objects {
		_, ok0 := index[object.Project]
		if !ok0 {
			subIndex := map[string]map[string]map[string]string{}
			index[object.Project] = subIndex
		}

		item, ok := index[object.Project][object.Name]
		if !ok {
			item = map[string]map[string]string{}
		}

		index[object.Project][object.Name] = item
		config, ok := item[object.Device]
		if !ok {
			// First time we see this device, let's int the config
			// and add the type.
			deviceType, err := deviceTypeToString(object.Type)
			if err != nil {
				return nil, errors.Wrapf(
					err, "unexpected device type code '%d'", object.Type)
			}
			config = map[string]string{}
			config["type"] = deviceType
			item[object.Device] = config
		}
		if object.Key != "" {
			config[object.Key] = object.Value
		}
	}

	return index, nil
}

// ProfileUsedByRef returns entities used by profiles.
func (c *ClusterTx) ProfileUsedByRef(filter ProfileFilter) (map[string]map[string][]string, error) {
	// Result slice.
	objects := make([]struct {
		Project string
		Name    string
		Value   string
	}, 0)

	// Check which filter criteria are active.
	criteria := map[string]interface{}{}
	if filter.Project != "" {
		criteria["Project"] = filter.Project
	}
	if filter.Name != "" {
		criteria["Name"] = filter.Name
	}

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if criteria["Project"] != nil && criteria["Name"] != nil {
		stmt = c.stmt(profileUsedByRefByProjectAndName)
		args = []interface{}{
			filter.Project,
			filter.Name,
		}
	} else if criteria["Project"] != nil {
		stmt = c.stmt(profileUsedByRefByProject)
		args = []interface{}{
			filter.Project,
		}
	} else {
		stmt = c.stmt(profileUsedByRef)
		args = []interface{}{}
	}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, struct {
			Project string
			Name    string
			Value   string
		}{})
		return []interface{}{
			&objects[i].Project,
			&objects[i].Name,
			&objects[i].Value,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch string ref for profiles")
	}

	// Build index by primary name.
	index := map[string]map[string][]string{}

	for _, object := range objects {
		_, ok0 := index[object.Project]
		if !ok0 {
			subIndex := map[string][]string{}
			index[object.Project] = subIndex
		}

		item, ok := index[object.Project][object.Name]
		if !ok {
			item = []string{}
		}

		index[object.Project][object.Name] = append(item, object.Value)
	}

	return index, nil
}

// CreateProfile adds a new profile to the database.
func (c *ClusterTx) CreateProfile(object Profile) (int64, error) {
	// Check if a profile with the same key exists.
	exists, err := c.ProfileExists(object.Project, object.Name)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to check for duplicates")
	}
	if exists {
		return -1, fmt.Errorf("This profile already exists")
	}

	args := make([]interface{}, 3)

	// Populate the statement arguments.
	args[0] = object.Project
	args[1] = object.Name
	args[2] = object.Description

	// Prepared statement to use.
	stmt := c.stmt(profileCreate)

	// Execute the statement.
	result, err := stmt.Exec(args...)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to create profile")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "Failed to fetch profile ID")
	}

	// Insert config reference.
	stmt = c.stmt(profileCreateConfigRef)
	for key, value := range object.Config {
		_, err := stmt.Exec(id, key, value)
		if err != nil {
			return -1, errors.Wrap(err, "Insert config for profile")
		}
	}

	// Insert devices reference.
	for name, config := range object.Devices {
		typ, ok := config["type"]
		if !ok {
			return -1, fmt.Errorf("No type for device %s", name)
		}
		typCode, err := deviceTypeToInt(typ)
		if err != nil {
			return -1, errors.Wrapf(err, "Device type code for %s", typ)
		}
		stmt = c.stmt(profileCreateDevicesRef)
		result, err := stmt.Exec(id, name, typCode)
		if err != nil {
			return -1, errors.Wrapf(err, "Insert device %s", name)
		}
		deviceID, err := result.LastInsertId()
		if err != nil {
			return -1, errors.Wrap(err, "Failed to fetch device ID")
		}
		stmt = c.stmt(profileCreateDevicesConfigRef)
		for key, value := range config {
			_, err := stmt.Exec(deviceID, key, value)
			if err != nil {
				return -1, errors.Wrap(err, "Insert config for profile")
			}
		}
	}

	return id, nil
}

// RenameProfile renames the profile matching the given key parameters.
func (c *ClusterTx) RenameProfile(project string, name string, to string) error {
	stmt := c.stmt(profileRename)
	result, err := stmt.Exec(to, project, name)
	if err != nil {
		return errors.Wrap(err, "Rename profile")
	}

	n, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Fetch affected rows")
	}
	if n != 1 {
		return fmt.Errorf("Query affected %d rows instead of 1", n)
	}
	return nil
}

// DeleteProfile deletes the profile matching the given key parameters.
func (c *ClusterTx) DeleteProfile(project string, name string) error {
	stmt := c.stmt(profileDelete)
	result, err := stmt.Exec(project, name)
	if err != nil {
		return errors.Wrap(err, "Delete profile")
	}

	n, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Fetch affected rows")
	}
	if n != 1 {
		return fmt.Errorf("Query deleted %d rows instead of 1", n)
	}

	return nil
}

// UpdateProfile updates the profile matching the given key parameters.
func (c *ClusterTx) UpdateProfile(project string, name string, object Profile) error {
	id, err := c.GetProfileID(project, name)
	if err != nil {
		return errors.Wrap(err, "Get profile")
	}

	stmt := c.stmt(profileUpdate)
	result, err := stmt.Exec(object.Project, object.Name, object.Description, id)
	if err != nil {
		return errors.Wrap(err, "Update profile")
	}

	n, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Fetch affected rows")
	}
	if n != 1 {
		return fmt.Errorf("Query updated %d rows instead of 1", n)
	}

	// Delete current config.
	stmt = c.stmt(profileDeleteConfigRef)
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "Delete current config")
	}

	// Insert config reference.
	stmt = c.stmt(profileCreateConfigRef)
	for key, value := range object.Config {
		if value == "" {
			continue
		}
		_, err := stmt.Exec(id, key, value)
		if err != nil {
			return errors.Wrap(err, "Insert config for profile")
		}
	}

	// Delete current devices.
	stmt = c.stmt(profileDeleteDevicesRef)
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "Delete current devices")
	}

	// Insert devices reference.
	for name, config := range object.Devices {
		typ, ok := config["type"]
		if !ok {
			return fmt.Errorf("No type for device %s", name)
		}
		typCode, err := deviceTypeToInt(typ)
		if err != nil {
			return errors.Wrapf(err, "Device type code for %s", typ)
		}
		stmt = c.stmt(profileCreateDevicesRef)
		result, err := stmt.Exec(id, name, typCode)
		if err != nil {
			return errors.Wrapf(err, "Insert device %s", name)
		}
		deviceID, err := result.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "Failed to fetch device ID")
		}
		stmt = c.stmt(profileCreateDevicesConfigRef)
		for key, value := range config {
			if value == "" {
				continue
			}
			_, err := stmt.Exec(deviceID, key, value)
			if err != nil {
				return errors.Wrap(err, "Insert config for profile")
			}
		}
	}

	return nil
}
