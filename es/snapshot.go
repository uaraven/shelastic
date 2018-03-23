package es

import (
	"fmt"
	"shelastic/utils"
)

// SnapshotInfo is a JSON wrapper for snapshot information
type SnapshotInfo map[string]interface{}

// RepositoryList is a JSON wrapper for list of repositories information
type RepositoryList map[string]interface{}

// ListRepository lists all available repositories
func (e Es) ListRepository() (*RepositoryList, error) {
	data, err := e.getJSON("/_snapshot/")
	if err == nil {
		result := &RepositoryList{}
		utils.DictToAny(data, result)
		return result, nil
	}
	return nil, err
}

// RegisterRepository registers a new repository
func (e Es) RegisterRepository(repo string, sType string, settings map[string]string) error {
	url := fmt.Sprintf("/_snapshot/%s", repo)
	config := make(map[string]interface{})
	config["type"] = sType
	config["settings"] = settings

	data, err := utils.MapToJSON(config)

	if err != nil {
		return err
	}

	resp, err := e.putJSON(url, data)
	if err != nil {
		return err
	}
	err = checkError(resp)
	return err
}

// VerifyRepository verifies repository configuration
func (e Es) VerifyRepository(repo string) error {
	url := fmt.Sprintf("/_snapshot/%s/_verify", repo)
	data, err := e.postJSON(url, "")
	if err != nil {
		return err
	}
	return checkError(data)
}

// CreateSnapshot creates a snapshot with a given name in a repository
func (e Es) CreateSnapshot(repo string, snapshotName string) error {
	url := fmt.Sprintf("/_snapshot/%s/%s?wait_for_completion=true", repo, snapshotName)
	data, err := e.putJSON(url, "")
	if err != nil {
		return err
	}
	return checkError(data)
}

// GetSnapshotInfo retrieves information of a snapshot with a given name in a repository
func (e Es) GetSnapshotInfo(repo string, snapshotName string) (*SnapshotInfo, error) {
	url := fmt.Sprintf("/_snapshot/%s/%s", repo, snapshotName)
	data, err := e.getJSON(url)
	if err != nil {
		return nil, err
	}
	err = checkError(data)
	if err != nil {
		return nil, err
	}
	result := &SnapshotInfo{}
	err = utils.DictToAny(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteSnapshot deletes a snapshot with a given name from a repository
func (e Es) DeleteSnapshot(repo string, snapshotName string) error {
	url := fmt.Sprintf("/_snapshot/%s/%s", repo, snapshotName)
	_, err := e.delete(url)
	return err
}

// RestoreSnapshot restores a snapshot with a given name from a repository
func (e Es) RestoreSnapshot(repo string, snapshotName string) error {
	indices, err := e.ListIndices()
	if err != nil {
		return err
	}

	for _, index := range indices {
		e.CloseIndex(index.Name)
	}

	url := fmt.Sprintf("/_snapshot/%s/%s/_restore", repo, snapshotName)
	resp, err := e.postJSON(url, "")
	if err != nil {
		return err
	}

	return checkError(resp)
}
