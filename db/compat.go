package db

import "github.com/cloud-barista/cm-cicada/pkg/api/rest/model"

// TaskComponentGetByName is a transitional shim kept to avoid breaking callers
// in lib/airflow while the gusty.go refactor (removal of db direct access) is
// in flight. Delete alongside the gusty.go call-site migration.
//
// Deprecated: use dao.TaskComponentGetByName.
func TaskComponentGetByName(name string) *model.TaskComponent {
	if DB == nil {
		return nil
	}
	tc := &model.TaskComponent{}
	if err := DB.Where("name = ?", name).First(tc).Error; err != nil {
		return nil
	}
	return tc
}
