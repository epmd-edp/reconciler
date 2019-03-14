package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"log"
)

type AppRepo struct {
	DB sql.DB
}

func (repo AppRepo) AddApplication(app model.AppEntity) error {

	var entityID int
	repo.DB.QueryRow(
		"insert into business_entity (name,tenant,be_type) values ($1,$2,$3) returning id",
		app.Name, app.Namespace, "app").Scan(&entityID)

	log.Println("Getting ID of the created Business Entity ", entityID)

	sqlStatement := "insert into be_properties (be_id, property,value) values ($1,$2,$3)"
	_, err := repo.DB.Exec(
		sqlStatement,
		entityID, "lang", app.Lang)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "framework", app.Framework)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "buildTool", app.BuildTool)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "strategy", app.Strategy)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "gitUrl", app.Repository)

	if app.Route != nil {

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "routeSite", app.Route.Site)

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "routePath", app.Route.Path)
	}

	if app.Database != nil {

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "databaseKind", app.Database.Kind)

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "databaseVersion", app.Database.Version)

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "databaseCapacity", app.Database.Capacity)

		_, err = repo.DB.Exec(
			sqlStatement,
			entityID, "databaseStorage", app.Database.Storage)
	}

	if app.Status != nil {

		var statusID int
		repo.DB.QueryRow(
			"select status_id from statuses_list where status_name = $1", app.Status.Status).Scan(&statusID)
		log.Println("Getting status ID for created Business Entity", statusID)

		_, err = repo.DB.Exec(
			"insert into be_status (be_id, status, last_time_update,available) values($1,$2,$3,$4,$5,$6)",
			entityID, statusID, app.Status.LastTimeUpdated, app.Status.Available)
	}

	return err
}
