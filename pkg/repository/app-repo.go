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
		entityID, "Strategy", app.Strategy)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "gitUrl", app.Repository)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "routeSite", app.Route.Site)

	_, err = repo.DB.Exec(
		sqlStatement,
		entityID, "routePath", app.Route.Path)

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

	return err

}
