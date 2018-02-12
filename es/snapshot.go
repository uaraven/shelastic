package es

import (
	"fmt"
	"shelastic/utils"
)

type SnapshotInfo map[string]interface{}
type RepositoryList map[string]interface{}

func (e Es) ListRepository() (*RepositoryList, error) {
	data, err := e.getJSON("/_snapshot/")
	if err == nil {
		result := &RepositoryList{}
		utils.DictToAny(data, result)
		return result, nil
	}
	return nil, err
}

func (e Es) RegisterRepository(repo string, sType string, settings map[string]string) error {
	url := fmt.Sprintf("/_snapshot/%s", repo)
	config := make(map[string]interface{})
	config["type"] = sType
	config["settings"] = settings

	data, err := utils.MapToJSON(config)

	if err != nil {
		return err
	}

	_, err = e.putJson(url, data)
	return err
}

func (e Es) VerifyRepository(repo string) error {
	url := fmt.Sprintf("/_snapshot/%s/_verify", repo)
	data, err := e.postJSON(url, "")
	if err != nil {
		return err
	}
	return checkError(data)
}

func (e Es) CreateSnapshot(repo string, snapshotName string) error {
	url := fmt.Sprintf("/_snapshot/%s/%s?wait_for_completion=true", repo, snapshotName)
	data, err := e.putJson(url, "")
	if err != nil {
		return err
	}
	return checkError(data)
}

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

func (e Es) DeleteSnapshot(repo string, snapshotName string) error {
	url := fmt.Sprintf("/_snapshot/%s/%s", repo, snapshotName)
	err := e.delete(url)
	return err
}

func (e Es) RestoreSnapshot(repo string, snapshotName string) error {
	url := fmt.Sprintf("/_snapshot/%s/%s/_restore", repo, snapshotName)
	resp, err := e.postJSON(url, "")
	if err != nil {
		return err
	}
	return checkError(resp)
}
